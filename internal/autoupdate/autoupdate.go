// File : internal/autoupdate/autoupdate.go
// Deskripsi : Fitur auto-update binary sfDBTools via GitHub Releases
// Author : Hadiyatna Muflihun
// Tanggal : 5 Januari 2026
// Last Modified : 5 Januari 2026
package autoupdate

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"

	"sfDBTools/pkg/consts"
	"sfDBTools/pkg/version"
)

type Logger interface {
	Infof(format string, args ...any)
	Warnf(format string, args ...any)
	Errorf(format string, args ...any)
}

type Options struct {
	// RepoOwner dan RepoName defaultnya mengarah ke repo upstream.
	RepoOwner string
	RepoName  string
	// Timeout untuk request HTTP.
	Timeout time.Duration
	// Force akan menjalankan update walau versi sama (biasanya false).
	Force bool
	// ReExec akan menjalankan ulang proses setelah update (disarankan untuk auto-update).
	ReExec bool
}

type ReleaseAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
	Size               int64  `json:"size"`
}

type Release struct {
	TagName string         `json:"tag_name"`
	Assets  []ReleaseAsset `json:"assets"`
}

func DefaultOptions() Options {
	return Options{
		RepoOwner: "hadiy961",
		RepoName:  "sfDBTools-New",
		Timeout:   25 * time.Second,
		Force:     false,
		ReExec:    true,
	}
}

func OptionsFromEnv() Options {
	opts := DefaultOptions()
	if v := strings.TrimSpace(os.Getenv(consts.ENV_UPDATE_REPO_OWNER)); v != "" {
		opts.RepoOwner = v
	}
	if v := strings.TrimSpace(os.Getenv(consts.ENV_UPDATE_REPO_NAME)); v != "" {
		opts.RepoName = v
	}
	return opts
}

// MaybeAutoUpdate mengecek update dan melakukan self-update jika ditemukan versi lebih baru.
// Jika update sukses, proses akan re-exec binary yang baru.
func MaybeAutoUpdate(ctx context.Context, log Logger) error {
	if !AutoUpdateEnabled() {
		return nil
	}

	opts := OptionsFromEnv()
	return UpdateIfNeeded(ctx, log, opts)
}

// AutoUpdateEnabled menentukan apakah auto-update aktif.
// Default: aktif (tanpa perlu set env), kecuali SFDB_NO_AUTO_UPDATE=1.
func AutoUpdateEnabled() bool {
	if strings.TrimSpace(os.Getenv(consts.ENV_NO_AUTO_UPDATE)) == "1" {
		return false
	}
	// Backward-compatible: jika user eksplisit set SFDB_AUTO_UPDATE=0/empty,
	// tetap anggap aktif kecuali disable flag digunakan.
	// Jika user set SFDB_AUTO_UPDATE=1 juga tetap aktif.
	return true
}

