// File : internal/restore/command.go
// Deskripsi : Command execution functions untuk cmd layer
// Author : Hadiyatna Muflihun
// Tanggal : 2025-12-16
// Last Modified : 2025-12-17

package restore

import (
	"context"
	"os"
	"os/signal"
	"sfDBTools/internal/parsing"
	"sfDBTools/internal/restore/display"
	"sfDBTools/internal/types"
	"sfDBTools/pkg/ui"
	"syscall"

	"github.com/spf13/cobra"
)

// ExecuteRestoreSingleCommand adalah entry point untuk restore single command
func ExecuteRestoreSingleCommand(cmd *cobra.Command, deps *types.Dependencies) error {
	logger := deps.Logger
	logger.Info("Memulai proses restore single database")

	// Parse options dari command flags
	parsedOpts, err := parsing.ParsingRestoreSingleOptions(cmd)
	if err != nil {
		logger.Error("gagal parsing opsi: " + err.Error())
		return err
	}

	// Inisialisasi service restore
	svc := NewRestoreService(logger, deps.Config, &parsedOpts)

	// Setup signal handling untuk graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		sig := <-sigChan
		logger.Warnf("Menerima signal %v, menghentikan restore... (Tekan sekali lagi untuk force exit)", sig)
		svc.HandleShutdown()
		cancel()

		<-sigChan
		logger.Warn("Menerima signal kedua, memaksa berhenti (force exit)...")
		os.Exit(1)
	}()

	// Setup restore session (koneksi database, validasi, dll)
	if err := svc.SetupRestoreSession(ctx); err != nil {
		return err
	}
	defer svc.Close()

	// Execute restore
	result, err := svc.ExecuteRestoreSingle(ctx)
	if err != nil {
		if ctx.Err() != nil {
			logger.Warn("Proses restore dibatalkan.")
			return nil
		}
		logger.Error("Restore gagal: " + err.Error())
		svc.ErrorLog.Log(map[string]interface{}{
			"function": "ExecuteRestoreSingle",
			"error":    err.Error(),
		}, err)
		return err
	}

	// Display result
	display.ShowRestoreSingleResult(result)

	ui.PrintSuccess("Restore database berhasil diselesaikan")
	logger.Info("Restore database berhasil diselesaikan")

	return nil
}

// ExecuteRestorePrimaryCommand adalah entry point untuk restore primary command
func ExecuteRestorePrimaryCommand(cmd *cobra.Command, deps *types.Dependencies) error {
	logger := deps.Logger
	logger.Info("Memulai proses restore primary database")

	// Parse options dari command flags
	parsedOpts, err := parsing.ParsingRestorePrimaryOptions(cmd)
	if err != nil {
		logger.Error("gagal parsing opsi: " + err.Error())
		return err
	}

	// Inisialisasi service restore primary
	svc := NewRestorePrimaryService(logger, deps.Config, &parsedOpts)

	// Setup signal handling untuk graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		sig := <-sigChan
		logger.Warnf("Menerima signal %v, menghentikan restore... (Tekan sekali lagi untuk force exit)", sig)
		svc.HandleShutdown()
		cancel()

		<-sigChan
		logger.Warn("Menerima signal kedua, memaksa berhenti (force exit)...")
		os.Exit(1)
	}()

	// Setup restore session (koneksi database, validasi, dll)
	if err := svc.SetupRestorePrimarySession(ctx); err != nil {
		return err
	}
	defer svc.Close()

	// Execute restore primary
	result, err := svc.ExecuteRestorePrimary(ctx)
	if err != nil {
		if ctx.Err() != nil {
			logger.Warn("Proses restore primary dibatalkan.")
			return nil
		}
		logger.Error("Restore primary gagal: " + err.Error())
		svc.ErrorLog.Log(map[string]interface{}{
			"function": "ExecuteRestorePrimary",
			"error":    err.Error(),
		}, err)
		return err
	}

	// Display result
	display.ShowRestorePrimaryResult(result)

	ui.PrintSuccess("Restore primary database berhasil diselesaikan")
	logger.Info("Restore primary database berhasil diselesaikan")

	return nil
}

// ExecuteRestoreAllCommand adalah entry point untuk restore all databases command
func ExecuteRestoreAllCommand(cmd *cobra.Command, deps *types.Dependencies) error {
	logger := deps.Logger
	logger.Info("Memulai proses restore all databases")

	// Parse options dari command flags
	parsedOpts, err := parsing.ParsingRestoreAllOptions(cmd)
	if err != nil {
		logger.Error("gagal parsing opsi: " + err.Error())
		return err
	}

	// Inisialisasi service restore all
	svc := NewRestoreAllService(logger, deps.Config, &parsedOpts)

	// Setup signal handling untuk graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		sig := <-sigChan
		logger.Warnf("Menerima signal %v, menghentikan restore... (Tekan sekali lagi untuk force exit)", sig)
		svc.HandleShutdown()
		cancel()

		<-sigChan
		logger.Warn("Menerima signal kedua, memaksa berhenti (force exit)...")
		os.Exit(1)
	}()

	// Setup restore session (koneksi database, validasi, dll)
	if err := svc.SetupRestoreAllSession(ctx); err != nil {
		return err
	}
	defer svc.Close()

	// Execute restore all
	result, err := svc.ExecuteRestoreAll(ctx)
	if err != nil {
		if ctx.Err() != nil {
			logger.Warn("Proses restore all dibatalkan.")
			return nil
		}
		logger.Error("Restore all gagal: " + err.Error())
		svc.ErrorLog.Log(map[string]interface{}{
			"function": "ExecuteRestoreAll",
			"error":    err.Error(),
		}, err)
		return err
	}

	// Display result
	display.ShowRestoreAllResult(result)

	ui.PrintSuccess("Restore all databases berhasil diselesaikan")
	logger.Info("Restore all databases berhasil diselesaikan")

	return nil
}
