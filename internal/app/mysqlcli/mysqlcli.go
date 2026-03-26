package mysqlcli

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os/exec"
	profileconn "sfdbtools/internal/app/profile/connection"
	"sfdbtools/internal/domain"
	applog "sfdbtools/internal/services/log"
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

// IsSSLMismatchServerNotSupport mendeteksi kasus umum: client default SSL=ON/REQUIRED,
// tapi server tidak support SSL, sehingga perlu retry dengan --skip-ssl.
func IsSSLMismatchServerNotSupport(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "tls/ssl error") && strings.Contains(msg, "server does not support")
}

// HasSkipSSLArg mengecek apakah args sudah mengandung --skip-ssl.
func HasSkipSSLArg(args []string) bool {
	for _, a := range args {
		if strings.TrimSpace(strings.ToLower(a)) == "--skip-ssl" {
			return true
		}
	}
	return false
}

// BuildArgs membuat argument list untuk mysql/mariadb command.
func BuildArgs(profile *domain.ProfileInfo, database string, extraArgs ...string) []string {
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

// ExecSummary adalah ringkasan issue yang terdeteksi dari output client mysql/mariadb.
// Catatan: client biasanya menulis ERROR/WARNING ke stderr, tapi sebagian environment
// bisa menulis ke stdout juga (mis. konfigurasi tertentu).
type ExecSummary struct {
	BinName      string
	SQLErrors    int
	SQLWarnings  int
	OtherOutputs int
	ErrLines     []string
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

// Execute menjalankan mysql/mariadb client dengan stdin reader, sambil
// mendeteksi dan mencatat error/warning dari outputnya ke logger.
//
// Penting:
//   - Jangan pernah log args secara penuh karena mengandung password.
//   - Saat client dijalankan dengan --force/-f, SQL error dapat terjadi namun command tetap exit 0.
//     Fungsi ini tetap akan mencatat error/warning tersebut ke logs.
func Execute(ctx context.Context, args []string, stdin io.Reader, logger applog.Logger) (*ExecSummary, error) {
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
	summary := &ExecSummary{BinName: binName}
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
				if shouldLog {
					summary.ErrLines = append(summary.ErrLines, fmt.Sprintf("[%s] %s", streamName, line))
				} else if justSuppressed {
					summary.ErrLines = append(summary.ErrLines, fmt.Sprintf("[%s] Terlalu banyak SQL error; log selanjutnya disembunyikan (>%d).", streamName, maxLoggedPerType))
				}
				mu.Unlock()
				if shouldLog {
					logger.Debugf("[%s] %s", streamName, line)
				} else if justSuppressed {
					logger.Debugf("[%s] Terlalu banyak SQL error; log selanjutnya disembunyikan (>%d).", streamName, maxLoggedPerType)
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
				logger.Debugf("[%s] %s", streamName, line)
			} else if justSuppressed {
				logger.Debugf("[%s] Terlalu banyak SQL warning; log selanjutnya disembunyikan (>%d).", streamName, maxLoggedPerType)
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

// ExecuteWithSSLFallback menjalankan Execute() dan jika terdeteksi mismatch SSL (client default SSL ON,
// server tidak support), retry sekali dengan --skip-ssl.
func ExecuteWithSSLFallback(ctx context.Context, profile *domain.ProfileInfo, database string, extraArgs []string, stdin io.Reader, logger applog.Logger) (*ExecSummary, error) {
	args := BuildArgs(profile, database, extraArgs...)
	sum, err := Execute(ctx, args, stdin, logger)
	if err == nil {
		return sum, nil
	}

	if IsSSLMismatchServerNotSupport(err) && !HasSkipSSLArg(args) {
		retryArgs := BuildArgs(profile, database, append([]string{"--skip-ssl"}, extraArgs...)...)
		sum2, err2 := Execute(ctx, retryArgs, stdin, logger)
		if err2 == nil {
			return sum2, nil
		}
		return sum2, fmt.Errorf("gagal menjalankan mysql client (retry --skip-ssl): %w", err2)
	}
	return sum, err
}