// UpdateIfNeeded melakukan pengecekan versi lalu update jika perlu.
// Jika update terjadi, fungsi akan mencoba re-exec binary baru.
func UpdateIfNeeded(ctx context.Context, log Logger, opts Options) error {
	if runtime.GOOS != "linux" || runtime.GOARCH != "amd64" {
		// Saat ini workflow release hanya membangun linux/amd64.
		if log != nil {
			log.Warnf("Auto-update dilewati (unsupported platform): %s/%s", runtime.GOOS, runtime.GOARCH)
		}
		return nil
	}

	current, ok := parseSemver(version.Version)
	if !ok {
		// Jika versi "dev" atau tidak valid, jangan auto-update.
		if log != nil {
			log.Warnf("Auto-update dilewati (versi lokal tidak semver): %s", version.Version)
		}
		return nil
	}

	// Cek koneksi internet dulu. Kalau offline, skip supaya startup tidak delay.
	checkTimeout := 2 * time.Second
	if opts.Timeout > 0 && opts.Timeout < checkTimeout {
		checkTimeout = opts.Timeout
	}
	online := isInternetAvailable(ctx, checkTimeout)
	if !online {
		if log != nil {
			log.Infof("Auto-update dilewati (tidak ada koneksi internet)")
		}
		return nil
	}

	rel, err := fetchLatestRelease(ctx, opts)
	if err != nil {
		return err
	}
	latest, ok := parseSemver(rel.TagName)
	if !ok {
		return fmt.Errorf("tag release tidak valid: %q", rel.TagName)
	}

	if !opts.Force {
		cmp := compareSemver(latest, current)
		if cmp <= 0 {
			if log != nil {
				log.Infof("Sudah versi terbaru: %s", version.Version)
			}
			return nil
		}
	}

	assetTar, err := findAsset(rel.Assets, "sfDBTools_linux_amd64.tar.gz")
	if err != nil {
		return err
	}

	// File sha256 versi spesifik bersifat optional.
	shaName := fmt.Sprintf("sfDBTools_%d.%d.%d_linux_amd64.sha256", latest.Major, latest.Minor, latest.Patch)
	assetSHA, _ := findAsset(rel.Assets, shaName)

	if log != nil {
		log.Infof("Update tersedia: %s -> %s. Mengunduh...", version.Version, rel.TagName)
	}

	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("gagal mendapatkan path executable: %w", err)
	}
	// Jika executable adalah symlink (mis. /usr/bin/sfdbtools), update target-nya.
	targetPath := execPath
	if resolved, rerr := filepath.EvalSymlinks(execPath); rerr == nil && resolved != "" {
		targetPath = resolved
	}

	tmpDir, err := os.MkdirTemp("", "sfdbtools-update-*")
	if err != nil {
		return fmt.Errorf("gagal membuat temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	tarPath := filepath.Join(tmpDir, assetTar.Name)
	if err := downloadToFile(ctx, opts, assetTar.BrowserDownloadURL, tarPath); err != nil {
		return err
	}

	var expectedSHA string
	if assetSHA != nil {
		shaPath := filepath.Join(tmpDir, assetSHA.Name)
		if err := downloadToFile(ctx, opts, assetSHA.BrowserDownloadURL, shaPath); err != nil {
			return err
		}
		expected, err := parseSHA256File(shaPath)
		if err != nil {
			return err
		}
		expectedSHA = expected
	}
	if expectedSHA != "" {
		actual, err := sha256FileHex(tarPath)
		if err != nil {
			return err
		}
		if !strings.EqualFold(actual, expectedSHA) {
			return fmt.Errorf("sha256 mismatch: expected=%s actual=%s", expectedSHA, actual)
		}
	}

	newBinPath := filepath.Join(tmpDir, "sfDBTools")
	if err := extractSingleFileFromTarGz(tarPath, "sfDBTools", newBinPath); err != nil {
		return err
	}
	if err := os.Chmod(newBinPath, 0o755); err != nil {
		return fmt.Errorf("gagal chmod binary baru: %w", err)
	}

	if err := replaceFileAtomic(targetPath, newBinPath); err != nil {
		// Biasanya terjadi jika install di /usr/bin tanpa sudo.
		return err
	}

	if log != nil {
		log.Infof("Update selesai. Menjalankan ulang binary terbaru...")
	}
	if opts.ReExec {
		argv := append([]string{targetPath}, os.Args[1:]...)
		if err := syscall.Exec(targetPath, argv, os.Environ()); err != nil {
			// Jika exec gagal, minimal binary sudah ter-update.
			if log != nil {
				log.Warnf("Update berhasil tapi gagal restart otomatis: %v. Silakan jalankan ulang perintah.", err)
			}
			return nil
		}
		return nil
	}

	if log != nil {
		log.Infof("Update selesai. Silakan jalankan ulang perintah.")
	}

	return nil
}

func isInternetAvailable(ctx context.Context, timeout time.Duration) bool {
	// Strategi sederhana dan cepat:
	// - dial ke endpoint publik (443 / 53) untuk mendeteksi akses internet
	// - jika gagal, coba dial ke api.github.com:443 (target utama update)
	// Tidak memakai ping/ICMP karena sering diblok.
	addrs := []string{
		"1.1.1.1:443",
		"8.8.8.8:53",
		"api.github.com:443",
	}

	d := net.Dialer{Timeout: timeout}
	for _, addr := range addrs {
		cctx, cancel := context.WithTimeout(ctx, timeout)
		conn, err := d.DialContext(cctx, "tcp", addr)
		cancel()
		if err == nil {
			_ = conn.Close()
			return true
		}
	}
	return false
}

func fetchLatestRelease(ctx context.Context, opts Options) (Release, error) {
	var rel Release
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", opts.RepoOwner, opts.RepoName)

	client := &http.Client{Timeout: opts.Timeout}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return rel, fmt.Errorf("gagal membuat request: %w", err)
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "sfDBTools-autoupdate")
	if token := strings.TrimSpace(os.Getenv(consts.ENV_GITHUB_TOKEN)); token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := client.Do(req)
	if err != nil {
		return rel, fmt.Errorf("gagal request GitHub API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return rel, fmt.Errorf("GitHub API error: status=%s body=%s", resp.Status, strings.TrimSpace(string(b)))
	}

	if err := json.NewDecoder(resp.Body).Decode(&rel); err != nil {
		return rel, fmt.Errorf("gagal decode response GitHub: %w", err)
	}
	return rel, nil
}

func findAsset(assets []ReleaseAsset, name string) (*ReleaseAsset, error) {
	for i := range assets {
		if assets[i].Name == name {
			return &assets[i], nil
		}
	}
	return nil, fmt.Errorf("asset tidak ditemukan di release: %s", name)
}

func downloadToFile(ctx context.Context, opts Options, url, dst string) error {
	client := &http.Client{Timeout: opts.Timeout}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("gagal membuat request download: %w", err)
	}
	req.Header.Set("User-Agent", "sfDBTools-autoupdate")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("gagal download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return fmt.Errorf("download error: status=%s body=%s", resp.Status, strings.TrimSpace(string(b)))
	}

	f, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return fmt.Errorf("gagal membuat file download: %w", err)
	}
	defer f.Close()

	if _, err := io.Copy(f, resp.Body); err != nil {
		return fmt.Errorf("gagal menulis file download: %w", err)
	}
	return nil
}

