// File : internal/jobs/interactive_subcommands.go
// Deskripsi : Mode interaktif untuk subcommand jobs saat argumen belum lengkap
// Author : Hadiyatna Muflihun
// Tanggal : 2026-01-04
// Last Modified : 2026-01-05
package jobs

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"sfDBTools/internal/services/scheduler"
	"sfDBTools/pkg/input"
	"sfDBTools/pkg/ui"
	"sfDBTools/pkg/validation"
)

func pickScopeInteractive(defaultScope schedulerutil.Scope) (schedulerutil.Scope, error) {
	ui.PrintInfo("Pilih scope unit")
	scopeOptions := []string{"auto", "user", "system", "both"}
	def := strings.ToLower(string(defaultScope))
	if def == "" {
		def = "auto"
	}
	if def != "auto" && def != "user" && def != "system" && def != "both" {
		def = "auto"
	}
	orderedScopes := reorderWithDefault(scopeOptions, def)
	idx, err := input.ShowMenu("Scope unit?", orderedScopes)
	if err != nil {
		return "", validation.HandleInputError(err)
	}
	selectedScope, err := schedulerutil.NormalizeScope(orderedScopes[idx-1])
	if err != nil {
		return "", err
	}
	return selectedScope, nil
}

func pickUnitInteractive(ctx context.Context, scope schedulerutil.Scope) (string, schedulerutil.Scope, error) {
	units, usedScope, err := collectUnits(ctx, scope)
	if err != nil {
		return "", usedScope, err
	}
	if len(units) == 0 {
		ui.PrintWarning("Tidak ada unit sfdbtools ditemukan")
		return "", usedScope, nil
	}

	labels := make([]string, 0, len(units))
	byLabel := make(map[string]unitInfo, len(units))
	for _, u := range units {
		labels = append(labels, u.Label)
		byLabel[u.Label] = u
	}
	sort.Strings(labels)
	picked, err := input.SelectSingleFromList(labels, "Pilih unit")
	if err != nil {
		return "", usedScope, validation.HandleInputError(err)
	}
	pickedUnit := byLabel[picked].Unit
	return pickedUnit, usedScope, nil
}

func RunInteractiveStatus(ctx context.Context, scope schedulerutil.Scope, scopeSet bool) error {
	if !scopeSet {
		picked, err := pickScopeInteractive(scope)
		if err != nil {
			return err
		}
		scope = picked
	}
	unit, usedScope, err := pickUnitInteractive(ctx, scope)
	if err != nil {
		return err
	}
	if unit == "" {
		return nil
	}
	ui.PrintInfo(fmt.Sprintf("Scope: %s", usedScope))
	out, used, err := Status(ctx, scope, unit)
	if err != nil {
		return err
	}
	if used != "" && used != usedScope {
		ui.PrintInfo(fmt.Sprintf("Scope: %s", used))
	}
	fmt.Print(out)
	return nil
}

func RunInteractiveLogs(ctx context.Context, scope schedulerutil.Scope, scopeSet bool, lines int, follow bool, linesSet bool, followSet bool) error {
	if !scopeSet {
		picked, err := pickScopeInteractive(scope)
		if err != nil {
			return err
		}
		scope = picked
	}
	unit, usedScope, err := pickUnitInteractive(ctx, scope)
	if err != nil {
		return err
	}
	if unit == "" {
		return nil
	}
	ui.PrintInfo(fmt.Sprintf("Scope: %s", usedScope))

	// Jika user belum set flag, tanya interaktif.
	if !linesSet {
		v, askErr := input.AskInt("Jumlah baris log?", lines, nil)
		if askErr != nil {
			return validation.HandleInputError(askErr)
		}
		lines = v
	}
	if !followSet {
		v, askErr := input.AskYesNo("Ikuti log realtime (-f)?", follow)
		if askErr != nil {
			return validation.HandleInputError(askErr)
		}
		follow = v
	}

	out, used, err := Logs(ctx, scope, unit, lines, follow)
	if err != nil {
		return err
	}
	if used != "" && used != usedScope {
		ui.PrintInfo(fmt.Sprintf("Scope: %s", used))
	}
	fmt.Print(out)
	return nil
}

func RunInteractiveStop(ctx context.Context, scope schedulerutil.Scope, scopeSet bool) error {
	if !scopeSet {
		picked, err := pickScopeInteractive(scope)
		if err != nil {
			return err
		}
		scope = picked
	}
	unit, usedScope, err := pickUnitInteractive(ctx, scope)
	if err != nil {
		return err
	}
	if unit == "" {
		return nil
	}
	ui.PrintInfo(fmt.Sprintf("Scope: %s", usedScope))
	ok, err := input.AskYesNo(fmt.Sprintf("Stop unit %s?", unit), false)
	if err != nil {
		return validation.HandleInputError(err)
	}
	if !ok {
		ui.PrintInfo("Dibatalkan")
		return nil
	}

	out, used, err := Stop(ctx, scope, unit)
	if err != nil {
		return err
	}
	if used != "" && used != usedScope {
		ui.PrintInfo(fmt.Sprintf("Scope: %s", used))
	}
	fmt.Print(out)
	ui.PrintSuccess("Stop command dikirim")
	return nil
}

func RunInteractiveRemove(ctx context.Context, scope schedulerutil.Scope, scopeSet bool, isRoot bool, purgeFlag bool, purgeSet bool) (string, schedulerutil.Scope, error) {
	if !scopeSet {
		picked, err := pickScopeInteractive(scope)
		if err != nil {
			return "", "", err
		}
		scope = picked
	}
	unit, usedScope, err := pickUnitInteractive(ctx, scope)
	if err != nil {
		return "", "", err
	}
	if unit == "" {
		return "", "", nil
	}
	ui.PrintInfo(fmt.Sprintf("Scope: %s", usedScope))
	ok, err := input.AskYesNo(fmt.Sprintf("Hapus job %s?", unit), false)
	if err != nil {
		return "", "", validation.HandleInputError(err)
	}
	if !ok {
		ui.PrintInfo("Dibatalkan")
		return "", "", validation.ErrUserCancelled
	}

	purge := purgeFlag
	if !purgeSet {
		purge = false
		if scope == schedulerutil.ScopeSystem {
			if isRoot {
				purge, err = input.AskYesNo("Sekalian hapus unit file (/etc/systemd/system) bila ada?", false)
				if err != nil {
					return "", "", validation.HandleInputError(err)
				}
			} else {
				ui.PrintInfo("Tip: jalankan dengan sudo untuk --purge")
			}
		}
	}

	// Jika purgeFlag sudah diset oleh user, hormati apa adanya (termasuk false).
	if purgeSet {
		purge = purgeFlag
	}

	out, used, err := Remove(ctx, unit, RemoveOptions{Scope: scope, Purge: purge})
	if err != nil {
		return out, used, err
	}

	return out, used, nil
}
