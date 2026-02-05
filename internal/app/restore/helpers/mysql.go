// File : internal/restore/helpers/mysql.go
// Deskripsi : Helper functions untuk MySQL restore operations
// Author : Hadiyatna Muflihun
// Tanggal : 17 Desember 2025
// Last Modified : 26 Januari 2026
package helpers

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	backupfile "sfdbtools/internal/app/backup/helpers/file"
	profileconn "sfdbtools/internal/app/profile/connection"
	"sfdbtools/internal/crypto"
	"sfdbtools/internal/domain"
	applog "sfdbtools/internal/services/log"
	"sfdbtools/internal/shared/compress"
	"sfdbtools/internal/shared/consts"
	"sfdbtools/internal/ui/progress"
	"strings"
	"sync"
)

func resolveMariaDBOrMySQLClient() (binPath string, binName string, err error) {
	// Default: mariadb client (mysql CLI compatible)
	if p, e := exec.LookPath("mariadb"); e == nil {
		return p, "mariadb", nil
	}
	// Fallback: mysql client
	if p, e := exec.LookPath("mysql"); e == nil {
		return p, "mysql", nil
	}
	return "", "", fmt.Errorf("binary client database tidak ditemukan: butuh 'mariadb' atau 'mysql' di PATH")
}

func isSSLMismatchServerNotSupport(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "tls/ssl error") && strings.Contains(msg, "server does not support")
}

func hasSkipSSLArg(args []string) bool {
	for _, a := range args {
		if strings.TrimSpace(strings.ToLower(a)) == "--skip-ssl" {
			return true
		}
	}
	return false
}

// BuildMySQLArgs membuat argument list untuk mysql command
func BuildMySQLArgs(profile *domain.ProfileInfo, database string, extraArgs ...string) []string {
	eff := profileconn.EffectiveDBInfo(profile)
	args := []string{
		fmt.Sprintf("--host=%s", eff.Host),
		fmt.Sprintf("--port=%d", eff.Port),
		fmt.Sprintf("--user=%s", profile.DBInfo.User),
		fmt.Sprintf("--password=%s", profile.DBInfo.Password),
	}

	// Tambahkan extra args jika ada
	args = append(args, extraArgs...)

	// Tambahkan database jika specified
	if database != "" {
		args = append(args, database)
	}

	return args
}

// MySQLExecSummary adalah ringkasan issue yang terdeteksi dari output client mysql/mariadb.
// Catatan: client biasanya menulis ERROR/WARNING ke stderr, tapi sebagian environment
// bisa menulis ke stdout juga (mis. konfigurasi tertentu).
type MySQLExecSummary struct {
	BinName      string
	SQLErrors    int
	SQLWarnings  int
	OtherOutputs int
}

func classifyMySQLClientLine(line string) (isErr bool, isWarn bool) {
	t := strings.TrimSpace(line)
	if t == "" {
		return false, false
	}
	l := strings.ToLower(t)
	// MariaDB/MySQL client common patterns:
	// - "ERROR 1418 (HY000) at line ...: ..."
	// - "Warning: Using a password on the command line interface can be insecure."
	if strings.HasPrefix(l, "error") || strings.Contains(l, " error ") {
		return true, false
	}
	if strings.HasPrefix(l, "warning") || strings.Contains(l, " warning") {
		return false, true
	}
	return false, false
}

