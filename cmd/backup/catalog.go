package backupcmd

import (
	"encoding/json"
	"fmt"
	"os"

	"sfdbtools/internal/app/backup/catalog"
	appdeps "sfdbtools/internal/cli/deps"
	applog "sfdbtools/internal/services/log"
	"sfdbtools/internal/ui/input"
	"sfdbtools/internal/ui/table"

	"github.com/AlecAivazis/survey/v2"
	"github.com/dustin/go-humanize"
	"github.com/spf13/cobra"
)

var (
	catalogDir     string
	catalogForce   bool
	catalogRecurse bool
	catalogFormat  string
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
			
			dir, err := input.SelectDirectoryInteractive(".", "Pilih directory untuk scan:")
			if err != nil || dir == "" {
				fmt.Println("Dibatalkan.")
				return
			}
			catalogDir = dir
			
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

		repo := catalog.NewJSONFileRepository(appdeps.Deps.Config.Backup.Catalog.FilePath)
		svc := catalog.NewService(repo, applog.NullLogger())

		fmt.Printf("\n🔍 Scanning directory: %s (recursive: %t)\n", catalogDir, catalogRecurse)
		count, err := svc.RebuildFromDirectory(catalogDir, catalogRecurse)
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
		repo := catalog.NewJSONFileRepository(appdeps.Deps.Config.Backup.Catalog.FilePath)
		svc := catalog.NewService(repo, applog.NullLogger())

		fmt.Println("\n🔍 Checking catalog entries for missing files...")
		prunable, err := svc.GetPrunableEntries()
		if err != nil {
			fmt.Printf("Error checking catalog: %v\n", err)
			os.Exit(1)
		}

		if len(prunable) == 0 {
			fmt.Println("✓ No prunable entries found.")
			return
		}

		fmt.Printf("Found %d entries with missing files:\n", len(prunable))
		var rows [][]string
		for i, e := range prunable {
			dbName := "N/A"
			if len(e.DatabaseNames) > 0 {
				dbName = e.DatabaseNames[0]
			}
			rows = append(rows, []string{
				fmt.Sprintf("%d", i+1),
				dbName,
				e.BackupTime.Format("2006-01-02 15:04"),
				e.BackupFile,
			})
		}
		table.Render([]string{"#", "DATABASE", "BACKUP TIME", "FILE PATH"}, rows)

		isQuiet, _ := cmd.Root().PersistentFlags().GetBool("quiet")
		if !isQuiet {
			var confirm bool
			survey.AskOne(&survey.Confirm{
				Message: fmt.Sprintf("Hapus %d entry catalog ini?", len(prunable)),
				Default: true,
			}, &confirm)
			if !confirm {
				fmt.Println("Dibatalkan.")
				return
			}
		}

		var ids []string
		for _, e := range prunable {
			ids = append(ids, e.ID)
		}

		removed, err := svc.DeleteEntries(ids)
		if err != nil {
			fmt.Printf("Error pruning catalog: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("✓ Pruned %d entries.\n", removed)
	},
}

var catalogStatsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Menampilkan statistik dari backup catalog",
	Run: func(cmd *cobra.Command, args []string) {
		repo := catalog.NewJSONFileRepository(appdeps.Deps.Config.Backup.Catalog.FilePath)
		svc := catalog.NewService(repo, applog.NullLogger())

		stats, err := svc.GetStats()
		if err != nil {
			fmt.Printf("Error fetching stats: %v\n", err)
			os.Exit(1)
		}

		if catalogFormat == "json" {
			b, _ := json.MarshalIndent(stats, "", "  ")
			fmt.Println(string(b))
			return
		}

		fmt.Println("\n📋 Catalog Statistics")
		
		headers := []string{"STATISTIC", "VALUE"}
		rows := [][]string{
			{"Total Entries", fmt.Sprintf("%d", stats.TotalEntries)},
			{"Total Size", stats.TotalSizeHuman},
		}

		if stats.TotalEntries > 0 {
			rows = append(rows, []string{"Oldest Backup", stats.OldestBackup.Format("2006-01-02 15:04")})
			rows = append(rows, []string{"Newest Backup", stats.NewestBackup.Format("2006-01-02 15:04")})
		} else {
			rows = append(rows, []string{"Oldest Backup", "N/A"})
			rows = append(rows, []string{"Newest Backup", "N/A"})
		}

		rows = append(rows, []string{"Unique Databases", fmt.Sprintf("%d", stats.UniqueDatabases)})
		rows = append(rows, []string{"Success Rate", fmt.Sprintf("%.1f%%", stats.SuccessRate)})
		rows = append(rows, []string{"Catalog File", stats.CatalogFile})
		rows = append(rows, []string{"Catalog Size", humanize.Bytes(uint64(stats.CatalogSize))})

		table.Render(headers, rows)
	},
}

func init() {
	CmdBackupCatalog.AddCommand(catalogRebuildCmd)
	CmdBackupCatalog.AddCommand(catalogPruneCmd)
	CmdBackupCatalog.AddCommand(catalogStatsCmd)

	catalogRebuildCmd.Flags().StringVar(&catalogDir, "dir", "", "Directory to scan")
	catalogRebuildCmd.Flags().BoolVar(&catalogForce, "force", false, "Skip confirmation prompt")
	catalogRebuildCmd.Flags().BoolVar(&catalogRecurse, "recursive", true, "Scan subdirectories")

	catalogStatsCmd.Flags().StringVar(&catalogFormat, "format", "table", "Output format (table/json)")
}
