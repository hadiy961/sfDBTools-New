// File : internal/restore/restore_executor.go
// Deskripsi : MySQL restore executor - modular command execution
// Author : Hadiyatna Muflihun
// Tanggal : 2025-11-11
// Last Modified : 2025-11-11

package restore

import (
	"context"
	"fmt"
	"io"
	"os/exec"
	"sfDBTools/pkg/ui"
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
	cmd.Stdin = reader

	// Capture output
	output, err := cmd.CombinedOutput()

	if err != nil {
		if spin != nil {
			spin.Stop()
			fmt.Println()
		}
		return fmt.Errorf("mysql restore failed: %w\nOutput: %s", err, string(output))
	}

	if len(output) > 0 {
		s.Log.Debugf("MySQL output: %s", string(output))
	}

	return nil
}
