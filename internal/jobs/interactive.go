// File : internal/jobs/interactive.go
// Deskripsi : Mode interaktif untuk monitoring job systemd sfdbtools
// Author : Hadiyatna Muflihun
// Tanggal : 4 Januari 2026
// Last Modified : 5 Januari 2026
package jobs

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"

	"sfdbtools/internal/services/scheduler"
	"sfdbtools/internal/shared/runtimecfg"
	"sfdbtools/internal/shared/validation"
	"sfdbtools/internal/ui/print"
	"sfdbtools/internal/ui/prompt"

	"github.com/mattn/go-isatty"
)

type unitInfo struct {
	Unit    string
	Label   string
	SortKey string
}

func IsInteractiveTTY() bool {
	return isatty.IsTerminal(os.Stdin.Fd()) && isatty.IsTerminal(os.Stdout.Fd())
}

// IsInteractiveAllowed menentukan apakah mode interaktif boleh digunakan.
// Aturan global: --quiet atau --daemon berarti non-interaktif.
// Selain itu, interaktif hanya aman jika stdin/stdout adalah TTY.
func IsInteractiveAllowed() bool {
	if runtimecfg.IsQuiet() || runtimecfg.IsDaemon() {
		return false
	}
	return IsInteractiveTTY()
}

func RunInteractive(ctx context.Context, defaultScope scheduler.Scope, isRoot bool) error {
	print.PrintInfo("Pilih scope lalu pilih aksi/unit")
	print.PrintInfo("Tip: scope=system untuk scheduler via sudo")

	// 1) Scope menu
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
		return validation.HandleInputError(err)
	}
	selectedScope, err := scheduler.NormalizeScope(selectedScopeStr)
	if err != nil {
		return err
	}

	// 2) Action menu
	actions := []string{"List", "Status", "Logs", "Stop", "Remove"}
	action, _, err := prompt.SelectOne("Aksi?", actions, -1)
	if err != nil {
		return validation.HandleInputError(err)
	}
	print.PrintSubHeader(action)

	// 3) Execute
	switch action {
	case "List":
		return PrintListBody(ctx, selectedScope, isRoot)
	case "Status", "Logs", "Stop", "Remove":
		units, usedScope, err := collectUnits(ctx, selectedScope)
		if err != nil {
			return err
		}
		print.PrintInfo(fmt.Sprintf("Scope: %s", usedScope))
		if len(units) == 0 {
			print.PrintWarn("Tidak ada unit sfdbtools ditemukan")
			return nil
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
			return validation.HandleInputError(err)
		}
		pickedUnit := byLabel[picked].Unit

		switch action {
		case "Status":
			out, used, err := Status(ctx, selectedScope, pickedUnit)
			if err != nil {
				return err
			}
			if used != "" && used != usedScope {
				print.PrintInfo(fmt.Sprintf("Scope: %s", used))
			}
			fmt.Print(out)
			return nil
		case "Logs":
			lines := 200
			follow := false
			if v, askErr := prompt.AskInt("Jumlah baris log?", 200, nil); askErr == nil {
				lines = v
			} else {
				return validation.HandleInputError(askErr)
			}
			if v, askErr := prompt.Confirm("Ikuti log realtime (-f)?", false); askErr == nil {
				follow = v
			} else {
				return validation.HandleInputError(askErr)
			}
			out, used, err := Logs(ctx, selectedScope, pickedUnit, lines, follow)
			if err != nil {
				return err
			}
			if used != "" && used != usedScope {
				print.PrintInfo(fmt.Sprintf("Scope: %s", used))
			}
			fmt.Print(out)
			return nil
		case "Stop":
			ok, err := prompt.Confirm(fmt.Sprintf("Stop unit %s?", pickedUnit), false)
			if err != nil {
				return validation.HandleInputError(err)
			}
			if !ok {
				print.PrintInfo("Dibatalkan")
				return nil
			}
			out, used, err := Stop(ctx, selectedScope, pickedUnit)
			if err != nil {
				return err
			}
			if used != "" && used != usedScope {
				print.PrintInfo(fmt.Sprintf("Scope: %s", used))
			}
			fmt.Print(out)
			print.PrintSuccess("Stop command dikirim")
			return nil
		case "Remove":
			ok, err := prompt.Confirm(fmt.Sprintf("Hapus job %s?", pickedUnit), false)
			if err != nil {
				return validation.HandleInputError(err)
			}
			if !ok {
				print.PrintInfo("Dibatalkan")
				return nil
			}

			purge := false
			if selectedScope == scheduler.ScopeSystem {
				// Purge butuh root untuk hapus file unit.
				if isRoot {
					purge, err = prompt.Confirm("Sekalian hapus unit file (/etc/systemd/system) bila ada?", false)
					if err != nil {
						return validation.HandleInputError(err)
					}
				} else {
					print.PrintInfo("Tip: jalankan dengan sudo untuk --purge")
				}
			}

			out, used, err := Remove(ctx, pickedUnit, RemoveOptions{Scope: selectedScope, Purge: purge})
			if err != nil {
				return err
			}
			if used != "" && used != usedScope {
				print.PrintInfo(fmt.Sprintf("Scope: %s", used))
			}
			if out != "" {
				fmt.Print(out)
			}
			print.PrintSuccess("Remove selesai")
			return nil
		}
	}

	// fallback
	return PrintListBody(ctx, selectedScope, isRoot)
}

