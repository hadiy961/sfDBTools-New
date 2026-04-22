package copycmd

import (
	"context"
	"fmt"
	"os"
	"sfdbtools/internal/app/copy"
	"sfdbtools/internal/cli/deps"
	"sfdbtools/internal/ui/prompt"
	"strings"

	"github.com/spf13/cobra"
)

var (
	copyTableProfile        string
	copyTableProfileKey     string
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
		var sourceDB string
		var sourceTables []string
		targetArg := ""

		if len(args) > 1 {
			// Multi arg: last one is target (if no dot) or another source
			// Simplification for CLI: use explicit arg handling
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

		svc := copy.NewService(deps.Deps.Logger, deps.Deps.Config)
		
		// 1. Load Profile
		profile, err := svc.LoadProfile(copyTableProfile, copyTableProfileKey, !copyTableNonInteractive)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\\n", err)
			os.Exit(1)
		}

		ctx := context.Background()

		// 2. Handle missing source arg (Interactive)
		if len(sourceTables) == 0 {
			if copyTableNonInteractive {
				fmt.Fprintf(os.Stderr, "Error: source 'db.table' wajib diisi pada mode non-interaktif\\n")
				os.Exit(1)
			}
			sourceDB, sourceTables, err = svc.SelectTablesInteractive(ctx, profile)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\\n", err)
				os.Exit(1)
			}
		}

		if len(sourceTables) == 0 {
			fmt.Println("Tidak ada tabel yang dipilih.")
			return
		}

		// 3. Handle Target Database
		targetDB := sourceDB
		if targetArg != "" && !strings.Contains(targetArg, ".") {
			targetDB = targetArg
		} else if !copyTableNonInteractive {
			targetDB, _ = prompt.AskText("Masukkan database tujuan:", prompt.WithDefault(sourceDB))
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
		for _, table := range sourceTables {
			currTargetTable := table + tableSuffix
			if specificTargetTable != "" {
				currTargetTable = specificTargetTable
			}

			fmt.Printf("\\n[Processing %d/%d]: %s.%s -> %s.%s\\n", successCount+1, len(sourceTables), sourceDB, table, targetDB, currTargetTable)
			_, _, err = svc.CopyTable(ctx, profile, sourceDB, table, targetDB, currTargetTable, copyTableSchemaOnly, copyTableForce, copyTableBackupFirst, copyTableNonInteractive)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error copy tabel %s: %v\\n", table, err)
				continue
			}
			successCount++
			fmt.Printf("✓ Berhasil: %s.%s\\n", targetDB, currTargetTable)
		}

		fmt.Printf("\\nSelesai: %d/%d tabel berhasil disalin.\\n", successCount, len(sourceTables))
	},
}

func init() {
	CmdCopyTable.Flags().StringVarP(&copyTableProfile, "profile", "p", "", "Nama atau path profil database")
	CmdCopyTable.Flags().StringVar(&copyTableProfileKey, "profile-key", "", "Kunci enkripsi profil (jika dienkripsi)")
	CmdCopyTable.Flags().BoolVar(&copyTableSchemaOnly, "schema-only", false, "Hanya salin struktur (tanpa data)")
	CmdCopyTable.Flags().BoolVar(&copyTableForce, "force", false, "Timpa tabel target jika sudah ada (tanpa backup)")
	CmdCopyTable.Flags().BoolVar(&copyTableBackupFirst, "backup-first", false, "Backup tabel target terlebih dahulu sebelum menimpa")
	CmdCopyTable.Flags().BoolVar(&copyTableNonInteractive, "non-interactive", false, "Jalankan tanpa prompt interaktif")
}
