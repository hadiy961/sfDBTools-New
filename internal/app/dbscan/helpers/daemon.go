// File : internal/app/dbscan/helpers/daemon.go
// Deskripsi : Helper untuk daemon/background process scanning
// Author : Hadiyatna Muflihun
// Tanggal : 16 Desember 2025
// Last Modified : 6 Januari 2026
package helpers

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	dbscanmodel "sfdbtools/internal/app/dbscan/model"
	"sfdbtools/internal/services/scheduler"
	"sfdbtools/internal/ui/print"
	"sfdbtools/internal/ui/text"
	"sfdbtools/pkg/consts"
)

func detectUserModeText(mode scheduler.RunMode) string {
	if mode == scheduler.RunModeUser {
		return "user"
	}
	return "system"
}

// SpawnScanDaemon spawns new process sebagai background daemon untuk scanning
func SpawnScanDaemon(config dbscanmodel.ScanEntryConfig) error {
	// Scan ID hanya untuk tampilan
	scanID := fmt.Sprintf("scan_%s", time.Now().Format("20060102_150405"))
	logDir := filepath.Join("logs", "dbscan")
	wd, _ := os.Getwd()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	res, err := scheduler.SpawnSelfInBackground(ctx, scheduler.SpawnSelfOptions{
		UnitPrefix:    "sfdbtools-dbscan",
		Mode:          scheduler.RunModeAuto,
		EnvFile:       "/etc/sfDBTools/.env",
		WorkDir:       wd,
		Collect:       true,
		NoAskPass:     true,
		WrapWithFlock: false,
	})
	if err != nil {
		return err
	}

	print.PrintHeader("DATABASE SCANNING - BACKGROUND MODE")
	print.PrintSuccess("Background process dimulai via systemd")
	print.PrintInfo(fmt.Sprintf("Scan ID: %s", text.ColorText(scanID, consts.UIColorCyan)))
	print.PrintInfo(fmt.Sprintf("Unit: %s", text.ColorText(res.UnitName, consts.UIColorCyan)))
	print.PrintInfo(fmt.Sprintf("Mode: %s", text.ColorText(detectUserModeText(res.Mode), consts.UIColorCyan)))
	print.PrintInfo(fmt.Sprintf("Log dir: %s", text.ColorText(logDir, consts.UIColorCyan)))
	if res.Mode == scheduler.RunModeUser {
		print.PrintInfo(fmt.Sprintf("Status: systemctl --user status %s", res.UnitName))
		print.PrintInfo(fmt.Sprintf("Logs: journalctl --user -u %s -f", res.UnitName))
	} else {
		print.PrintInfo(fmt.Sprintf("Status: sudo systemctl status %s", res.UnitName))
		print.PrintInfo(fmt.Sprintf("Logs: sudo journalctl -u %s -f", res.UnitName))
	}
	print.PrintInfo("Process berjalan di background.")
	_ = res // output disimpan untuk troubleshooting bila diperlukan
	return nil
}
