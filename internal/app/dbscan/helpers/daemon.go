// File : internal/app/dbscan/helpers/daemon.go
// Deskripsi : Helper untuk daemon/background process scanning
// Author : Hadiyatna Muflihun
// Tanggal : 16 Desember 2025
// Last Modified : 2026-01-05
package helpers

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	dbscanmodel "sfDBTools/internal/app/dbscan/model"
	schedulerutil "sfDBTools/internal/services/scheduler"
	"sfDBTools/pkg/consts"
	"sfDBTools/pkg/ui"
)

func detectUserModeText(mode schedulerutil.RunMode) string {
	if mode == schedulerutil.RunModeUser {
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
	res, err := schedulerutil.SpawnSelfInBackground(ctx, schedulerutil.SpawnSelfOptions{
		UnitPrefix:    "sfdbtools-dbscan",
		Mode:          schedulerutil.RunModeAuto,
		EnvFile:       "/etc/sfDBTools/.env",
		WorkDir:       wd,
		Collect:       true,
		NoAskPass:     true,
		WrapWithFlock: false,
	})
	if err != nil {
		return err
	}

	ui.PrintHeader("DATABASE SCANNING - BACKGROUND MODE")
	ui.PrintSuccess("Background process dimulai via systemd")
	ui.PrintInfo(fmt.Sprintf("Scan ID: %s", ui.ColorText(scanID, consts.UIColorCyan)))
	ui.PrintInfo(fmt.Sprintf("Unit: %s", ui.ColorText(res.UnitName, consts.UIColorCyan)))
	ui.PrintInfo(fmt.Sprintf("Mode: %s", ui.ColorText(detectUserModeText(res.Mode), consts.UIColorCyan)))
	ui.PrintInfo(fmt.Sprintf("Log dir: %s", ui.ColorText(logDir, consts.UIColorCyan)))
	if res.Mode == schedulerutil.RunModeUser {
		ui.PrintInfo(fmt.Sprintf("Status: systemctl --user status %s", res.UnitName))
		ui.PrintInfo(fmt.Sprintf("Logs: journalctl --user -u %s -f", res.UnitName))
	} else {
		ui.PrintInfo(fmt.Sprintf("Status: sudo systemctl status %s", res.UnitName))
		ui.PrintInfo(fmt.Sprintf("Logs: sudo journalctl -u %s -f", res.UnitName))
	}
	ui.PrintInfo("Process berjalan di background.")
	_ = res // output disimpan untuk troubleshooting bila diperlukan
	return nil
}
