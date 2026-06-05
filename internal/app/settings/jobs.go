package settings

import (
	"database/sql"
	"fmt"
	"sfdbtools/internal/shared/database"
	"sfdbtools/internal/ui/prompt"
	"sfdbtools/internal/ui/table"
	"strings"

	"github.com/fatih/color"
)

func (s *Service) ManageJobsMenu(db *sql.DB) {
	for {
		options := []string{
			"1. List All Backup Jobs",
			"2. Add New Backup Job",
			"3. Edit Existing Job",
			"4. Delete Backup Job",
			"0. Back",
		}

		sel, _, err := prompt.SelectOne("Manage Backup Jobs:", options, 0)
		if err != nil || strings.Contains(sel, "Back") {
			return
		}

		switch sel[0:1] {
		case "1":
			s.ListJobs(db)
		case "2":
			s.AddJob(db)
		case "3":
			s.EditJob(db)
		case "4":
			s.DeleteJob(db)
		}
	}
}

func (s *Service) ListJobs(db *sql.DB) {
	rows, err := db.Query("SELECT name, enabled, schedule, mode, profile_name, last_run, last_status FROM backup_jobs")
	if err != nil {
		fmt.Println(color.RedString("Error loading jobs: %v", err))
		return
	}
	defer rows.Close()

	var data [][]string
	for rows.Next() {
		var name, sched, mode, profile, lastRun, lastStatus string
		var enabled int
		rows.Scan(&name, &enabled, &sched, &mode, &profile, &lastRun, &lastStatus)

		status := color.GreenString("ENABLED")
		if enabled == 0 {
			status = color.RedString("DISABLED")
		}

		data = append(data, []string{name, status, sched, mode, profile, lastStatus})
	}

	if len(data) == 0 {
		fmt.Println(color.YellowString("\nNo backup jobs found in database."))
		return
	}

	fmt.Println(color.HiCyanString("\n--- REGISTERED BACKUP JOBS ---"))
	table.Render([]string{"Job Name", "Status", "Schedule", "Mode", "Profile", "Last Status"}, data)
	prompt.WaitForEnter()
}

func (s *Service) AddJob(db *sql.DB) {
	// Implement simple Add Job form if needed, or point to Hub
	fmt.Println(color.YellowString("\nFeature 'Add Job' is best managed from the Central Hub for consistency."))
	fmt.Println("New jobs synced from the Hub will automatically appear here.")
	confirm, _ := prompt.Confirm("Do you still want to add a local job manually?", false)
	if !confirm {
		return
	}
	
	name, _ := prompt.AskText("Job Name (Unique):", nil)
	if name == "" { return }
	
	sched, _ := prompt.AskText("Schedule (Cron format, e.g. 0 2 * * *):", prompt.WithDefault("0 2 * * *"))
	mode, _, _ := prompt.SelectOne("Mode:", []string{"full", "filter", "schema-only"}, 0)
	profile, _ := prompt.AskText("Profile Name:", prompt.WithDefault("default"))
	
	_, err := db.Exec(`INSERT INTO backup_jobs (name, enabled, schedule, mode, profile_name) VALUES (?, 1, ?, ?, ?)`,
		name, sched, mode, profile)
	
	if err != nil {
		fmt.Println(color.RedString("Failed to add job: %v", err))
	} else {
		fmt.Println(color.GreenString("Job added successfully!"))
	}
}

func (s *Service) EditJob(db *sql.DB) {
	// Simple implementation: toggle enabled
	rows, _ := db.Query("SELECT name FROM backup_jobs")
	var names []string
	for rows.Next() {
		var n string
		rows.Scan(&n)
		names = append(names, n)
	}
	rows.Close()

	if len(names) == 0 {
		fmt.Println("No jobs to edit.")
		return
	}

	selected, _, err := prompt.SelectOne("Select Job to Toggle Enabled/Disabled:", names, 0)
	if err != nil { return }

	var current int
	db.QueryRow("SELECT enabled FROM backup_jobs WHERE name = ?", selected).Scan(&current)
	
	newVal := 1
	if current == 1 { newVal = 0 }

	_, err = db.Exec("UPDATE backup_jobs SET enabled = ?, updated_at = CURRENT_TIMESTAMP WHERE name = ?", newVal, selected)
	if err != nil {
		fmt.Println(color.RedString("Update failed: %v", err))
	} else {
		fmt.Println(color.GreenString("Job status updated!"))
	}
}

func (s *Service) DeleteJob(db *sql.DB) {
	rows, _ := db.Query("SELECT name FROM backup_jobs")
	var names []string
	for rows.Next() {
		var n string
		rows.Scan(&n)
		names = append(names, n)
	}
	rows.Close()

	if len(names) == 0 {
		fmt.Println("No jobs to delete.")
		return
	}

	selected, _, err := prompt.SelectOne("Select Job to DELETE:", names, 0)
	if err != nil { return }

	confirm, _ := prompt.Confirm(fmt.Sprintf("Are you sure you want to delete job '%s'?", selected), false)
	if !confirm { return }

	_, err = db.Exec("DELETE FROM backup_jobs WHERE name = ?", selected)
	if err != nil {
		fmt.Println(color.RedString("Delete failed: %v", err))
	} else {
		fmt.Println(color.GreenString("Job deleted successfully!"))
	}
}