// ExecuteMySQLCommand menjalankan mysql/mariadb client dengan stdin reader, sambil
// mendeteksi dan mencatat error/warning dari outputnya ke logger.
//
// Penting:
//   - Jangan pernah log args secara penuh karena mengandung password.
//   - Saat client dijalankan dengan --force/-f, SQL error dapat terjadi namun command tetap exit 0.
//     Fungsi ini tetap akan mencatat error/warning tersebut ke logs.
func ExecuteMySQLCommand(ctx context.Context, args []string, stdin io.Reader, logger applog.Logger) (*MySQLExecSummary, error) {
	binPath, binName, err := resolveMariaDBOrMySQLClient()
	if err != nil {
		return nil, err
	}
	if logger == nil {
		logger = applog.NullLogger()
	}

	cmd := exec.CommandContext(ctx, binPath, args...)
	cmd.Stdin = stdin
	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("%s command error: %w", binName, err)
	}
	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return nil, fmt.Errorf("%s command error: %w", binName, err)
	}

	const maxLoggedPerType = 50
	summary := &MySQLExecSummary{BinName: binName}
	var mu sync.Mutex
	errSuppressedLogged := false
	warnSuppressedLogged := false

	scanAndLog := func(streamName string, r io.Reader) {
		sc := bufio.NewScanner(r)
		// bump buffer for atypical long lines
		sc.Buffer(make([]byte, 1024), 1024*1024)
		for sc.Scan() {
			line := sc.Text()
			isErr, isWarn := classifyMySQLClientLine(line)
			if !isErr && !isWarn {
				mu.Lock()
				summary.OtherOutputs++
				mu.Unlock()
				continue
			}

			mu.Lock()
			if isErr {
				summary.SQLErrors++
				shouldLog := summary.SQLErrors <= maxLoggedPerType
				justSuppressed := summary.SQLErrors == maxLoggedPerType+1 && !errSuppressedLogged
				if justSuppressed {
					errSuppressedLogged = true
				}
				mu.Unlock()
				if shouldLog {
					logger.Errorf("[%s] %s", streamName, line)
				} else if justSuppressed {
					logger.Errorf("[%s] Terlalu banyak SQL error; log selanjutnya disembunyikan (>%d).", streamName, maxLoggedPerType)
				}
				continue
			}

			// warning
			summary.SQLWarnings++
			shouldLog := summary.SQLWarnings <= maxLoggedPerType
			justSuppressed := summary.SQLWarnings == maxLoggedPerType+1 && !warnSuppressedLogged
			if justSuppressed {
				warnSuppressedLogged = true
			}
			mu.Unlock()
			if shouldLog {
				logger.Warnf("[%s] %s", streamName, line)
			} else if justSuppressed {
				logger.Warnf("[%s] Terlalu banyak SQL warning; log selanjutnya disembunyikan (>%d).", streamName, maxLoggedPerType)
			}
		}
		// Ignore scanner error: biasanya karena token terlalu panjang; kita tetap lanjut (best-effort logging).
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("%s command error: %w", binName, err)
	}

	var wg sync.WaitGroup
	wg.Add(2)
	go func() { defer wg.Done(); scanAndLog("stdout", stdoutPipe) }()
	go func() { defer wg.Done(); scanAndLog("stderr", stderrPipe) }()

	runErr := cmd.Wait()
	wg.Wait()

	if runErr != nil {
		return summary, fmt.Errorf("%s command error: %w", binName, runErr)
	}

	return summary, nil
}

// RestoreFromFile melakukan restore database dari file backup
func RestoreFromFile(ctx context.Context, filePath string, targetDB string, profile *domain.ProfileInfo, encryptionKey string, logger applog.Logger) (*MySQLExecSummary, error) {
	spin := progress.NewSpinnerWithElapsed(fmt.Sprintf("Restore database %s dari %s", targetDB, filepath.Base(filePath)))
	spin.Start()
	defer spin.Stop()

	// Build mysql args dengan force flag
	args := BuildMySQLArgs(profile, targetDB, "-f")

	// Helper closure supaya retry bisa reopen file (stdin streaming tidak bisa diulang).
	execRestore := func(a []string) (*MySQLExecSummary, error) {
		reader, closers, err := OpenAndPrepareReader(filePath, encryptionKey)
		if err != nil {
			return nil, err
		}
		defer CloseReaders(closers)
		return ExecuteMySQLCommand(ctx, a, reader, logger)
	}

	// Execute mysql restore
	sum, err := execRestore(args)
	if err != nil {
		// Fallback: beberapa environment punya default SSL=ON/REQUIRED di client config.
		// Jika target server tidak support SSL, retry sekali dengan SSL dimatikan.
		if isSSLMismatchServerNotSupport(err) && !hasSkipSSLArg(args) {
			retryArgs := BuildMySQLArgs(profile, targetDB, "--skip-ssl", "-f")
			sum2, err2 := execRestore(retryArgs)
			if err2 == nil {
				return sum2, nil
			} else {
				return sum2, fmt.Errorf("gagal menjalankan mysql restore (retry --skip-ssl): %w", err2)
			}
		}
		return sum, fmt.Errorf("gagal menjalankan mysql restore: %w", err)
	}

	return sum, nil
}

