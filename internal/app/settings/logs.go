package settings

import (
	"database/sql"
	"fmt"
	"sfdbtools/internal/ui/prompt"
	"sfdbtools/internal/ui/table"
	"strings"

	"github.com/fatih/color"
)

func (s *Service) ViewAuditLogs(db *sql.DB) {
	fmt.Println(color.HiCyanString("\n--- LOCAL AUDIT LOGS (Last 20 entries) ---"))

	rows, err := db.Query("SELECT timestamp, event_type, details, is_synced FROM audit_logs ORDER BY timestamp DESC LIMIT 20")
	if err != nil {
		fmt.Println(color.RedString("Error loading logs: %v", err))
		return
	}
	defer rows.Close()

	var data [][]string
	for rows.Next() {
		var ts, event, details string
		var synced int
		rows.Scan(&ts, &event, &details, &synced)

		displaySynced := color.GreenString("YES")
		if synced == 0 {
			displaySynced = color.YellowString("NO")
		}

		data = append(data, []string{ts, event, details, displaySynced})
	}

	if len(data) == 0 {
		fmt.Println(color.YellowString("No audit logs found."))
	} else {
		table.Render([]string{"Timestamp", "Event", "Details", "Synced"}, data)
	}

	prompt.WaitForEnter()
}
