package copycmd

import (
	"context"
	"fmt"
	"runtime"
	"sfdbtools/internal/app/copy"
	appdeps "sfdbtools/internal/cli/deps"
	"sfdbtools/internal/cli/runner"
	"sfdbtools/internal/shared/execx"
	"sfdbtools/internal/ui/prompt"
	"sfdbtools/internal/ui/table"
	"time"

	"github.com/spf13/cobra"
)

var (
	copyDBProfile        string
	copyDBProfileKey     string
	copyDBTicket         string
	copyDBSchemaOnly     bool
	copyDBForce          bool
	copyDBBackupFirst    bool
	copyDBIncludeGrants  bool
	copyDBNonInteractive bool
	copyDBWorkers        int
	copyDBLimitSpeed     string
	copyDBVerify         bool
	copyDBSkipRoutines   bool
	copyDBSkipEvents     bool
	copyDBSkipTriggers   bool
)

// CmdCopyDB menyalin database utuh
var CmdCopyDB = &cobra.Command{
	Use:   "db [source_db...] [target_db]",
	Short: "Salin database utuh",
	Run: func(cmd *cobra.Command, args []string) {
		runner.RunContext(cmd, func(ctx context.Context) error {
			var sourceDBs []string
			targetDB := ""

			if len(args) > 1 {
				// Jika > 1 arg, maka arg terakhir dianggap target, sisanya source
				sourceDBs = args[:len(args)-1]
				targetDB = args[len(args)-1]
			} else if len(args) == 1 {
				sourceDBs = []string{args[0]}
			}

			svc := copy.NewService(appdeps.Deps.Logger, appdeps.Deps.Config)
			svc.SetTicket(copyDBTicket)

			// 1. Load Profile (Interactive picker if empty)
			profile, err := svc.LoadProfile(copyDBProfile, copyDBProfileKey, !copyDBNonInteractive)
			if err != nil {
				return err
			}

			// 2. Handle missing source databases (Interactive)
			if len(sourceDBs) == 0 {
				if copyDBNonInteractive {
					return fmt.Errorf("source database wajib diisi pada mode non-interaktif")
				}
				sourceDBs, err = svc.SelectDatabasesInteractive(ctx, profile)
				if err != nil {
					return err
				}
			}

			if len(sourceDBs) == 0 {
				fmt.Println("Tidak ada database yang dipilih.")
				return nil
			}

			// 3. Handle Target Name/Suffix
			var suffix string
			if len(sourceDBs) == 1 {
				if targetDB == "" && !copyDBNonInteractive {
					targetDB, _ = prompt.AskText("Masukkan nama database target:", prompt.WithDefault(fmt.Sprintf("%s_copy_%s", sourceDBs[0], time.Now().Format("20060102"))))
				}
			} else {
				// Multi-selection: ask for suffix instead of target name
				if !copyDBNonInteractive {
					suffix, _ = prompt.AskText("Banyak database dipilih. Masukkan akhiran (suffix) untuk nama target:", prompt.WithDefault(fmt.Sprintf("_copy_%s", time.Now().Format("20060102"))))
				} else if targetDB != "" {
					// In non-interactive, if targetDB is provided for multi-select, treat it as suffix
					suffix = targetDB
				}
			}

			// 4. Handle Copy Method (Piping vs Concurrent)
			useConcurrent := false
			if !copyDBNonInteractive {
				options := []string{
					"Direct Stream (Cepat, RAM-based)",
					"Concurrent Stream (Sangat Cepat, Multi-threading)",
				}
				choice, _, err := prompt.SelectOne("Pilih metode penyalinan:", options, 0)
				if err == nil {
					if choice == "Concurrent Stream (Sangat Cepat, Multi-threading)" {
						useConcurrent = true
					}
				}
			}

			// Parse Limit Speed
			var limitSpeed int64
			if copyDBLimitSpeed != "" {
				limitSpeed, err = execx.ParseSpeed(copyDBLimitSpeed)
				if err != nil {
					return fmt.Errorf("gagal parse limit speed: %w", err)
				}
			}

			// 5. Handle Copy Grants
			if !copyDBIncludeGrants && !copyDBNonInteractive {
				copyDBIncludeGrants, _ = prompt.Confirm("Salin hak akses (grants) user dari database sumber?", true)
			}

			// 6. Execution
			fmt.Println() // Space before start
						resList, err := svc.CopyDatabases(ctx, copy.CopyDatabasesOptions{
				CopyDatabaseOptions: copy.CopyDatabaseOptions{
					Profile:        profile,
					SchemaOnly:     copyDBSchemaOnly,
					UseConcurrent:  useConcurrent,
					Workers:        copyDBWorkers,
					LimitSpeed:     limitSpeed,
					Force:          copyDBForce,
					BackupFirst:    copyDBBackupFirst,
					IncludeGrants:  copyDBIncludeGrants,
					Verify:         copyDBVerify,
					SkipRoutines:   copyDBSkipRoutines,
					SkipEvents:     copyDBSkipEvents,
					SkipTriggers:   copyDBSkipTriggers,
					NonInteractive: copyDBNonInteractive,
				},
				SourceDBs:        sourceDBs,
				Suffix:           suffix,
				TargetDBIfSingle: targetDB,
			})
			if err != nil && len(resList) == 0 {
				return err
			}

			// 7. Final Summary
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
					icon + " " + r.SourceDB,
					r.TargetDB,
					r.Method,
					r.Duration.String(),
					r.Status,
				})
			}

			fmt.Printf("\n--- Ringkasan Copy Database ---\n")
			headers := []string{"Sumber", "Tujuan", "Metode", "Durasi", "Status"}
			table.Render(headers, results)

			fmt.Printf("\nSelesai: %d/%d database berhasil disalin.\n", successCount, len(sourceDBs))
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

	CmdCopyDB.Flags().StringVarP(&copyDBProfile, "profile", "p", "", "Nama atau path profil database")
	CmdCopyDB.Flags().StringVar(&copyDBProfileKey, "profile-key", "", "Kunci enkripsi profil (jika dienkripsi)")
	CmdCopyDB.Flags().StringVarP(&copyDBTicket, "ticket", "t", "", "Ticket number untuk audit")
	CmdCopyDB.Flags().IntVarP(&copyDBWorkers, "workers", "w", defaultWorkers, "Jumlah worker untuk concurrent copy (default: jumlah CPU)")
	CmdCopyDB.Flags().StringVar(&copyDBLimitSpeed, "limit-speed", "", "Batasi kecepatan transfer (misal: 10MB/s)")
	CmdCopyDB.Flags().BoolVar(&copyDBVerify, "verify", false, "Verifikasi integritas data dengan checksum setelah copy")
	CmdCopyDB.Flags().BoolVar(&copyDBSkipRoutines, "skip-routines", false, "Jangan salin stored procedures & functions")
	CmdCopyDB.Flags().BoolVar(&copyDBSkipEvents, "skip-events", false, "Jangan salin events")
	CmdCopyDB.Flags().BoolVar(&copyDBSkipTriggers, "skip-triggers", false, "Jangan salin triggers")
	CmdCopyDB.Flags().BoolVar(&copyDBSchemaOnly, "schema-only", false, "Hanya salin struktur (tanpa data)")
	CmdCopyDB.Flags().BoolVar(&copyDBForce, "force", false, "Timpa database target jika sudah ada (tanpa backup)")
	CmdCopyDB.Flags().BoolVar(&copyDBBackupFirst, "backup-first", false, "Backup database target terlebih dahulu sebelum menimpa")
	CmdCopyDB.Flags().BoolVar(&copyDBIncludeGrants, "include-grants", false, "Salin hak akses user (grants) dari database sumber")
	CmdCopyDB.Flags().BoolVar(&copyDBNonInteractive, "non-interactive", false, "Jalankan tanpa prompt interaktif")
}
