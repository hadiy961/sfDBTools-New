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

		repo := catalog.NewJSONFileRepository("/etc/sfDBTools/catalog.json") // Todo: read from config
		svc := catalog.NewService(repo, applog.NullLogger())

		report, err := svc.GenerateReport(reportPeriod)
		if err != nil {
			fmt.Printf("Error generating report: %v\n", err)
			os.Exit(1)
		}

		if reportFormat != "table" {
			fmt.Printf("[Format %s placeholder - implement rendering here]\n", reportFormat)
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
		for _, db := range report.DatabaseCoverage {
			fmt.Printf("  - %s: %d backups (Last: %s)\n", db.DatabaseName, db.BackupCount, db.LastBackup.Format("2006-01-02 15:04"))
		}
	},
}

func init() {
	CmdBackupReport.Flags().StringVar(&reportPeriod, "period", "weekly", "Report period: daily, weekly, monthly")
	CmdBackupReport.Flags().StringVar(&reportFormat, "format", "table", "Output format: table, json, markdown")
}
