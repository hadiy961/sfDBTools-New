package backupcmd

import (
	"fmt"
	"os"

	"sfdbtools/internal/app/backup/catalog"
	applog "sfdbtools/internal/services/log"
	"sfdbtools/internal/ui/table"

	"github.com/spf13/cobra"
)

var (
	listDB       string
	listSince    string
	listStatus   string
	listHostname string
	listFormat   string
	listLimit    int
)

var CmdBackupList = &cobra.Command{
	Use:   "list",
	Short: "Menampilkan daftar backup dari catalog",
	Run: func(cmd *cobra.Command, args []string) {
		repo := catalog.NewJSONFileRepository("/etc/sfDBTools/catalog.json") // Todo: load from config
		svc := catalog.NewService(repo, applog.NullLogger())

		opts := catalog.QueryOptions{
			Database: listDB,
			Since:    listSince,
			Status:   listStatus,
			Hostname: listHostname,
			Limit:    listLimit,
		}

		entries, err := svc.Query(opts)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

		if len(entries) == 0 {
			fmt.Println("Tidak ada backup yang ditemukan di catalog.")
			return
		}

		// Rendering UI
		var rows [][]string
		for i, e := range entries {
			rows = append(rows, []string{
				fmt.Sprintf("%d", i+1),
				e.DatabaseNames[0], // Simplified
				e.BackupTime.Format("2006-01-02 15:04"),
				e.FileSizeHuman,
				e.BackupStatus,
				e.BackupType,
				fmt.Sprintf("%t", e.Compressed),
				e.ChecksumHash[:8], // Shortened for space
			})
		}
		headers := []string{"#", "DATABASE", "BACKUP TIME", "SIZE", "STATUS", "TYPE", "COMPRESSED", "CHECKSUM"}
		table.Render(headers, rows)
	},
}

func init() {
	CmdBackupList.Flags().StringVar(&listDB, "db", "", "Filter by database name (substring)")
	CmdBackupList.Flags().StringVar(&listSince, "since", "", "Filter by time (e.g. 24h, 7d)")
	CmdBackupList.Flags().StringVar(&listStatus, "status", "", "Filter by status (success, failed, partial)")
	CmdBackupList.Flags().StringVar(&listHostname, "host", "", "Filter by hostname")
	CmdBackupList.Flags().StringVar(&listFormat, "format", "table", "Output format")
	CmdBackupList.Flags().IntVar(&listLimit, "limit", 50, "Limit number of entries to show")
}