// OpenAndPrepareReader membuka file dan menyiapkan reader dengan decrypt/decompress
// Returns: reader, list of closers, error
func OpenAndPrepareReader(filePath string, encryptionKey string) (io.Reader, []io.Closer, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, nil, fmt.Errorf("gagal membuka file: %w", err)
	}

	// Buffer file reads to improve large sequential throughput
	reader := io.Reader(bufio.NewReaderSize(file, 4*1024*1024))
	closers := []io.Closer{file}

	// Decrypt if encrypted
	isEncrypted := backupfile.IsEncryptedFile(filePath)
	if isEncrypted {
		decReader, err := crypto.NewStreamDecryptor(reader, encryptionKey)
		if err != nil {
			CloseReaders(closers)
			return nil, nil, fmt.Errorf("gagal membuat decrypting reader: %w", err)
		}
		reader = decReader
		closers = append(closers, io.NopCloser(decReader))
	}

	// Decompress if compressed
	compressionType := compress.DetectCompressionTypeFromFile(filePath)
	if compressionType != compress.CompressionType(consts.CompressionTypeNone) {
		decompReader, err := compress.NewDecompressingReader(reader, compressionType)
		if err != nil {
			CloseReaders(closers)
			return nil, nil, fmt.Errorf("gagal membuat decompressing reader: %w", err)
		}
		reader = decompReader
		closers = append(closers, decompReader)
	}

	return reader, closers, nil
}

// CloseReaders menutup semua readers dengan urutan terbalik
func CloseReaders(closers []io.Closer) {
	for i := len(closers) - 1; i >= 0; i-- {
		if closer := closers[i]; closer != nil {
			_ = closer.Close()
		}
	}
}

// RestoreUserGrants melakukan restore user grants dari file
func RestoreUserGrants(ctx context.Context, grantsFile string, profile *domain.ProfileInfo, logger applog.Logger) (*MySQLExecSummary, error) {
	if grantsFile == "" {
		return &MySQLExecSummary{BinName: ""}, nil
	}

	spin := progress.NewSpinnerWithElapsed(fmt.Sprintf("Restore user grants dari %s", filepath.Base(grantsFile)))
	spin.Start()
	defer spin.Stop()

	grantsSQL, err := os.ReadFile(grantsFile)
	if err != nil {
		return nil, fmt.Errorf("gagal membaca file grants: %w", err)
	}

	// Build mysql args tanpa database target
	args := BuildMySQLArgs(profile, "")

	// Execute mysql restore
	sum, err := ExecuteMySQLCommand(ctx, args, strings.NewReader(string(grantsSQL)), logger)
	if err != nil {
		if isSSLMismatchServerNotSupport(err) && !hasSkipSSLArg(args) {
			retryArgs := BuildMySQLArgs(profile, "", "--skip-ssl")
			sum2, err2 := ExecuteMySQLCommand(ctx, retryArgs, strings.NewReader(string(grantsSQL)), logger)
			if err2 == nil {
				return sum2, nil
			} else {
				return sum2, fmt.Errorf("gagal restore user grants (retry --skip-ssl): %w", err2)
			}
		}
		return sum, fmt.Errorf("gagal restore user grants: %w", err)
	}

	return sum, nil
}
