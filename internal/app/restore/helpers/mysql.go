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
	"path/filepath"
	backupfile "sfdbtools/internal/app/backup/helpers/file"
	"sfdbtools/internal/app/mysqlcli"
	profileconn "sfdbtools/internal/app/profile/connection"
	"sfdbtools/internal/crypto"
	"sfdbtools/internal/domain"
	applog "sfdbtools/internal/services/log"
	"sfdbtools/internal/shared/compress"
	"sfdbtools/internal/shared/consts"
	"sfdbtools/internal/ui/progress"
	"strings"
)

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

// ExecuteMySQLCommand menjalankan mysql/mariadb client dengan stdin reader, sambil
// mendeteksi dan mencatat error/warning dari outputnya ke logger.
//
// Penting:
//   - Jangan pernah log args secara penuh karena mengandung password.
//   - Saat client dijalankan dengan --force/-f, SQL error dapat terjadi namun command tetap exit 0.
//     Fungsi ini tetap akan mencatat error/warning tersebut ke logs.
func ExecuteMySQLCommand(ctx context.Context, args []string, stdin io.Reader, logger applog.Logger) (*MySQLExecSummary, error) {
	sum, err := mysqlcli.Execute(ctx, args, stdin, logger)
	if sum == nil {
		return nil, err
	}
	return &MySQLExecSummary{
		BinName:      sum.BinName,
		SQLErrors:    sum.SQLErrors,
		SQLWarnings:  sum.SQLWarnings,
		OtherOutputs: sum.OtherOutputs,
	}, err
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
		if mysqlcli.IsSSLMismatchServerNotSupport(err) && !mysqlcli.HasSkipSSLArg(args) {
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
		if mysqlcli.IsSSLMismatchServerNotSupport(err) && !mysqlcli.HasSkipSSLArg(args) {
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
