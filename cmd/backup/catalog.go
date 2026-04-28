package backupcmd

import (
	"fmt"
	"os"

	"sfdbtools/internal/app/backup/catalog"
	applog "sfdbtools/internal/services/log"

	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/cobra"
)

var (
	catalogDir     string
	catalogForce   bool
	catalogRecurse bool
)

var CmdBackupCatalog = &cobra.Command{
	Use:   "catalog",
	Short: "Manajemen backup catalog",
}

var catalogRebuildCmd = &cobra.Command{
	Use:   "rebuild",
	Short: "Membangun ulang catalog dari direktori yang ditentukan",
	Run: func(cmd *cobra.Command, args []string) {
		isQuiet, _ := cmd.Root().PersistentFlags().GetBool("quiet")
		if !isQuiet && catalogDir == "" {
			fmt.Println("=== Catalog Rebuild ===")
			err := survey.AskOne(&survey.Input{
				Message: "Masukkan directory untuk scan:",
				Default: ".",
			}, &catalogDir)
			if err != nil || catalogDir == "" {
				fmt.Println("Dibatalkan.")
				return
			}
			
			if !catalogForce {
				var confirm bool
				survey.AskOne(&survey.Confirm{
					Message: "Ini akan memindai dan me-register ulang metadata. Lanjutkan?",
					Default: true,
				}, &confirm)
				if !confirm {
					fmt.Println("Dibatalkan.")
					return
				}
			}
		}

		if catalogDir == "" {
			fmt.Println("Error: --dir harus diisi untuk rebuild catalog")
			os.Exit(1)
		}

		repo := catalog.NewJSONFileRepository("/etc/sfDBTools/catalog.json") // Todo: read from config
		svc := catalog.NewService(repo, applog.NullLogger())

		fmt.Printf("\n🔍 Scanning directory: %s\n", catalogDir)
		count, err := svc.RebuildFromDirectory(catalogDir)
		if err != nil {
			fmt.Printf("Error rebuilding catalog: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("✓ Catalog rebuild complete. Registered %d entries.\n", count)
	},
}

var catalogPruneCmd = &cobra.Command{
	Use:   "prune",
	Short: "Menghapus entry di catalog yang file backupnya sudah tidak ada di disk",
	Run: func(cmd *cobra.Command, args []string) {
		repo := catalog.NewJSONFileRepository("/etc/sfDBTools/catalog.json") // Todo: read from config
		svc := catalog.NewService(repo, applog.NullLogger())

		isQuiet, _ := cmd.Root().PersistentFlags().GetBool("quiet")
		if !isQuiet {
			fmt.Println("=== Catalog Prune ===")
			var confirm bool
			survey.AskOne(&survey.Confirm{
				Message: "Hapus entry catalog yang filenya sudah tidak ada di disk?",
				Default: true,
			}, &confirm)
			if !confirm {
				fmt.Println("Dibatalkan.")
				return
			}
		}

		fmt.Println("\n🔍 Checking catalog entries for missing files...")
		removed, err := svc.Prune()
		if err != nil {
			fmt.Printf("Error pruning catalog: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("✓ Pruned %d entries.\n", removed)
	},
}

func init() {
	CmdBackupCatalog.AddCommand(catalogRebuildCmd)
	CmdBackupCatalog.AddCommand(catalogPruneCmd)

	catalogRebuildCmd.Flags().StringVar(&catalogDir, "dir", "", "Directory to scan")
	catalogRebuildCmd.Flags().BoolVar(&catalogForce, "force", false, "Skip confirmation prompt")
	catalogRebuildCmd.Flags().BoolVar(&catalogRecurse, "recursive", true, "Scan subdirectories")
}
