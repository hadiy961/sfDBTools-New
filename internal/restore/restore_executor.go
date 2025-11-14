// File : internal/restore/restore_executor.go
// Deskripsi : MySQL restore executor - modular command execution
// Author : Hadiyatna Muflihun
// Tanggal : 2025-11-11
// Last Modified : 2025-11-14

package restore

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os/exec"
	"sfDBTools/pkg/ui"
)

const (
	// pipeBufferSize untuk buffered pipe ke mysql stdin (256KB untuk smooth data flow)
	pipeBufferSize = 256 * 1024
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
	}

	// Tambahkan force flag jika diperlukan (untuk continue on errors)
	if opts.Force {
		args = append(args, "--force")
	}

	// Tambahkan target database jika specified (untuk single restore)
	if opts.TargetDatabase != "" {
		args = append(args, opts.TargetDatabase)
	}

	// Setup spinner jika diminta
	var spin *ui.SpinnerWithElapsed
	if opts.WithSpinner {
		spin = ui.NewSpinnerWithElapsed("Melakukan restore database")
		spin.Start()
		defer spin.Stop()
	}

	// Execute command
	cmd := exec.CommandContext(ctx, "mysql", args...)

	// Setup stdin pipe dengan buffer untuk smooth data flow
	stdinPipe, err := cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("gagal setup stdin pipe: %w", err)
	}

	// Start command
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("gagal start mysql command: %w", err)
	}

	// Copy data dari reader ke stdin dengan buffer
	bufWriter := bufio.NewWriterSize(stdinPipe, pipeBufferSize)
	_, copyErr := io.Copy(bufWriter, reader)

	// Flush buffer dan close stdin pipe
	if flushErr := bufWriter.Flush(); flushErr != nil && copyErr == nil {
		copyErr = flushErr
	}
	if closeErr := stdinPipe.Close(); closeErr != nil && copyErr == nil {
		copyErr = closeErr
	}

	// Wait for command to finish
	waitErr := cmd.Wait()

	// Handle errors
	if copyErr != nil {
		if spin != nil {
			spin.Stop()
			fmt.Println()
		}
		return fmt.Errorf("gagal copy data ke mysql: %w", copyErr)
	}

	if waitErr != nil {
		if spin != nil {
			spin.Stop()
			fmt.Println()
		}
		return fmt.Errorf("mysql restore failed: %w", waitErr)
	}

	return nil
}
