package copycmd

import (
	"context"
	"fmt"
	"runtime"
	"sfdbtools/internal/app/copy"
	appdeps "sfdbtools/internal/cli/deps"
	"sfdbtools/internal/cli/runner"
	"sfdbtools/internal/services/notify"
	"sfdbtools/internal/ui/prompt"
	"sfdbtools/internal/ui/table"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var (
	copyTableProfile        string
	copyTableProfileKey     string
	copyTableTicket         string
	copyTableSchemaOnly     bool
	copyTableForce          bool
	copyTableBackupFirst    bool
	copyTableIncludeGrants  bool
	copyTableNonInteractive bool
	copyTableWorkers        int
	copyTableVerify         bool
)

// CmdCopyTable menyalin tabel spesifik
var CmdCopyTable = &cobra.Command{
	Use:   "table [source_db.table...] [target_db]",
	Short: "Salin tabel spesifik",
	Run: func(cmd *cobra.Command, args []string) {
		runner.RunContext(cmd, func(ctx context.Context) error {
			var sourceDB string
			var sourceTables []string
			targetArg := ""

			if len(args) > 1 {
				for _, arg := range args[:len(args)-1] {
					parts := strings.Split(arg, ".")
					if len(parts) == 2 {
						sourceDB = parts[0]
						sourceTables = append(sourceTables, parts[1])
					}
				}
				targetArg = args[len(args)-1]
			} else if len(args) == 1 {
				parts := strings.Split(args[0], ".")
				if len(parts) == 2 {
					sourceDB = parts[0]
					sourceTables = append(sourceTables, parts[1])
				}
			}

			svc := copy.NewService(appdeps.Deps.Logger, appdeps.Deps.Config)
			svc.SetTicket(copyTableTicket)

			// 1. Load Profile
			profile, err := svc.LoadProfile(copyTableProfile, copyTableProfileKey, !copyTableNonInteractive)
			if err != nil {
				return err
			}

			// 2. Handle missing source arg (Interactive)
			if len(sourceTables) == 0 {
				if copyTableNonInteractive {
					return fmt.Errorf("source 'db.table' wajib diisi pada mode non-interaktif")
				}
				sourceDB, sourceTables, err = svc.SelectTablesInteractive(ctx, profile)
				if err != nil {
					return err
				}
			}

			if len(sourceTables) == 0 {
				fmt.Println("Tidak ada tabel yang dipilih.")
				return nil
			}

			// 3. Handle Target Database
			targetDB := sourceDB
			if targetArg != "" && !strings.Contains(targetArg, ".") {
				targetDB = targetArg
			} else if !copyTableNonInteractive {
				targetDB, err = svc.SelectTargetDatabaseInteractive(ctx, profile, sourceDB)
				if err != nil {
					return err
				}
			}

			// 4. Handle Target Table Name/Suffix
			var tableSuffix string
			var specificTargetTable string

			if len(sourceTables) == 1 {
				if !copyTableNonInteractive {
					specificTargetTable, _ = prompt.AskText("Masukkan nama tabel tujuan:", prompt.WithDefault(sourceTables[0]))
				} else {
					specificTargetTable = sourceTables[0]
				}
			} else {
				if !copyTableNonInteractive {
					defaultSuffix := ""
					if targetDB == sourceDB {
						defaultSuffix = "_copy"
					}
					tableSuffix, _ = prompt.AskText("Banyak tabel dipilih. Masukkan akhiran (suffix) untuk nama tabel target:", prompt.WithDefault(defaultSuffix))
				}
			}

			// 5. Handle Copy Grants
			if !copyTableIncludeGrants && !copyTableNonInteractive {
				copyTableIncludeGrants, _ = prompt.Confirm("Salin hak akses (grants) user untuk database target?", true)
			}

			// 6. Execution
			fmt.Println()
						resList, err := svc.CopyTablesConcurrent(ctx, copy.CopyTablesConcurrentOptions{
				CopyTableOptions: copy.CopyTableOptions{
					Profile:        profile,
					SourceDB:       sourceDB,
					TargetDB:       targetDB,
					SchemaOnly:     copyTableSchemaOnly,
					Force:          copyTableForce,
					BackupFirst:    copyTableBackupFirst,
					IncludeGrants:  copyTableIncludeGrants,
					Verify:         copyTableVerify,
					NonInteractive: copyTableNonInteractive,
				},
				SourceTables:        sourceTables,
				Suffix:              tableSuffix,
				TargetTableIfSingle: specificTargetTable,
				Workers:             copyTableWorkers,
			})

			// Kirim notifikasi
			var totalDuration time.Duration
			for _, r := range resList {
				totalDuration += r.Duration
			}
			msg := notify.BuildCopyTableMessage(sourceDB, targetDB, sourceTables, totalDuration, err, copyTableTicket)
			appdeps.Deps.NotifyService.Send(msg)

			if err != nil && len(resList) == 0 {
				return err
			}

			// 7. Final Summary Table
			var results [][]string
			successCount := 0
			for _, r := range resList {
				icon := "✅"
				if r.Error != nil {
					icon = "❌"
				} else {
					successCount++
				}
				results = append(results, []string{
					icon + " " + fmt.Sprintf("%s.%s", r.SourceDB, r.SourceTable),
					fmt.Sprintf("%s.%s", r.TargetDB, r.TargetTable),
					r.Status,
					r.VerifyStatus,
					r.Duration.String(),
				})
			}

			fmt.Printf("\n--- Ringkasan Copy Tabel ---\n")
			headers := []string{"Sumber", "Tujuan", "Status", "Checksum", "Durasi"}
			table.Render(headers, results)

			fmt.Printf("\nSelesai: %d/%d tabel berhasil disalin.\n", successCount, len(sourceTables))
			return nil
		})
	},
}

func init() {
	defaultWorkers := runtime.NumCPU()
	if defaultWorkers > 16 {
		defaultWorkers = 16
	}
	if defaultWorkers < 1 {
		defaultWorkers = 1
	}

	CmdCopyTable.Flags().StringVarP(&copyTableProfile, "profile", "p", "", "Nama atau path profil database")
	CmdCopyTable.Flags().StringVar(&copyTableProfileKey, "profile-key", "", "Kunci enkripsi profil (jika dienkripsi)")
	CmdCopyTable.Flags().StringVarP(&copyTableTicket, "ticket", "t", "", "Ticket number untuk audit")
	CmdCopyTable.Flags().IntVarP(&copyTableWorkers, "workers", "w", defaultWorkers, "Jumlah worker untuk concurrent table copy (default: jumlah CPU)")
	CmdCopyTable.Flags().BoolVar(&copyTableVerify, "verify", false, "Verifikasi integritas data dengan checksum setelah copy")
	CmdCopyTable.Flags().BoolVar(&copyTableSchemaOnly, "schema-only", false, "Hanya salin struktur (tanpa data)")
	CmdCopyTable.Flags().BoolVar(&copyTableForce, "force", false, "Timpa tabel target jika sudah ada (tanpa backup)")
	CmdCopyTable.Flags().BoolVar(&copyTableBackupFirst, "backup-first", false, "Backup tabel target terlebih dahulu sebelum menimpa")
	CmdCopyTable.Flags().BoolVar(&copyTableIncludeGrants, "include-grants", false, "Salin hak akses user (grants) dari database sumber")
	CmdCopyTable.Flags().BoolVar(&copyTableNonInteractive, "non-interactive", false, "Jalankan tanpa prompt interaktif")
}
