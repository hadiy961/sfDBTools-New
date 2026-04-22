package copycmd

import (
	"context"
	"fmt"
	"sfdbtools/internal/app/copy"
	appdeps "sfdbtools/internal/cli/deps"
	"sfdbtools/internal/cli/runner"
	"sfdbtools/internal/ui/prompt"
	"strings"

	"github.com/spf13/cobra"
)

var (
	copyTableProfile        string
	copyTableProfileKey     string
	copyTableTicket         string
	copyTableSchemaOnly     bool
	copyTableForce          bool
	copyTableBackupFirst    bool
	copyTableNonInteractive bool
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
				// Multi arg: last one is target (if no dot) or another source
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

			// 5. Loop Execution
		successCount := 0
			var results [][]string
			
			fmt.Println() // Space before start
			for i, table := range sourceTables {
				// Check for graceful shutdown
				select {
				case <-ctx.Done():
					return ctx.Err()
				default:
				}

				currTargetTable := table + tableSuffix
				if specificTargetTable != "" {
					currTargetTable = specificTargetTable
				}

				fmt.Printf("[%d/%d] Kloning %s.%s -> %s.%s ...\n", i+1, len(sourceTables), sourceDB, table, targetDB, currTargetTable)
				
				_, _, err = svc.CopyTable(ctx, profile, sourceDB, table, targetDB, currTargetTable, copyTableSchemaOnly, copyTableForce, copyTableBackupFirst, copyTableNonInteractive)
				
				status := "Sukses"
				note := "-"
				if err != nil {
					status = "Gagal"
					note = err.Error()
					fmt.Printf("  ❌ Error: %v\n\n", err)
				} else {
					successCount++
					fmt.Printf("  ✅ Berhasil\n\n")
				}
				
				results = append(results, []string{
					fmt.Sprintf("%s.%s", sourceDB, table),
					fmt.Sprintf("%s.%s", targetDB, currTargetTable),
					status,
					note,
				})
			}

			// 6. Final Summary Table
			fmt.Printf("\n--- Ringkasan Copy Tabel ---\n")
			for _, res := range results {
				icon := "✓"
				if res[2] == "Gagal" { icon = "✗" }
				fmt.Printf("%s %-30s -> %-30s [%s]\n", icon, res[0], res[1], res[2])
			}
			
			fmt.Printf("\nSelesai: %d/%d tabel berhasil disalin.\n", successCount, len(sourceTables))
			return nil
		})
	},
}

func init() {
	CmdCopyTable.Flags().StringVarP(&copyTableProfile, "profile", "p", "", "Nama atau path profil database")
	CmdCopyTable.Flags().StringVar(&copyTableProfileKey, "profile-key", "", "Kunci enkripsi profil (jika dienkripsi)")
	CmdCopyTable.Flags().StringVarP(&copyTableTicket, "ticket", "t", "", "Ticket number untuk audit")
	CmdCopyTable.Flags().BoolVar(&copyTableSchemaOnly, "schema-only", false, "Hanya salin struktur (tanpa data)")
	CmdCopyTable.Flags().BoolVar(&copyTableForce, "force", false, "Timpa tabel target jika sudah ada (tanpa backup)")
	CmdCopyTable.Flags().BoolVar(&copyTableBackupFirst, "backup-first", false, "Backup tabel target terlebih dahulu sebelum menimpa")
	CmdCopyTable.Flags().BoolVar(&copyTableNonInteractive, "non-interactive", false, "Jalankan tanpa prompt interaktif")
}
