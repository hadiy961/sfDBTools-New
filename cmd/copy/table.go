package copycmd

import (
	"context"
	"fmt"
	"os"
	"sfdbtools/internal/app/copy"
	"sfdbtools/internal/cli/deps"
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
	Use:   "table [source_db.table] [target_db.table]",
	Short: "Salin tabel spesifik",
	Run: func(cmd *cobra.Command, args []string) {
		sourceArg := ""
		if len(args) > 0 {
			sourceArg = args[0]
		}
		targetArg := ""
		if len(args) > 1 {
			targetArg = args[1]
		}

		svc := copy.NewService(deps.Deps.Logger, deps.Deps.Config)
		
		// 1. Load Profile
		profile, err := svc.LoadProfile(copyTableProfile, copyTableProfileKey, !copyTableNonInteractive)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		ctx := context.Background()
		var sourceDB, sourceTable string

		// 2. Handle missing source arg
		if sourceArg == "" {
			if copyTableNonInteractive {
				fmt.Fprintf(os.Stderr, "Error: source 'db.table' wajib diisi pada mode non-interaktif\n")
				os.Exit(1)
			}
			sourceDB, sourceTable, err = svc.SelectTableInteractive(ctx, profile)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
		} else {
			parts := strings.Split(sourceArg, ".")
			if len(parts) != 2 {
				fmt.Fprintf(os.Stderr, "Error: format source harus 'database.table'\n")
				os.Exit(1)
			}
			sourceDB, sourceTable = parts[0], parts[1]
		}

		targetDB, targetTable := sourceDB, sourceTable // Default same name
		if targetArg != "" {
			targetParts := strings.Split(targetArg, ".")
			if len(targetParts) == 2 {
				targetDB, targetTable = targetParts[0], targetParts[1]
			} else {
				targetTable = targetArg
			}
		}

		err = svc.CopyTable(ctx, profile, sourceDB, sourceTable, targetDB, targetTable, copyTableSchemaOnly, copyTableForce, copyTableBackupFirst, copyTableNonInteractive)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("\n✓ Tabel '%s.%s' berhasil disalin ke '%s.%s'\n", sourceDB, sourceTable, targetDB, targetTable)
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
