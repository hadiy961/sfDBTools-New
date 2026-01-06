// File : internal/jobs/interactive_subcommands.go
// Deskripsi : Mode interaktif untuk subcommand jobs saat argumen belum lengkap
// Author : Hadiyatna Muflihun
// Tanggal : 4 Januari 2026
// Last Modified : 5 Januari 2026
package jobs

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"sfdbtools/internal/services/scheduler"
	"sfdbtools/internal/ui/print"
	"sfdbtools/internal/ui/prompt"
	"sfdbtools/pkg/validation"
)

func pickScopeInteractive(defaultScope scheduler.Scope) (scheduler.Scope, error) {
	print.PrintInfo("Pilih scope unit")
	scopeOptions := []string{"auto", "user", "system", "both"}
	def := strings.ToLower(string(defaultScope))
	if def == "" {
		def = "auto"
	}
	if def != "auto" && def != "user" && def != "system" && def != "both" {
		def = "auto"
	}
	orderedScopes := reorderWithDefault(scopeOptions, def)
	selectedScopeStr, _, err := prompt.SelectOne("Scope unit?", orderedScopes, 0)
	if err != nil {
		return "", validation.HandleInputError(err)
	}
	selectedScope, err := scheduler.NormalizeScope(selectedScopeStr)
	if err != nil {
		return "", err
	}
	return selectedScope, nil
}

func pickUnitInteractive(ctx context.Context, scope scheduler.Scope) (string, scheduler.Scope, error) {
	units, usedScope, err := collectUnits(ctx, scope)
	if err != nil {
		return "", usedScope, err
	}
	if len(units) == 0 {
		print.PrintWarn("Tidak ada unit sfdbtools ditemukan")
		return "", usedScope, nil
	}

	labels := make([]string, 0, len(units))
	byLabel := make(map[string]unitInfo, len(units))
	for _, u := range units {
		labels = append(labels, u.Label)
		byLabel[u.Label] = u
	}
	sort.Strings(labels)
	picked, _, err := prompt.SelectOne("Pilih unit", labels, -1)
	if err != nil {
		return "", usedScope, validation.HandleInputError(err)
	}
	pickedUnit := byLabel[picked].Unit
	return pickedUnit, usedScope, nil
}

func RunInteractiveStatus(ctx context.Context, scope scheduler.Scope, scopeSet bool) error {
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
	print.PrintInfo(fmt.Sprintf("Scope: %s", usedScope))
	out, used, err := Status(ctx, scope, unit)
	if err != nil {
		return err
	}
	if used != "" && used != usedScope {
		print.PrintInfo(fmt.Sprintf("Scope: %s", used))
	}
	fmt.Print(out)
	return nil
}

func RunInteractiveLogs(ctx context.Context, scope scheduler.Scope, scopeSet bool, lines int, follow bool, linesSet bool, followSet bool) error {
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
	print.PrintInfo(fmt.Sprintf("Scope: %s", usedScope))

	// Jika user belum set flag, tanya interaktif.
	if !linesSet {
		v, askErr := prompt.AskInt("Jumlah baris log?", lines, nil)
		if askErr != nil {
			return validation.HandleInputError(askErr)
		}
		lines = v
	}
	if !followSet {
		v, askErr := prompt.Confirm("Ikuti log realtime (-f)?", follow)
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
		print.PrintInfo(fmt.Sprintf("Scope: %s", used))
	}
	fmt.Print(out)
	return nil
}

func RunInteractiveStop(ctx context.Context, scope scheduler.Scope, scopeSet bool) error {
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
	print.PrintInfo(fmt.Sprintf("Scope: %s", usedScope))
	ok, err := prompt.Confirm(fmt.Sprintf("Stop unit %s?", unit), false)
	if err != nil {
		return validation.HandleInputError(err)
	}
	if !ok {
		print.PrintInfo("Dibatalkan")
		return nil
	}

	out, used, err := Stop(ctx, scope, unit)
	if err != nil {
		return err
	}
	if used != "" && used != usedScope {
		print.PrintInfo(fmt.Sprintf("Scope: %s", used))
	}
	fmt.Print(out)
	print.PrintSuccess("Stop command dikirim")
	return nil
}

func RunInteractiveRemove(ctx context.Context, scope scheduler.Scope, scopeSet bool, isRoot bool, purgeFlag bool, purgeSet bool) (string, scheduler.Scope, error) {
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
	print.PrintInfo(fmt.Sprintf("Scope: %s", usedScope))
	ok, err := prompt.Confirm(fmt.Sprintf("Hapus job %s?", unit), false)
	if err != nil {
		return "", "", validation.HandleInputError(err)
	}
	if !ok {
		print.PrintInfo("Dibatalkan")
		return "", "", validation.ErrUserCancelled
	}

	purge := purgeFlag
	if !purgeSet {
		purge = false
		if scope == scheduler.ScopeSystem {
			if isRoot {
				purge, err = prompt.Confirm("Sekalian hapus unit file (/etc/systemd/system) bila ada?", false)
				if err != nil {
					return "", "", validation.HandleInputError(err)
				}
			} else {
				print.PrintInfo("Tip: jalankan dengan sudo untuk --purge")
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