func reorderWithDefault(options []string, def string) []string {
	if def == "" {
		return options
	}
	out := make([]string, 0, len(options))
	for _, o := range options {
		if o == def {
			out = append(out, o)
			break
		}
	}
	for _, o := range options {
		if o != def {
			out = append(out, o)
		}
	}
	return out
}

func collectUnits(ctx context.Context, scope scheduler.Scope) ([]unitInfo, scheduler.Scope, error) {
	out, err := List(ctx, scope)
	if err != nil {
		return nil, "", err
	}

	used := out.UsedServices
	if used == "" {
		used = out.UsedTimers
	}

	units := make([]unitInfo, 0, 64)
	units = append(units, parseTimers(out.Timers)...)
	units = append(units, parseServices(out.Services)...)
	// stable-ish ordering before label sort
	sort.Slice(units, func(i, j int) bool { return units[i].SortKey < units[j].SortKey })
	return units, used, nil
}

func parseTimers(out string) []unitInfo {
	lines := strings.Split(strings.TrimSpace(out), "\n")
	units := make([]unitInfo, 0, len(lines))
	for _, line := range lines {
		l := strings.TrimSpace(line)
		if l == "" {
			continue
		}
		fields := strings.Fields(l)
		if len(fields) < 2 {
			continue
		}
		// systemctl bisa prefix status marker seperti "●" di awal baris
		if fields[0] == "●" {
			fields = fields[1:]
			if len(fields) < 2 {
				continue
			}
		}
		unit := fields[len(fields)-2]
		activates := fields[len(fields)-1]
		label := fmt.Sprintf("%s (timer -> %s)", unit, activates)
		units = append(units, unitInfo{Unit: unit, Label: label, SortKey: "1|" + unit})
	}
	return units
}

func parseServices(out string) []unitInfo {
	lines := strings.Split(strings.TrimSpace(out), "\n")
	units := make([]unitInfo, 0, len(lines))
	for _, line := range lines {
		l := strings.TrimSpace(line)
		if l == "" {
			continue
		}
		fields := strings.Fields(l)
		if len(fields) < 4 {
			continue
		}
		// systemctl bisa prefix status marker seperti "●" di awal baris
		if fields[0] == "●" {
			fields = fields[1:]
			if len(fields) < 4 {
				continue
			}
		}
		unit := fields[0]
		active := fields[2]
		sub := fields[3]
		label := fmt.Sprintf("%s (%s/%s)", unit, active, sub)
		units = append(units, unitInfo{Unit: unit, Label: label, SortKey: "0|" + unit})
	}
	return units
}
