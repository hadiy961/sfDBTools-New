package execution

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"os/exec"
	"regexp"
	"sfdbtools/internal/app/backup/execution"
	"sfdbtools/internal/app/mysqlcli"
	"sfdbtools/internal/domain"
	applog "sfdbtools/internal/services/log"
	"sfdbtools/internal/shared/execx"
	"sfdbtools/internal/ui/progress"
	"time"
)

// PipingOptions berisi konfigurasi untuk proses streaming copy.
type PipingOptions struct {
	Profile      *domain.ProfileInfo
	SourceDB     string
	TargetDB     string
	TableName    string // Opsional, jika ingin copy table tertentu
	SchemaOnly   bool
	BaseDumpArgs string
	LimitSpeed   int64 // Bytes per second
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

	// Tambahkan flag untuk menghindari error GTID di server target
	dumpArgs = append(dumpArgs, "--set-gtid-purged=OFF")

	// 3. Build mysql client arguments
	mysqlArgs := mysqlcli.BuildArgs(opts.Profile, opts.TargetDB)

	// 4. Setup Commands
	// Kita gunakan CommandContext agar OS process otomatis di-kill jika context dibatalkan.
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
	// Serta hapus klausa DEFINER untuk menghindari error hak akses.
	transformedReader := transformSQLStream(pr, opts.SourceDB, opts.TargetDB)

	// Throttler: Batasi kecepatan jika diperlukan
	finalReader := io.Reader(transformedReader)
	var throttler *execx.ThrottledReader
	if opts.LimitSpeed > 0 {
		throttler = execx.NewThrottledReader(ctx, transformedReader, opts.LimitSpeed)
		finalReader = throttler
	}

	mysqlCmd.Stdin = finalReader

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

	// Metrics reporter goroutine
	stopMetrics := make(chan struct{})
	defer close(stopMetrics)
	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()
		var lastBytes int64
		for {
			select {
			case <-ticker.C:
				var currentBytes int64
				if throttler != nil {
					currentBytes = throttler.TotalBytes()
				}
				// Kita bisa hitung speed real-time walau tanpa throttler jika kita bungkus reader selalu
				// Tapi sementara tampilkan info throttle jika aktif
				if opts.LimitSpeed > 0 {
					speedMB := float64(opts.LimitSpeed) / (1024 * 1024)
					spin.Update(fmt.Sprintf("Streaming %s -> %s [%.2f MB/s limit]", opts.SourceDB, opts.TargetDB, speedMB))
				} else if currentBytes > 0 {
					// Hitung speed sesaat
					diff := currentBytes - lastBytes
					speedMB := float64(diff) / (1024 * 1024)
					spin.Update(fmt.Sprintf("Streaming %s -> %s [%.2f MB/s]", opts.SourceDB, opts.TargetDB, speedMB))
				}
				lastBytes = currentBytes
			case <-stopMetrics:
				return
			}
		}
	}()

	// Log stderr in background
	go func() {
		sl := io.MultiReader(dumpStderr, mysqlStderr)
		scanner := bufio.NewScanner(sl)
		for scanner.Scan() {
			log.Debugf("[SQL CLI] %s", scanner.Text())
		}
	}()

	// Wait for commands
	// Error handling diperkuat: jika salah satu gagal, kill yang lain.
	errDumpChan := make(chan error, 1)
	go func() {
		errDumpChan <- dumpCmd.Wait()
		_ = pw.Close() // Pastikan pipe ditutup agar mysqlCmd.Stdin mendapat EOF
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

	// Regex untuk DEFINER: /*!50013 DEFINER=`user`@`host` */ atau DEFINER=`user`@`host`
	// Kita buat di luar goroutine agar tidak compile berulang kali.
	reDefiner := regexp.MustCompile(`(?i)/\*!50017\s+DEFINER=\s*.*?\s*\*/`)
	reDefiner2 := regexp.MustCompile(`(?i)DEFINER=\s*` + "`" + `.*?` + "`" + `@` + "`" + `.*?` + "`")

	go func() {
		defer pw.Close()
		scanner := bufio.NewScanner(r)

		const maxCapacity = 10 * 1024 * 1024 // 10MB per line limit
		buf := make([]byte, 64*1024)
		scanner.Buffer(buf, maxCapacity)

		// Pattern database replacement
		p1Source := []byte("`" + sourceDB + "`.")
		p1Target := []byte("`" + targetDB + "`.")

		p2Source := []byte(" " + sourceDB + ".")
		p2Target := []byte(" " + targetDB + ".")

		for scanner.Scan() {
			line := scanner.Bytes()

			// 1. Ganti nama Database (Trigger/View fix)
			line = bytes.ReplaceAll(line, p1Source, p1Target)
			line = bytes.ReplaceAll(line, p2Source, p2Target)

			// 2. Strip DEFINER (Access Denied fix)
			line = reDefiner.ReplaceAll(line, []byte(""))
			line = reDefiner2.ReplaceAll(line, []byte(""))

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
