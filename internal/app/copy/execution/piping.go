package execution

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"os/exec"
	"regexp"
	backup_exec "sfdbtools/internal/app/backup/execution"
	"sfdbtools/internal/app/mysqlcli"
	"sfdbtools/internal/domain"
	applog "sfdbtools/internal/services/log"
	"sfdbtools/internal/shared/execx"
	"sfdbtools/internal/ui/progress"
	"strings"
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
	HideProgress bool  // Jika true, jangan tampilkan spinner/progress internal
	Label        string // Label kustom untuk progress bar
	Force        bool   // Jika true, gunakan --force di mysqldump dan mysql
}

// ExecutePiping menjalankan streaming copy menggunakan mysqldump | mysql.
// Sekarang mendukung retry otomatis jika ada argumen mysqldump yang tidak didukung.
func ExecutePiping(ctx context.Context, log applog.Logger, opts PipingOptions) error {
	// 1. Resolve mysqldump/mariadb-dump binary
	dumpBin, err := execx.ResolveMariaDBDumpOrMysqldump()
	if err != nil {
		return err
	}

	// 2. Build initial mysqldump arguments
	filter := domain.FilterOptions{
		ExcludeData: opts.SchemaOnly,
	}

	dumpArgs := backup_exec.BuildMysqldumpArgs(
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

	// Tambahkan flag --force jika diminta (Best effort)
	if opts.Force {
		dumpArgs = append(dumpArgs, "--force")
		mysqlArgs = append(mysqlArgs, "--force")
	}

	// 4. Resolve mysql binary
	mysqlBin, _, err := mysqlcli.ResolveMariaDBOrMySQLClient()
	if err != nil {
		return err
	}

	// Retry Loop
	attempts := 0
	const maxAttempts = 3

	for {
		attempts++

		// Setup Commands
		dumpCmd := exec.CommandContext(ctx, dumpBin.Path, dumpArgs...)
		mysqlCmd := exec.CommandContext(ctx, mysqlBin, mysqlArgs...)

		// Connect Pipe
		pr, pw := io.Pipe()
		dumpCmd.Stdout = pw

		// Transformer
		transformedReader := transformSQLStream(pr, opts.SourceDB, opts.TargetDB)
		tPR, _ := transformedReader.(*io.PipeReader)

		// Throttler
		finalReader := io.Reader(transformedReader)
		var throttler *execx.ThrottledReader
		if opts.LimitSpeed > 0 {
			throttler = execx.NewThrottledReader(ctx, transformedReader, opts.LimitSpeed)
			finalReader = throttler
		}
		mysqlCmd.Stdin = finalReader

		// Capture stderr for error detection
		var dumpStderrBuf, mysqlStderrBuf bytes.Buffer
		dumpCmd.Stderr = &dumpStderrBuf
		mysqlCmd.Stderr = &mysqlStderrBuf

		// Feedback visual
		var spin *progress.Spinner
		if !opts.HideProgress {
			label := opts.Label
			if label == "" {
				label = fmt.Sprintf("Streaming %s -> %s", opts.SourceDB, opts.TargetDB)
			}
			spin = progress.NewSpinnerWithElapsed(label)
			spin.Start()
		}

		// Execute
		if err := dumpCmd.Start(); err != nil {
			if spin != nil {
				spin.Stop()
			}
			_ = pw.Close()
			return fmt.Errorf("gagal memulai %s: %w", dumpBin.Name, err)
		}

		if err := mysqlCmd.Start(); err != nil {
			if spin != nil {
				spin.Stop()
			}
			_ = pw.Close()
			_ = pr.Close()
			if tPR != nil {
				_ = tPR.Close()
			}
			_ = dumpCmd.Process.Kill()
			return fmt.Errorf("gagal memulai mysql client: %w", err)
		}

		// Metrics reporter
		stopMetrics := make(chan struct{})
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
					if spin != nil {
						label := opts.Label
						if label == "" {
							label = fmt.Sprintf("Streaming %s -> %s", opts.SourceDB, opts.TargetDB)
						}
						
						if opts.LimitSpeed > 0 {
							speedMB := float64(opts.LimitSpeed) / (1024 * 1024)
							spin.Update(fmt.Sprintf("%s [%.2f MB/s limit]", label, speedMB))
						} else if currentBytes > 0 {
							diff := currentBytes - lastBytes
							speedMB := float64(diff) / (1024 * 1024)
							spin.Update(fmt.Sprintf("%s [%.2f MB/s]", label, speedMB))
						}
					}
					lastBytes = currentBytes
				case <-stopMetrics:
					return
				}
			}
		}()

		// Wait
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
		timedOut := false

		numFinished := 0
		for numFinished < 2 {
			select {
			case e := <-errDumpChan:
				errDump = e
				numFinished++
				if errDump != nil {
					_ = mysqlCmd.Process.Kill()
					_ = pw.Close()
					_ = pr.Close()
					if tPR != nil {
						_ = tPR.Close()
					}
				numFinished = 2 // Stop waiting if one fails critically
				}
			case e := <-errMysqlChan:
				errMysql = e
				numFinished++
				if errMysql != nil {
					_ = dumpCmd.Process.Kill()
					_ = pw.Close()
					_ = pr.Close()
					if tPR != nil {
						_ = tPR.Close()
					}
				numFinished = 2
				}
			case <-ctx.Done():
				_ = dumpCmd.Process.Kill()
				_ = mysqlCmd.Process.Kill()
				_ = pw.Close()
				_ = pr.Close()
				if tPR != nil {
					_ = tPR.Close()
				}
			timedOut = true
			numFinished = 2
			}
		}

		close(stopMetrics)
		if spin != nil {
			spin.Stop()
		}

		if timedOut {
			return ctx.Err()
		}

		// Error Detection & Retry Logic
		if errDump != nil {
			stderrStr := dumpStderrBuf.String()
			log.Debugf("[%s Error] %s", dumpBin.Name, strings.TrimSpace(stderrStr))

			// Cek apakah ada opsi yang tidak didukung
			if newArgs, removed, canRetry := backup_exec.RemoveUnsupportedMysqldumpOption(dumpArgs, stderrStr); canRetry && attempts < maxAttempts {
				log.Warnf("Opsi dump '%s' tidak didukung oleh binary sistem, mencoba ulang tanpa opsi tersebut...", removed)
				dumpArgs = newArgs
				continue // Retry!
			}

			// Cek SSL mismatch
			if backup_exec.IsSSLMismatchRequiredButServerNoSupport(stderrStr) && attempts < maxAttempts {
				if newArgs, added := backup_exec.AddDisableSSLArgs(dumpArgs); added {
					log.Warnf("SSL mismatch terdeteksi, mencoba ulang dengan --skip-ssl...")
					dumpArgs = newArgs
					continue // Retry!
				}
			}

			return fmt.Errorf("error pada %s: %w", dumpBin.Name, errDump)
		}

		if errMysql != nil {
			log.Debugf("[MySQL Error] %s", strings.TrimSpace(mysqlStderrBuf.String()))
			return fmt.Errorf("error pada mysql client: %w", errMysql)
		}

		break // Success!
	}

	return nil
}

// transformSQLStream melakukan penggantian nama database secara real-time pada stream SQL.
func transformSQLStream(r io.Reader, sourceDB, targetDB string) io.Reader {
	pr, pw := io.Pipe()

	// Regex untuk DEFINER: /*!50013 DEFINER=`user`@`host` */ atau DEFINER=`user`@`host`
	reDefiner := regexp.MustCompile(`(?i)/\*!50017\s+DEFINER=\s*.*?\s*\*/`)
	reDefiner2 := regexp.MustCompile(`(?i)DEFINER=\s*` + "`" + `.*?` + "`" + `@` + "`" + `.*?` + "`")

	go func() {
		defer pw.Close()
		scanner := bufio.NewScanner(r)

		const maxCapacity = 10 * 1024 * 1024 // 10MB per line limit
		buf := make([]byte, 64*1024)
		scanner.Buffer(buf, maxCapacity)

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