func extractSingleFileFromTarGz(tarGzPath, fileName, dst string) error {
	f, err := os.Open(tarGzPath)
	if err != nil {
		return fmt.Errorf("gagal membuka tar.gz: %w", err)
	}
	defer f.Close()

	gz, err := gzip.NewReader(f)
	if err != nil {
		return fmt.Errorf("gagal membaca gzip: %w", err)
	}
	defer gz.Close()

	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return fmt.Errorf("gagal membaca tar: %w", err)
		}
		if hdr.Typeflag != tar.TypeReg {
			continue
		}
		name := strings.TrimPrefix(hdr.Name, "./")
		if name != fileName {
			continue
		}

		out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o755)
		if err != nil {
			return fmt.Errorf("gagal membuat file extract: %w", err)
		}
		_, copyErr := io.Copy(out, tr)
		closeErr := out.Close()
		if copyErr != nil {
			return fmt.Errorf("gagal extract file: %w", copyErr)
		}
		if closeErr != nil {
			return fmt.Errorf("gagal close file extract: %w", closeErr)
		}
		return nil
	}
	return fmt.Errorf("file tidak ditemukan dalam tar.gz: %s", fileName)
}

func replaceFileAtomic(dst, src string) error {
	dir := filepath.Dir(dst)
	tmp := filepath.Join(dir, fmt.Sprintf(".%s.new", filepath.Base(dst)))

	// Copy ke lokasi yang sama (agar rename atomic di filesystem yang sama).
	if err := copyFile(src, tmp, 0o755); err != nil {
		return err
	}

	// Coba rename langsung. Jika dst ada, rename biasanya overwrite di unix?
	// Go os.Rename akan replace jika dst bukan directory, tapi permission tetap berlaku.
	if err := os.Rename(tmp, dst); err == nil {
		return nil
	}

	// Fallback: rename dst -> .old lalu tmp -> dst.
	old := filepath.Join(dir, fmt.Sprintf(".%s.old", filepath.Base(dst)))
	_ = os.Remove(old)
	if err := os.Rename(dst, old); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("gagal mengganti binary (butuh permission?): %w", err)
	}
	if err := os.Rename(tmp, dst); err != nil {
		// Coba balikin
		_ = os.Rename(old, dst)
		_ = os.Remove(tmp)
		return fmt.Errorf("gagal menaruh binary baru: %w", err)
	}
	_ = os.Remove(old)
	return nil
}

func copyFile(src, dst string, mode os.FileMode) error {
	in, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("gagal buka src: %w", err)
	}
	defer in.Close()

	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode)
	if err != nil {
		return fmt.Errorf("gagal buka dst: %w", err)
	}
	defer func() {
		_ = out.Close()
	}()

	if _, err := io.Copy(out, in); err != nil {
		return fmt.Errorf("gagal copy file: %w", err)
	}
	if err := out.Close(); err != nil {
		return fmt.Errorf("gagal close dst: %w", err)
	}
	return nil
}

func sha256FileHex(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("gagal membuka file untuk sha256: %w", err)
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", fmt.Errorf("gagal hitung sha256: %w", err)
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func parseSHA256File(path string) (string, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("gagal baca sha256 file: %w", err)
	}
	// Format sha256sum: "<hash>  <filename>"
	fields := strings.Fields(string(b))
	if len(fields) < 1 {
		return "", fmt.Errorf("format sha256 tidak valid")
	}
	return strings.TrimSpace(fields[0]), nil
}

type Semver struct {
	Major int
	Minor int
	Patch int
}

func parseSemver(v string) (Semver, bool) {
	vv := strings.TrimSpace(v)
	vv = strings.TrimPrefix(vv, "v")

	// Ambil prefix angka.semver (abaikan suffix seperti -dirty atau +meta)
	cut := len(vv)
	for i := 0; i < len(vv); i++ {
		c := vv[i]
		if (c >= '0' && c <= '9') || c == '.' {
			continue
		}
		cut = i
		break
	}
	vv = vv[:cut]

	parts := strings.Split(vv, ".")
	if len(parts) != 3 {
		return Semver{}, false
	}
	maj, err1 := strconv.Atoi(parts[0])
	min, err2 := strconv.Atoi(parts[1])
	pat, err3 := strconv.Atoi(parts[2])
	if err1 != nil || err2 != nil || err3 != nil {
		return Semver{}, false
	}
	if maj < 0 || min < 0 || pat < 0 {
		return Semver{}, false
	}
	return Semver{Major: maj, Minor: min, Patch: pat}, true
}

// compareSemver mengembalikan 1 jika a>b, -1 jika a<b, 0 jika sama.
func compareSemver(a, b Semver) int {
	if a.Major != b.Major {
		if a.Major > b.Major {
			return 1
		}
		return -1
	}
	if a.Minor != b.Minor {
		if a.Minor > b.Minor {
			return 1
		}
		return -1
	}
	if a.Patch != b.Patch {
		if a.Patch > b.Patch {
			return 1
		}
		return -1
	}
	return 0
}
