package copycmd

import (
	"context"
	"fmt"
	"sfdbtools/internal/app/copy"
	appdeps "sfdbtools/internal/cli/deps"
	"sfdbtools/internal/cli/runner"
	"sfdbtools/internal/shared/execx"
	"sfdbtools/internal/ui/prompt"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var (
	copyDBProfile        string
	copyDBProfileKey     string
	copyDBTicket         string
	copyDBSchemaOnly     bool
	copyDBUseDisk        bool
	copyDBForce          bool
	copyDBBackupFirst    bool
	copyDBIncludeGrants  bool
	copyDBNonInteractive bool
	copyDBWorkers        int
	copyDBLimitSpeed     string
	copyDBVerify         bool
	copyDBCompression    string
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

			// 4. Handle Copy Method (Piping vs Concurrent vs Disk)
			methodLabel := "Piping" 
			useConcurrent := false
			if copyDBUseDisk {
				methodLabel = "Disk-based"
			} else if !copyDBNonInteractive {
				options := []string{
					"Direct Stream (Cepat, RAM-based)",
					"Concurrent Stream (Sangat Cepat, Multi-threading)",
					"Disk-based (Aman, Dump file)",
				}
				choice, _, err := prompt.SelectOne("Pilih metode penyalinan:", options, 0)
				if err == nil {
					switch choice {
					case "Concurrent Stream (Sangat Cepat, Multi-threading)":
						methodLabel = "Concurrent"
						useConcurrent = true
					case "Disk-based (Aman, Dump file)":
						methodLabel = "Disk-based"
						copyDBUseDisk = true
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

			// 6. Loop Execution
			successCount := 0
			var results [][]string
			
			fmt.Println() // Space before start
			for i, db := range sourceDBs {
				// Check for graceful shutdown
				select {
				case <-ctx.Done():
					return ctx.Err()
				default:
				}

				currTarget := targetDB
				if len(sourceDBs) > 1 {
					currTarget = db + suffix
				}

				start := time.Now()
				fmt.Printf("[%d/%d] Kloning %s -> %s [%s]...\n", i+1, len(sourceDBs), db, currTarget, methodLabel)
				finalTarget, err := svc.CopyDatabase(ctx, profile, db, currTarget, copyDBSchemaOnly, copyDBUseDisk, useConcurrent, copyDBWorkers, limitSpeed, copyDBForce, copyDBBackupFirst, copyDBIncludeGrants, copyDBVerify, copyDBSkipRoutines, copyDBSkipEvents, copyDBSkipTriggers, copyDBNonInteractive)
				duration := time.Since(start).Round(time.Second)

				status := "Sukses"
				if err != nil {
					status = "Gagal"
					fmt.Printf("  ❌ Error: %v\n\n", err)
				} else {
					successCount++
					fmt.Printf("  ✅ Berhasil: %s (%s)\n\n", finalTarget, duration)
				}
				results = append(results, []string{
					db, 
					currTarget, 
					methodLabel, 
					duration.String(), 
					status,
				})
			}

			// 7. Final Summary
			fmt.Printf("\n--- Ringkasan Copy Database ---\n")
			fmt.Printf("% -30s -> % -30s [% -10s] [% -8s] [%s]\n", "Sumber", "Tujuan", "Metode", "Durasi", "Status")
			fmt.Println(strings.Repeat("-", 100))
			for _, res := range results {
				icon := "✓"
				if res[4] == "Gagal" { icon = "✗" }
				fmt.Printf("%s % -28s -> % -30s [% -10s] [% -8s] [%s]\n", icon, res[0], res[1], res[2], res[3], res[4])
			}

			fmt.Printf("\nSelesai: %d/%d database berhasil disalin.\n", successCount, len(sourceDBs))
			return nil
		})
	},
}

func init() {
	CmdCopyDB.Flags().StringVarP(&copyDBProfile, "profile", "p", "", "Nama atau path profil database")
	CmdCopyDB.Flags().StringVar(&copyDBProfileKey, "profile-key", "", "Kunci enkripsi profil (jika dienkripsi)")
	CmdCopyDB.Flags().StringVarP(&copyDBTicket, "ticket", "t", "", "Ticket number untuk audit")
	CmdCopyDB.Flags().IntVarP(&copyDBWorkers, "workers", "w", 4, "Jumlah worker untuk concurrent copy")
	CmdCopyDB.Flags().StringVar(&copyDBLimitSpeed, "limit-speed", "", "Batasi kecepatan transfer (misal: 10MB/s)")
	CmdCopyDB.Flags().BoolVar(&copyDBVerify, "verify", false, "Verifikasi integritas data dengan checksum setelah copy")
	CmdCopyDB.Flags().StringVar(&copyDBCompression, "compression", "gzip", "Metode kompresi untuk disk-based (gzip, pgzip, zstd)")
	CmdCopyDB.Flags().BoolVar(&copyDBSkipRoutines, "skip-routines", false, "Jangan salin stored procedures & functions")
	CmdCopyDB.Flags().BoolVar(&copyDBSkipEvents, "skip-events", false, "Jangan salin events")
	CmdCopyDB.Flags().BoolVar(&copyDBSkipTriggers, "skip-triggers", false, "Jangan salin triggers")
	CmdCopyDB.Flags().BoolVar(&copyDBSchemaOnly, "schema-only", false, "Hanya salin struktur (tanpa data)")
	CmdCopyDB.Flags().BoolVar(&copyDBUseDisk, "use-disk", false, "Gunakan media disk (dump & restore) alih-alih streaming RAM")
	CmdCopyDB.Flags().BoolVar(&copyDBForce, "force", false, "Timpa database target jika sudah ada (tanpa backup)")
	CmdCopyDB.Flags().BoolVar(&copyDBBackupFirst, "backup-first", false, "Backup database target terlebih dahulu sebelum menimpa")
	CmdCopyDB.Flags().BoolVar(&copyDBIncludeGrants, "include-grants", false, "Salin hak akses user (grants) dari database sumber")
	CmdCopyDB.Flags().BoolVar(&copyDBNonInteractive, "non-interactive", false, "Jalankan tanpa prompt interaktif")
}
