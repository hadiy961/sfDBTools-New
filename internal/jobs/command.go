// File : internal/jobs/command.go
// Deskripsi : Logic command untuk monitoring systemd jobs (list/status/logs/stop)
// Author : Hadiyatna Muflihun
// Tanggal : 4 Januari 2026
// Last Modified : 5 Januari 2026
package jobs

import (
	"context"
	"fmt"
	"strings"

	"sfdbtools/internal/services/scheduler"
	"sfdbtools/internal/ui/print"
)

type ListOutput struct {
	Timers       string
	UsedTimers   scheduler.Scope
	Services     string
	UsedServices scheduler.Scope
}

func List(ctx context.Context, scope scheduler.Scope) (ListOutput, error) {
	timers, usedTimers, err := scheduler.TryWithScopes(scope, func(s scheduler.Scope) (string, error) {
		return scheduler.ListTimers(ctx, s)
	})
	if err != nil {
		// tetap lanjut list services di bawah, supaya user tetap dapat info
		usedTimers = ""
	}

	services, usedServices, sErr := scheduler.TryWithScopes(scope, func(s scheduler.Scope) (string, error) {
		return scheduler.ListServices(ctx, s)
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

func PrintList(ctx context.Context, scope scheduler.Scope, isRoot bool) error {
	return PrintListBody(ctx, scope, isRoot)
}

// PrintListBody menampilkan list jobs tanpa mencetak header utama.
// Dipakai oleh mode interaktif agar header tidak dobel.
func PrintListBody(ctx context.Context, scope scheduler.Scope, isRoot bool) error {
	print.PrintInfo("Menampilkan unit sfdbtools (service + timer)")
	print.PrintInfo("Tip: gunakan --scope=system bila scheduler dijalankan via sudo")

	out, err := List(ctx, scope)
	if err != nil {
		print.PrintWarn(fmt.Sprintf("Gagal list jobs (scope=%s): %v", scope, err))
		return err
	}

	if out.Timers != "" {
		print.PrintInfo(fmt.Sprintf("Timers (scope=%s)", out.UsedTimers))
		v := strings.TrimSpace(out.Timers)
		if v == "" {
			fmt.Println("- (tidak ada)")
		} else {
			fmt.Println(v)
		}
	} else {
		print.PrintInfo("Timers")
		fmt.Println("- (tidak ada)")
	}

	print.PrintSeparator()

	print.PrintInfo(fmt.Sprintf("Services (scope=%s)", out.UsedServices))
	v := strings.TrimSpace(out.Services)
	if v == "" {
		fmt.Println("- (tidak ada)")
	} else {
		fmt.Println(v)
	}

	// Hint untuk system scope.
	if !isRoot && (scope == scheduler.ScopeSystem || scope == scheduler.ScopeBoth) {
		print.PrintInfo("Jika ada error permission saat akses system scope, jalankan dengan sudo.")
	}
	return nil
}

func Status(ctx context.Context, scope scheduler.Scope, unit string) (string, scheduler.Scope, error) {
	out, used, err := scheduler.TryWithScopes(scope, func(s scheduler.Scope) (string, error) {
		return scheduler.StatusUnit(ctx, s, unit)
	})
	return out, used, err
}

func Logs(ctx context.Context, scope scheduler.Scope, unit string, lines int, follow bool) (string, scheduler.Scope, error) {
	out, used, err := scheduler.TryWithScopes(scope, func(s scheduler.Scope) (string, error) {
		return scheduler.LogsUnit(ctx, s, unit, lines, follow)
	})
	return out, used, err
}

func Stop(ctx context.Context, scope scheduler.Scope, unit string) (string, scheduler.Scope, error) {
	out, used, err := scheduler.TryWithScopes(scope, func(s scheduler.Scope) (string, error) {
		return scheduler.StopUnit(ctx, s, unit)
	})
	return out, used, err
}
