// File : internal/restore/command_helpers.go
// Deskripsi : Helper untuk command layer restore (signal handling + lifecycle)
// Author : Hadiyatna Muflihun
// Tanggal : 30 Desember 2025
// Last Modified : 5 Januari 2026
package restore

import (
	"context"
	"os"
	"os/signal"
	appdeps "sfDBTools/internal/cli/deps"
	applog "sfDBTools/internal/services/log"
	"sfDBTools/internal/ui/print"
	"syscall"

	restoremodel "sfDBTools/internal/app/restore/model"

	"github.com/spf13/cobra"
)

type restoreSetupFunc func(ctx context.Context) error
type restoreExecFunc func(ctx context.Context) (*restoremodel.RestoreResult, error)

type restoreParseFunc func(cmd *cobra.Command) (interface{}, error)

type restoreSetupMethod func(svc *Service, ctx context.Context) error
type restoreExecMethod func(svc *Service, ctx context.Context) (*restoremodel.RestoreResult, error)
type restoreShowFunc func(result *restoremodel.RestoreResult)

func runRestoreWithLifecycle(
	logger applog.Logger,
	svc *Service,
	setup restoreSetupFunc,
	exec restoreExecFunc,
	cancelMsg string,
	errMsgPrefix string,
	errFunction string,
) (*restoremodel.RestoreResult, error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	finished := make(chan struct{})
	sigChan := make(chan os.Signal, 2)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(sigChan)
	defer close(finished)

	go func() {
		select {
		case sig := <-sigChan:
			logger.Warnf("Menerima signal %v, menghentikan restore... (Tekan sekali lagi untuk force exit)", sig)
			svc.HandleShutdown()
			cancel()
		case <-finished:
			return
		}

		select {
		case <-sigChan:
			logger.Warn("Menerima signal kedua, memaksa berhenti (force exit)...")
			os.Exit(1)
		case <-finished:
			return
		}
	}()

	if err := setup(ctx); err != nil {
		return nil, err
	}
	defer svc.Close()

	result, err := exec(ctx)
	if err != nil {
		if ctx.Err() != nil {
			logger.Warn(cancelMsg)
			return nil, nil
		}
		logger.Error(errMsgPrefix + err.Error())
		svc.ErrorLog.Log(map[string]interface{}{
			"function": errFunction,
			"error":    err.Error(),
		}, err)
		return nil, err
	}

	return result, nil
}

func executeRestoreCommand(
	cmd *cobra.Command,
	deps *appdeps.Dependencies,
	startMsg string,
	parse restoreParseFunc,
	setup restoreSetupMethod,
	exec restoreExecMethod,
	show restoreShowFunc,
	cancelMsg string,
	errMsgPrefix string,
	errFunction string,
	successMsg string,
) error {
	logger := deps.Logger
	logger.Info(startMsg)

	parsedOpts, err := parse(cmd)
	if err != nil {
		logger.Error("gagal parsing opsi: " + err.Error())
		return err
	}

	svc := NewRestoreService(logger, deps.Config, parsedOpts)

	result, err := runRestoreWithLifecycle(
		logger,
		svc,
		func(ctx context.Context) error {
			return setup(svc, ctx)
		},
		func(ctx context.Context) (*restoremodel.RestoreResult, error) {
			return exec(svc, ctx)
		},
		cancelMsg,
		errMsgPrefix,
		errFunction,
	)
	if err != nil {
		return err
	}
	if result == nil {
		return nil
	}

	show(result)

	print.PrintSuccess(successMsg)
	logger.Info(successMsg)
	return nil
}
