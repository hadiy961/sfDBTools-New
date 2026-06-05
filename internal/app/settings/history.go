package settings

import (
	"database/sql"
	"fmt"
	"sfdbtools/internal/ui/prompt"
	"sfdbtools/internal/ui/table"
	"strings"

	"github.com/fatih/color"
)

func (s *Service) ViewSyncHistory(db *sql.DB) {
	fmt.Println(color.HiCyanString("\n--- SYNC HISTORY (Last 20 entries) ---"))

	rows, err := db.Query("SELECT timestamp, direction, status, changes_count, error_msg FROM sync_history ORDER BY timestamp DESC LIMIT 20")
	if err != nil {
		fmt.Println(color.RedString("Error loading history: %v", err))
		return
	}
	defer rows.Close()

	var data [][]string
	for rows.Next() {
		var ts, dir, status, errMsg string
		var count int
		rows.Scan(&ts, &dir, &status, &count, &errMsg)

		displayStatus := color.GreenString(status)
		if status == "failed" {
			displayStatus = color.RedString(status)
		}

		displayChanges := fmt.Sprintf("%d", count)
		if count > 0 {
			displayChanges = color.YellowString("%d", count)
		}

		data = append(data, []string{ts, dir, displayStatus, displayChanges, errMsg})
	}

	if len(data) == 0 {
		fmt.Println(color.YellowString("No sync history found."))
	} else {
		table.Render([]string{"Timestamp", "Dir", "Status", "Changes", "Error"}, data)
	}

	prompt.WaitForEnter()
}
