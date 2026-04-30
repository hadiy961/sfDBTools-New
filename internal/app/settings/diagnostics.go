package settings

import (
	"context"
	"fmt"
	"sfdbtools/internal/shared/connectivity"
	"sfdbtools/internal/shared/database"
	"sfdbtools/internal/shared/systemd"
	"sfdbtools/internal/shared/systeminfo"
	"sfdbtools/internal/ui/print"
	"sfdbtools/internal/ui/prompt"
	"strconv"
	"time"

	"github.com/fatih/color"
)

func (s *Service) RunConnectivityDiagnostics() {
	print.PrintAppHeader("Diagnostics")
	fmt.Println("Checking system health and connectivity...")

	// 1. Internet Check
	isOnline := connectivity.IsInternetAvailable(context.Background(), 2*time.Second)
	internetStatus := color.GreenString("OK (ONLINE)")
	if !isOnline {
		internetStatus = color.RedString("FAILED (OFFLINE)")
	}
	fmt.Printf("- Internet Connection: %s\n", internetStatus)

	// 2. Remote Hub Check
	fmt.Print("- Connecting to Remote Hub...")
	remote, err := database.ConnectToRemoteHub()
	if err != nil {
		fmt.Printf(" %s\n", color.RedString("FAILED (%v)", err))
	} else {
		fmt.Printf(" %s\n", color.GreenString("OK"))
		remote.Close()
	}

	// 3. Local DB Check (Default Profile)
	// For simplicity, we just check if SQLite is okay, or we could try a random profile.
	db, err := database.GetSQLite()
	if err != nil {
		fmt.Printf("- Local SQLite DB: %s\n", color.RedString("FAILED (%v)", err))
	} else {
		fmt.Printf("- Local SQLite DB: %s\n", color.GreenString("OK"))
		_ = db.Ping()
	}

	// 4. Disk Space
	backupDir := database.GetSetting("backup_output_base_directory")
	info, _ := systeminfo.GetSystemInfo(backupDir)
	if info != nil {
		freeGB := float64(info.DiskFree) / 1e9
		totalGB := float64(info.DiskTotal) / 1e9
		status := color.GreenString("OK")
		if freeGB < 1.0 {
			status = color.RedString("LOW")
		}
		fmt.Printf("- Backup Disk: %s (%.2f GB Free / %.2f GB Total)\n", status, freeGB, totalGB)
	}

	// 5. Dependency Check
	fmt.Println("\nDependency Check:")
	versions := systeminfo.GetToolVersions()
	deps := []string{"mysqldump", "gzip"}
	for _, dep := range deps {
		ver, ok := versions[dep]
		if !ok || ver == "not found" {
			fmt.Printf("- %s: %s\n", dep, color.RedString("MISSING"))
		} else {
			fmt.Printf("- %s: %s (%s)\n", dep, color.GreenString("FOUND"), ver)
		}
	}

	prompt.WaitForEnter()
}

// UpdateSystemdTimer handles changes to sync settings.
func (s *Service) UpdateSystemdTimer() {
	syncAuto := database.GetSettingBool("sync_auto")
	intervalStr := database.GetSetting("sync_interval")
	interval, _ := strconv.Atoi(intervalStr)

	m, err := systemd.NewManager()
	if err != nil {
		return
	}

	if syncAuto {
		_ = m.InstallTimer(interval)
		fmt.Println(color.BlueString("[SYSTEMD] Background sync timer updated."))
	} else {
		_ = m.RemoveTimer()
		fmt.Println(color.YellowString("[SYSTEMD] Background sync timer removed."))
	}
}
