package execution

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"os/exec"
	"sfdbtools/internal/app/backup/execution"
	"sfdbtools/internal/app/mysqlcli"
	"sfdbtools/internal/domain"
	applog "sfdbtools/internal/services/log"
	"sfdbtools/internal/shared/execx"
	"sfdbtools/internal/ui/progress"
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

	if opts.TableName != "" {
		dumpArgs = append(dumpArgs, opts.TableName)
	}

	// 3. Build mysql client arguments
	mysqlArgs := mysqlcli.BuildArgs(opts.Profile, opts.TargetDB)

	// 4. Setup Commands
	dumpCmd := exec.CommandContext(ctx, dumpBin.Path, dumpArgs...)

	mysqlBin, _, err := mysqlcli.ResolveMariaDBOrMySQLClient()
	if err != nil {
		return err
	}
	mysqlCmd := exec.CommandContext(ctx, mysqlBin, mysqlArgs...)

	// Connect Pipe
	pr, pw := io.Pipe()
	dumpCmd.Stdout = pw

	// Transformer: Ganti nama sourceDB dengan targetDB di dalam stream SQL
	transformedReader := transformSQLStream(pr, opts.SourceDB, opts.TargetDB)
	mysqlCmd.Stdin = transformedReader

	// Capture errors
	dumpStderr, _ := dumpCmd.StderrPipe()
	mysqlStderr, _ := mysqlCmd.StderrPipe()

	// Start commands
	if err := dumpCmd.Start(); err != nil {
		_ = pw.Close()
		return fmt.Errorf("gagal memulai %s: %w", dumpBin.Name, err)
	}

	if err := mysqlCmd.Start(); err != nil {
		_ = pw.Close()
		_ = pr.Close()
		_ = dumpCmd.Process.Kill()
		return fmt.Errorf("gagal memulai mysql client: %w", err)
	}

	// Tampilkan spinner untuk feedback visual saat streaming berjalan
	spin := progress.NewSpinnerWithElapsed(fmt.Sprintf("Streaming %s -> %s", opts.SourceDB, opts.TargetDB))
	spin.Start()
	defer spin.Stop()

	// Log stderr in background
	go func() {
		sl := io.MultiReader(dumpStderr, mysqlStderr)
		scanner := bufio.NewScanner(sl)
		for scanner.Scan() {
			log.Debugf("[SQL CLI] %s", scanner.Text())
		}
	}()

	// Wait for commands
	errDumpChan := make(chan error, 1)
	go func() {
		errDumpChan <- dumpCmd.Wait()
		_ = pw.Close()
	}()

	errMysqlChan := make(chan error, 1)
	go func() {
		errMysqlChan <- mysqlCmd.Wait()
		_ = pr.Close()
	}()

	var errDump, errMysql error
	for i := 0; i < 2; i++ {
		select {
		case errDump = <-errDumpChan:
			if errDump != nil {
				_ = mysqlCmd.Process.Kill()
			}
		case errMysql = <-errMysqlChan:
			if errMysql != nil {
				_ = dumpCmd.Process.Kill()
			}
		case <-ctx.Done():
			_ = dumpCmd.Process.Kill()
			_ = mysqlCmd.Process.Kill()
			return ctx.Err()
		}
	}

	if errDump != nil {
		return fmt.Errorf("error pada %s: %w", dumpBin.Name, errDump)
	}
	if errMysql != nil {
		return fmt.Errorf("error pada mysql client: %w", errMysql)
	}

	return nil
}

// transformSQLStream melakukan penggantian nama database secara real-time pada stream SQL.
func transformSQLStream(r io.Reader, sourceDB, targetDB string) io.Reader {
	pr, pw := io.Pipe()

	go func() {
		defer pw.Close()
		scanner := bufio.NewScanner(r)

		const maxCapacity = 10 * 1024 * 1024
		buf := make([]byte, 64*1024)
		scanner.Buffer(buf, maxCapacity)

		// Pattern replacement
		p1Source := []byte("`" + sourceDB + "`.")
		p1Target := []byte("`" + targetDB + "`.")

		p2Source := []byte(" " + sourceDB + ".")
		p2Target := []byte(" " + targetDB + ".")

		for scanner.Scan() {
			line := scanner.Bytes()
			line = bytes.ReplaceAll(line, p1Source, p1Target)
			line = bytes.ReplaceAll(line, p2Source, p2Target)

			if _, err := pw.Write(line); err != nil {
				return
			}
			if _, err := pw.Write([]byte("\n")); err != nil {
				return
			}
		}
	}()

	return pr
}
