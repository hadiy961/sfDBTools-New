package copycmd

import (
	"context"
	"fmt"
	"sfdbtools/internal/app/copy"
	appdeps "sfdbtools/internal/cli/deps"
	"sfdbtools/internal/cli/runner"
	"sfdbtools/internal/ui/prompt"
	"strings"
	"sync"

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

		// 6. Execution with Worker Pool
		successCount := 0
		var results [][]string
		var mu sync.Mutex
		
		fmt.Println()
		
		// Setup Worker Pool
		type tableTask struct {
			index int
			name  string
		}
		taskChan := make(chan tableTask, len(sourceTables))
		wg := sync.WaitGroup{}
		
		// Start Workers
		numWorkers := copyTableWorkers
		if numWorkers > len(sourceTables) { numWorkers = len(sourceTables) }
		if numWorkers < 1 { numWorkers = 1 }

		for w := 0; w < numWorkers; w++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for task := range taskChan {
					// Check Context
					select {
					case <-ctx.Done():
						return
					default:
					}

					currTargetTable := task.name + tableSuffix
					if specificTargetTable != "" {
						currTargetTable = specificTargetTable
					}

					mu.Lock()
					fmt.Printf("[%d/%d] Kloning %s.%s -> %s.%s ...\n", task.index+1, len(sourceTables), sourceDB, task.name, targetDB, currTargetTable)
					mu.Unlock()

					_, _, verifyStatus, err := svc.CopyTable(ctx, profile, sourceDB, task.name, targetDB, currTargetTable, copyTableSchemaOnly, copyTableForce, copyTableBackupFirst, copyTableIncludeGrants, copyTableVerify, copyTableNonInteractive)
					
					status := "Sukses"
					if err != nil {
						status = "Gagal"
						mu.Lock()
						fmt.Printf("  ❌ Error %s: %v\n\n", task.name, err)
						mu.Unlock()
					} else {
						mu.Lock()
						successCount++
						fmt.Printf("  ✅ %s Berhasil (Checksum: %s)\n\n", task.name, verifyStatus)
						mu.Unlock()
					}

					mu.Lock()
					results = append(results, []string{
						fmt.Sprintf("%s.%s", sourceDB, task.name),
						fmt.Sprintf("%s.%s", targetDB, currTargetTable),
						status,
						verifyStatus,
					})
					mu.Unlock()
					}
			}()
		}

		// Feed Tasks
		for i, tbl := range sourceTables {
			taskChan <- tableTask{index: i, name: tbl}
		}
		close(taskChan)
		wg.Wait()

		// 7. Final Summary Table
		fmt.Printf("\n--- Ringkasan Copy Tabel ---\n")
		fmt.Printf("%-35s -> %-35s [%-6s] [%s]\n", "Sumber", "Tujuan", "Status", "Checksum")
		fmt.Println(strings.Repeat("-", 100))
		for _, res := range results {
			icon := "✓"
			if res[2] == "Gagal" { icon = "✗" }
			fmt.Printf("%s %-33s -> %-35s [%-6s] [%s]\n", icon, res[0], res[1], res[2], res[3])
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
	CmdCopyTable.Flags().IntVarP(&copyTableWorkers, "workers", "w", 4, "Jumlah worker untuk concurrent table copy")
	CmdCopyTable.Flags().BoolVar(&copyTableVerify, "verify", false, "Verifikasi integritas data dengan checksum setelah copy")
	CmdCopyTable.Flags().BoolVar(&copyTableSchemaOnly, "schema-only", false, "Hanya salin struktur (tanpa data)")
	CmdCopyTable.Flags().BoolVar(&copyTableForce, "force", false, "Timpa tabel target jika sudah ada (tanpa backup)")
	CmdCopyTable.Flags().BoolVar(&copyTableBackupFirst, "backup-first", false, "Backup tabel target terlebih dahulu sebelum menimpa")
	CmdCopyTable.Flags().BoolVar(&copyTableIncludeGrants, "include-grants", false, "Salin hak akses user (grants) dari database sumber")
	CmdCopyTable.Flags().BoolVar(&copyTableNonInteractive, "non-interactive", false, "Jalankan tanpa prompt interaktif")
}
