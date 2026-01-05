// File : internal/jobs/command.go
// Deskripsi : Logic command untuk monitoring systemd jobs (list/status/logs/stop)
// Author : Hadiyatna Muflihun
// Tanggal : 2026-01-04
// Last Modified : 2026-01-05
package jobs

import (
	"context"
	"fmt"
	"strings"

	"sfDBTools/internal/services/scheduler"
	"sfDBTools/pkg/ui"
)

type ListOutput struct {
	Timers       string
	UsedTimers   schedulerutil.Scope
	Services     string
	UsedServices schedulerutil.Scope
}

func List(ctx context.Context, scope schedulerutil.Scope) (ListOutput, error) {
	timers, usedTimers, err := schedulerutil.TryWithScopes(scope, func(s schedulerutil.Scope) (string, error) {
		return schedulerutil.ListTimers(ctx, s)
	})
	if err != nil {
		// tetap lanjut list services di bawah, supaya user tetap dapat info
		usedTimers = ""
	}

	services, usedServices, sErr := schedulerutil.TryWithScopes(scope, func(s schedulerutil.Scope) (string, error) {
		return schedulerutil.ListServices(ctx, s)
	})
	if sErr != nil {
		// jika dua-duanya error, return error pertama
		if err != nil {
			return ListOutput{}, err
		}
		return ListOutput{}, sErr
	}

	return ListOutput{
		Timers:       timers,
		UsedTimers:   usedTimers,
		Services:     services,
		UsedServices: usedServices,
	}, nil
}

func PrintList(ctx context.Context, scope schedulerutil.Scope, isRoot bool) error {
	return PrintListBody(ctx, scope, isRoot)
}

// PrintListBody menampilkan list jobs tanpa mencetak header utama.
// Dipakai oleh mode interaktif agar header tidak dobel.
func PrintListBody(ctx context.Context, scope schedulerutil.Scope, isRoot bool) error {
	ui.PrintInfo("Menampilkan unit sfdbtools (service + timer)")
	ui.PrintInfo("Tip: gunakan --scope=system bila scheduler dijalankan via sudo")

	out, err := List(ctx, scope)
	if err != nil {
		ui.PrintWarning(fmt.Sprintf("Gagal list jobs (scope=%s): %v", scope, err))
		return err
	}

	if out.Timers != "" {
		ui.PrintInfo(fmt.Sprintf("Timers (scope=%s)", out.UsedTimers))
		v := strings.TrimSpace(out.Timers)
		if v == "" {
			fmt.Println("- (tidak ada)")
		} else {
			fmt.Println(v)
		}
	} else {
		ui.PrintInfo("Timers")
		fmt.Println("- (tidak ada)")
	}

	ui.PrintSeparator()

	ui.PrintInfo(fmt.Sprintf("Services (scope=%s)", out.UsedServices))
	v := strings.TrimSpace(out.Services)
	if v == "" {
		fmt.Println("- (tidak ada)")
	} else {
		fmt.Println(v)
	}

	// Hint untuk system scope.
	if !isRoot && (scope == schedulerutil.ScopeSystem || scope == schedulerutil.ScopeBoth) {
		ui.PrintInfo("Jika ada error permission saat akses system scope, jalankan dengan sudo.")
	}
	return nil
}

func Status(ctx context.Context, scope schedulerutil.Scope, unit string) (string, schedulerutil.Scope, error) {
	out, used, err := schedulerutil.TryWithScopes(scope, func(s schedulerutil.Scope) (string, error) {
		return schedulerutil.StatusUnit(ctx, s, unit)
	})
	return out, used, err
}

func Logs(ctx context.Context, scope schedulerutil.Scope, unit string, lines int, follow bool) (string, schedulerutil.Scope, error) {
	out, used, err := schedulerutil.TryWithScopes(scope, func(s schedulerutil.Scope) (string, error) {
		return schedulerutil.LogsUnit(ctx, s, unit, lines, follow)
	})
	return out, used, err
}

func Stop(ctx context.Context, scope schedulerutil.Scope, unit string) (string, schedulerutil.Scope, error) {
	out, used, err := schedulerutil.TryWithScopes(scope, func(s schedulerutil.Scope) (string, error) {
		return schedulerutil.StopUnit(ctx, s, unit)
	})
	return out, used, err
}
