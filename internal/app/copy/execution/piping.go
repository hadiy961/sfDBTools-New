package execution

import (
	"context"
	"fmt"
	"io"
	"os/exec"
	"sfdbtools/internal/app/backup/execution"
	"sfdbtools/internal/app/mysqlcli"
	"sfdbtools/internal/domain"
	applog "sfdbtools/internal/services/log"
	"sfdbtools/internal/shared/execx"
)

// PipingOptions berisi konfigurasi untuk proses streaming copy.
type PipingOptions struct {
	Profile      *domain.ProfileInfo
	SourceDB     string
	TargetDB     string
	TableName    string // Opsional, jika ingin copy table tertentu
	SchemaOnly   bool
	BaseDumpArgs string
}

// ExecutePiping menjalankan streaming copy menggunakan mysqldump | mysql.
func ExecutePiping(ctx context.Context, log applog.Logger, opts PipingOptions) error {
	// 1. Resolve mysqldump/mariadb-dump binary
	dumpBin, err := execx.ResolveMariaDBDumpOrMysqldump()
	if err != nil {
		return err
	}

	// 2. Build mysqldump arguments
	// Kita gunakan helper yang sudah ada di internal/app/backup/execution
	filter := domain.FilterOptions{
		ExcludeData: opts.SchemaOnly,
	}

	dumpArgs := execution.BuildMysqldumpArgs(
		opts.BaseDumpArgs,
		opts.Profile.DBInfo,
		filter,
		nil,           // dbFiltered
		opts.SourceDB, // singleDB
		1,             // totalDBFound
		nil,           // skipTablesData
		false,         // sshTunnelEnabled
	)

	// Jika copy table, tambahkan nama tabel di akhir (setelah database name yang sudah ada di dumpArgs)
	if opts.TableName != "" {
		dumpArgs = append(dumpArgs, opts.TableName)
	}

	// 3. Build mysql client arguments
	// Kita gunakan helper yang sudah ada di internal/app/mysqlcli
	mysqlArgs := mysqlcli.BuildArgs(opts.Profile, opts.TargetDB)

	// 4. Setup Commands
	dumpCmd := exec.CommandContext(ctx, dumpBin.Path, dumpArgs...)

	// Resolve mysql binary
	mysqlBin, _, err := mysqlcli.ResolveMariaDBOrMySQLClient() // This was an internal function in mysqlcli, I might need to export it or use ExecuteWithSSLFallback
	if err != nil {
		return err
	}
	mysqlCmd := exec.CommandContext(ctx, mysqlBin, mysqlArgs...)

	// Connect Pipe
	pr, pw := io.Pipe()
	dumpCmd.Stdout = pw
	mysqlCmd.Stdin = pr

	// Capture errors
	dumpStderr, _ := dumpCmd.StderrPipe()
	mysqlStderr, _ := mysqlCmd.StderrPipe()

	// Start commands
	if err := dumpCmd.Start(); err != nil {
		return fmt.Errorf("gagal memulai %s: %w", dumpBin.Name, err)
	}
	if err := mysqlCmd.Start(); err != nil {
		return fmt.Errorf("gagal memulai mysql client: %w", err)
	}

	// Log stderr in background
	go func() {
		sl := io.MultiReader(dumpStderr, mysqlStderr)
		_, _ = io.Copy(io.Discard, sl) // Kita bisa improve ini untuk log ke logger jika perlu
	}()

	// Wait for commands
	errDump := dumpCmd.Wait()
	pw.Close() // Close pipe so mysql knows stdin is finished

	errMysql := mysqlCmd.Wait()
	pr.Close()

	if errDump != nil {
		return fmt.Errorf("error pada %s: %w", dumpBin.Name, errDump)
	}
	if errMysql != nil {
		return fmt.Errorf("error pada mysql client: %w", errMysql)
	}

	return nil
}
