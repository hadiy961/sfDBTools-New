package copycmd

import (
	"context"
	"fmt"
	"os"
	"sfdbtools/internal/app/copy"
	"sfdbtools/internal/cli/deps"
	"sfdbtools/internal/ui/prompt"
	"time"

	"github.com/spf13/cobra"
)

var (
	copyDBProfile        string
	copyDBProfileKey     string
	copyDBSchemaOnly     bool
	copyDBUseDisk        bool
	copyDBForce          bool
	copyDBBackupFirst    bool
	copyDBNonInteractive bool
)

// CmdCopyDB menyalin database utuh
var CmdCopyDB = &cobra.Command{
	Use:   "db [source_db...] [target_db]",
	Short: "Salin database utuh",
	Run: func(cmd *cobra.Command, args []string) {
		var sourceDBs []string
		targetDB := ""

		if len(args) > 1 {
			// Jika > 1 arg, maka arg terakhir dianggap target, sisanya source
			sourceDBs = args[:len(args)-1]
			targetDB = args[len(args)-1]
		} else if len(args) == 1 {
			sourceDBs = []string{args[0]}
		}

		svc := copy.NewService(deps.Deps.Logger, deps.Deps.Config)
		
		// 1. Load Profile
		profile, err := svc.LoadProfile(copyDBProfile, copyDBProfileKey, !copyDBNonInteractive)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\\n", err)
			os.Exit(1)
		}

		ctx := context.Background()

		// 2. Handle missing source databases (Interactive)
		if len(sourceDBs) == 0 {
			if copyDBNonInteractive {
				fmt.Fprintf(os.Stderr, "Error: source database wajib diisi pada mode non-interaktif\\n")
				os.Exit(1)
			}
			sourceDBs, err = svc.SelectDatabasesInteractive(ctx, profile)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\\n", err)
				os.Exit(1)
			}
		}

		if len(sourceDBs) == 0 {
			fmt.Println("Tidak ada database yang dipilih.")
			return
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

		// 4. Loop Execution
		successCount := 0
		var results [][]string
		
		fmt.Println() // Space before start
		for i, db := range sourceDBs {
			currTarget := targetDB
			if len(sourceDBs) > 1 {
				currTarget = db + suffix
			}

			fmt.Printf("[%d/%d] Kloning %s -> %s ...\n", i+1, len(sourceDBs), db, currTarget)
			finalTarget, err := svc.CopyDatabase(ctx, profile, db, currTarget, copyDBSchemaOnly, copyDBUseDisk, copyDBForce, copyDBBackupFirst, copyDBNonInteractive)
			
			status := "Sukses"
			note := "-"
			if err != nil {
				status = "Gagal"
				note = err.Error()
				fmt.Printf("  ❌ Error: %v\n\n", err)
			} else {
				successCount++
				fmt.Printf("  ✅ Berhasil: %s\n\n", finalTarget)
			}
			
			results = append(results, []string{db, currTarget, status, note})
		}

		// 5. Final Summary
		fmt.Printf("\n--- Ringkasan Copy Database ---\n")
		for _, res := range results {
			icon := "✓"
			if res[2] == "Gagal" { icon = "✗" }
			fmt.Printf("%s %-30s -> %-30s [%s]\n", icon, res[0], res[1], res[2])
		}

		fmt.Printf("\nSelesai: %d/%d database berhasil disalin.\n", successCount, len(sourceDBs))
	},
}

func init() {
	CmdCopyDB.Flags().StringVarP(&copyDBProfile, "profile", "p", "", "Nama atau path profil database")
	CmdCopyDB.Flags().StringVar(&copyDBProfileKey, "profile-key", "", "Kunci enkripsi profil (jika dienkripsi)")
	CmdCopyDB.Flags().BoolVar(&copyDBSchemaOnly, "schema-only", false, "Hanya salin struktur (tanpa data)")
	CmdCopyDB.Flags().BoolVar(&copyDBUseDisk, "use-disk", false, "Gunakan media disk (dump & restore) alih-alih streaming RAM")
	CmdCopyDB.Flags().BoolVar(&copyDBForce, "force", false, "Timpa database target jika sudah ada (tanpa backup)")
	CmdCopyDB.Flags().BoolVar(&copyDBBackupFirst, "backup-first", false, "Backup database target terlebih dahulu sebelum menimpa")
	CmdCopyDB.Flags().BoolVar(&copyDBNonInteractive, "non-interactive", false, "Jalankan tanpa prompt interaktif")
}
