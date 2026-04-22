package copycmd

import (
	"context"
	"fmt"
	"os"
	"sfdbtools/internal/app/copy"
	"sfdbtools/internal/cli/deps"

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
	Use:   "db [source_db] [target_db]",
	Short: "Salin database utuh",
	Run: func(cmd *cobra.Command, args []string) {
		sourceDB := ""
		if len(args) > 0 {
			sourceDB = args[0]
		}
		targetDB := ""
		if len(args) > 1 {
			targetDB = args[1]
		}

		svc := copy.NewService(deps.Deps.Logger, deps.Deps.Config)
		
		// 1. Load Profile (Interactive picker if empty)
		profile, err := svc.LoadProfile(copyDBProfile, copyDBProfileKey, !copyDBNonInteractive)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		ctx := context.Background()

		// 2. Handle missing source database in interactive mode
		if sourceDB == "" {
			if copyDBNonInteractive {
				fmt.Fprintf(os.Stderr, "Error: source database wajib diisi pada mode non-interaktif\n")
				os.Exit(1)
			}
			
			// Ambil list database untuk picker
			sourceDB, err = svc.SelectDatabaseInteractive(ctx, profile)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
		}

		targetDB, err = svc.CopyDatabase(ctx, profile, sourceDB, targetDB, copyDBSchemaOnly, copyDBUseDisk, copyDBForce, copyDBBackupFirst, copyDBNonInteractive)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("\n✓ Database '%s' berhasil disalin ke '%s'\n", sourceDB, targetDB)
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
