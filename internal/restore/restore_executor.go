// File : internal/restore/restore_executor.go
// Deskripsi : MySQL restore executor - modular command execution
// Author : Hadiyatna Muflihun
// Tanggal : 2025-11-11
// Last Modified : 2025-11-14

package restore

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"os/exec"
	"sfDBTools/pkg/ui"
	"strings"
)

const (
	// pipeBufferSize untuk buffered pipe ke mysql stdin (1MB untuk smooth data flow dan mengurangi write calls)
	pipeBufferSize = 1024 * 1024
)

// MysqlRestoreOptions berisi opsi untuk eksekusi mysql restore command
type MysqlRestoreOptions struct {
	Host           string
	Port           int
	User           string
	Password       string
	TargetDatabase string // Kosong untuk restore all (combined backup)
	Force          bool   // Continue on errors
	WithSpinner    bool   // Tampilkan spinner dengan elapsed time
}

// executeMysqlCommand menjalankan mysql command untuk restore dengan konfigurasi yang diberikan
func (s *Service) executeMysqlCommand(ctx context.Context, reader io.Reader, opts MysqlRestoreOptions) error {
	// Build mysql command args
	args := []string{
		fmt.Sprintf("--host=%s", opts.Host),
		fmt.Sprintf("--port=%d", opts.Port),
		fmt.Sprintf("--user=%s", opts.User),
		fmt.Sprintf("--password=%s", opts.Password),
		// Timeout settings untuk prevent connection loss
		"--connect-timeout=300",           // 5 menit untuk connect timeout
		"--max-allowed-packet=1073741824", // 1GB max packet size
		// Note: net_read_timeout dan net_write_timeout tidak didukung di mysql client CLI
		// Timeout ini harus diset di server atau via session variable dalam SQL
	}

	// Tambahkan force flag jika diperlukan (untuk continue on errors)
	if opts.Force {
		args = append(args, "--force")
	}

	// Tambahkan target database jika specified (untuk single restore)
	if opts.TargetDatabase != "" {
		args = append(args, opts.TargetDatabase)
	}

	// Execute command
	cmd := exec.CommandContext(ctx, "mysql", args...)

	// Setup stdin pipe dengan buffer untuk smooth data flow
	stdinPipe, err := cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("gagal setup stdin pipe: %w", err)
	}

	// Capture stderr untuk mendapatkan error detail dari mysql
	var stderrBuf bytes.Buffer
	cmd.Stderr = &stderrBuf

	// Start command
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("gagal start mysql command: %w", err)
	}

	s.Log.Debugf("MySQL command started with PID: %d", cmd.Process.Pid)
	// Setup spinner jika diminta
	var spin *ui.SpinnerWithElapsed
	if opts.WithSpinner {
		spin = ui.NewSpinnerWithElapsed("Melakukan restore database")
		spin.Start()
		defer spin.Stop()
	}

	// Copy data dari reader ke stdin dengan buffer
	bufWriter := bufio.NewWriterSize(stdinPipe, pipeBufferSize)
	bytesCopied, copyErr := io.Copy(bufWriter, reader)

	// Flush buffer dan close stdin pipe
	if flushErr := bufWriter.Flush(); flushErr != nil && copyErr == nil {
		s.Log.Errorf("Gagal flush buffer: %v", flushErr)
		copyErr = flushErr
	}
	if closeErr := stdinPipe.Close(); closeErr != nil && copyErr == nil {
		s.Log.Errorf("Gagal close stdin pipe: %v", closeErr)
		copyErr = closeErr
	}

	// Wait for command to finish
	waitErr := cmd.Wait()

	// Get stderr output untuk error detail
	stderrOutput := strings.TrimSpace(stderrBuf.String())

	// Handle errors dengan detail lebih lengkap
	if copyErr != nil {
		if spin != nil {
			spin.Stop()
			fmt.Println()
		}
		// Log detail untuk troubleshooting
		s.Log.Errorf("Error saat copy data: %v (bytes copied: %d)", copyErr, bytesCopied)
		if stderrOutput != "" {
			s.Log.Errorf("MySQL stderr: %s", stderrOutput)
			return fmt.Errorf("gagal copy data ke mysql: %w | MySQL error: %s", copyErr, stderrOutput)
		}
		return fmt.Errorf("gagal copy data ke mysql: %w", copyErr)
	}

	if waitErr != nil {
		if spin != nil {
			spin.Stop()
			fmt.Println()
		}
		s.Log.Errorf("MySQL process error: %v (bytes copied: %d)", waitErr, bytesCopied)
		if stderrOutput != "" {
			s.Log.Errorf("MySQL stderr: %s", stderrOutput)
			return fmt.Errorf("mysql restore failed: %w | MySQL error: %s", waitErr, stderrOutput)
		}
		return fmt.Errorf("mysql restore failed: %w", waitErr)
	}

	// Log warning jika ada output di stderr tapi tidak error (bisa berisi warnings)
	if stderrOutput != "" {
		s.Log.Warnf("MySQL warnings: %s", stderrOutput)
	}

	return nil
}
