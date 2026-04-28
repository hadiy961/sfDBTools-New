package backupcmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"sfdbtools/internal/app/backup/catalog"
	appdeps "sfdbtools/internal/cli/deps"
	applog "sfdbtools/internal/services/log"
	"sfdbtools/internal/ui/table"

	"github.com/AlecAivazis/survey/v2"
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
		isQuiet, _ := cmd.Root().PersistentFlags().GetBool("quiet")
		noFlags := listDB == "" && listSince == "" && listStatus == "" && listHostname == ""

		if !isQuiet && noFlags {
			fmt.Println("=== Backup Catalog - List ===")
			
			var action string
			survey.AskOne(&survey.Select{
				Message: "Apa yang ingin Anda tampilkan?",
				Options: []string{"Tampilkan Semua", "Gunakan Filter"},
			}, &action)

			if action == "Gunakan Filter" {
				var filters []string
				survey.AskOne(&survey.MultiSelect{
					Message: "Pilih filter yang ingin digunakan (spasi untuk memilih):",
					Options: []string{"Database Name", "Time Range", "Status", "Hostname"},
				}, &filters)

				for _, f := range filters {
					switch f {
					case "Database Name":
						survey.AskOne(&survey.Input{Message: "Masukkan nama database (partial match):"}, &listDB)
					case "Time Range":
						survey.AskOne(&survey.Input{Message: "Masukkan time range (e.g., 24h, 7d):"}, &listSince)
					case "Status":
						survey.AskOne(&survey.Select{
							Message: "Pilih status:",
							Options: []string{"success", "failed", "partial"},
						}, &listStatus)
					case "Hostname":
						survey.AskOne(&survey.Input{Message: "Masukkan hostname:"}, &listHostname)
					}
				}
			}

			survey.AskOne(&survey.Select{
				Message: "Pilih format output:",
				Options: []string{"table", "json"},
				Default: "table",
			}, &listFormat)
		}

		repo := catalog.NewJSONFileRepository(appdeps.Deps.Config.Backup.Catalog.FilePath)
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
				fmt.Println("\nTidak ada backup yang ditemukan di catalog.")
				return
			}

			fmt.Println()
			if listFormat == "json" {
				b, _ := json.MarshalIndent(entries, "", "  ")
				fmt.Println(string(b))
				return
			}

			// Rendering UI
			var rows [][]string
			for i, e := range entries {
				dbName := "N/A"
				if e.BackupType == "all" {
					dbName = "[ALL DATABASES]"
				} else if len(e.DatabaseNames) > 0 {
					dbName = strings.Join(e.DatabaseNames, ", ")
					if len(dbName) > 40 {
						dbName = dbName[:37] + "..."
					}
				}

				ticket := e.Ticket
				if ticket == "" {
					ticket = "N/A"
				}

				rows = append(rows, []string{
					fmt.Sprintf("%d", i+1),
					dbName,
					e.BackupTime.Format("2006-01-02 15:04"),
					e.FileSizeHuman,
					e.BackupStatus,
					e.BackupType,
					fmt.Sprintf("%t", e.Compressed),
					ticket,
					filepath.Base(e.BackupFile),
				})
			}
			headers := []string{"#", "DATABASE", "BACKUP TIME", "SIZE", "STATUS", "TYPE", "COMPRESSED", "TICKET", "FILE NAME"}
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
