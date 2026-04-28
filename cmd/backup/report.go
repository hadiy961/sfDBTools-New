package backupcmd

import (
	"encoding/json"
	"fmt"
	"os"

	"sfdbtools/internal/app/backup/catalog"
	appdeps "sfdbtools/internal/cli/deps"
	applog "sfdbtools/internal/services/log"
	"sfdbtools/internal/ui/table"

	"github.com/AlecAivazis/survey/v2"
	"github.com/dustin/go-humanize"
	"github.com/spf13/cobra"
)

var (
	reportPeriod string
	reportFormat string
)

var CmdBackupReport = &cobra.Command{
	Use:   "report",
	Short: "Menghasilkan laporan statistik backup dari catalog",
	Run: func(cmd *cobra.Command, args []string) {
		isQuiet, _ := cmd.Root().PersistentFlags().GetBool("quiet")
		if !isQuiet {
			fmt.Println("=== Backup Report ===")
			survey.AskOne(&survey.Select{
				Message: "Pilih periode report:",
				Options: []string{"daily", "weekly", "monthly", "all"},
				Default: "weekly",
			}, &reportPeriod)

			survey.AskOne(&survey.Select{
				Message: "Pilih format output:",
				Options: []string{"table", "json", "markdown"},
				Default: "table",
			}, &reportFormat)
			
			fmt.Println("\n📊 Generating report...")
		}

		repo := catalog.NewJSONFileRepository(appdeps.Deps.Config.Backup.Catalog.FilePath)
		svc := catalog.NewService(repo, applog.NullLogger())

		report, err := svc.GenerateReport(reportPeriod)
		if err != nil {
			fmt.Printf("Error generating report: %v\n", err)
			os.Exit(1)
		}

		if reportFormat == "json" {
			b, _ := json.MarshalIndent(report, "", "  ")
			fmt.Println(string(b))
			return
		}

		if reportFormat == "markdown" {
			fmt.Printf("## Backup Report — %s\n\n", report.Period)
			fmt.Println("### Summary")
			fmt.Printf("- **Total Backups:** %d\n", report.TotalBackups)
			fmt.Printf("- **Total Size:** %s\n", report.TotalSizeHuman)
			fmt.Printf("- **Success Rate:** %.1f%%\n", report.SuccessRate)
			fmt.Printf("- **Failed:** %d\n\n", report.FailedCount)
			
			fmt.Println("### Database Coverage")
			fmt.Println("| Database | Backup Count | Total Size | Last Backup |")
			fmt.Println("|---|---|---|---|")
			for _, db := range report.DatabaseCoverage {
				fmt.Printf("| %s | %d | %s | %s |\n", db.DatabaseName, db.BackupCount, humanize.Bytes(uint64(db.TotalSize)), db.LastBackup.Format("2006-01-02 15:04"))
			}
			return
		}

		fmt.Printf("📊 Backup Report — %s\n", report.Period)
		fmt.Println("────────────────────────────────────────────────────────────────────────────")
		fmt.Printf("  Total Backups : %d\n", report.TotalBackups)
		fmt.Printf("  Total Size    : %s\n", report.TotalSizeHuman)
		fmt.Printf("  Success Rate  : %.1f%%\n", report.SuccessRate)
		fmt.Printf("  Failed        : %d\n", report.FailedCount)
		fmt.Println()
		
		fmt.Println("Database Coverage:")
		var rows [][]string
		for _, db := range report.DatabaseCoverage {
			rows = append(rows, []string{
				db.DatabaseName,
				db.LastBackup.Format("2006-01-02 15:04"),
				fmt.Sprintf("%d", db.BackupCount),
				humanize.Bytes(uint64(db.TotalSize)),
			})
		}
		headers := []string{"DATABASE", "LAST BACKUP", "COUNT", "SIZE"}
		table.Render(headers, rows)
	},
}

func init() {
	CmdBackupReport.Flags().StringVar(&reportPeriod, "period", "weekly", "Report period: daily, weekly, monthly")
	CmdBackupReport.Flags().StringVar(&reportFormat, "format", "table", "Output format: table, json, markdown")
}
