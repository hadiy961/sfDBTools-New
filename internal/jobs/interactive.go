// File : internal/jobs/interactive.go
// Deskripsi : Mode interaktif untuk monitoring job systemd sfDBTools
// Author : Hadiyatna Muflihun
// Tanggal : 2026-01-04
// Last Modified : 2026-01-04

package jobs

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"

	"sfDBTools/internal/schedulerutil"
	"sfDBTools/pkg/input"
	"sfDBTools/pkg/runtimecfg"
	"sfDBTools/pkg/ui"
	"sfDBTools/pkg/validation"

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

func RunInteractive(ctx context.Context, defaultScope schedulerutil.Scope, isRoot bool) error {
	ui.PrintInfo("Pilih scope lalu pilih aksi/unit")
	ui.PrintInfo("Tip: scope=system untuk scheduler via sudo")

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
	idx, err := input.ShowMenu("Scope unit?", orderedScopes)
	if err != nil {
		return validation.HandleInputError(err)
	}
	selectedScope, err := schedulerutil.NormalizeScope(orderedScopes[idx-1])
	if err != nil {
		return err
	}

	// 2) Action menu
	actions := []string{"List", "Status", "Logs", "Stop", "Remove"}
	aIdx, err := input.ShowMenu("Aksi?", actions)
	if err != nil {
		return validation.HandleInputError(err)
	}
	action := actions[aIdx-1]
	ui.PrintSubHeader(action)

	// 3) Execute
	switch action {
	case "List":
		return PrintListBody(ctx, selectedScope, isRoot)
	case "Status", "Logs", "Stop", "Remove":
		units, usedScope, err := collectUnits(ctx, selectedScope)
		if err != nil {
			return err
		}
		ui.PrintInfo(fmt.Sprintf("Scope: %s", usedScope))
		if len(units) == 0 {
			ui.PrintWarning("Tidak ada unit sfdbtools ditemukan")
			return nil
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
				ui.PrintInfo(fmt.Sprintf("Scope: %s", used))
			}
			fmt.Print(out)
			return nil
		case "Logs":
			lines := 200
			follow := false
			if v, askErr := input.AskInt("Jumlah baris log?", 200, nil); askErr == nil {
				lines = v
			} else {
				return validation.HandleInputError(askErr)
			}
			if v, askErr := input.AskYesNo("Ikuti log realtime (-f)?", false); askErr == nil {
				follow = v
			} else {
				return validation.HandleInputError(askErr)
			}
			out, used, err := Logs(ctx, selectedScope, pickedUnit, lines, follow)
			if err != nil {
				return err
			}
			if used != "" && used != usedScope {
				ui.PrintInfo(fmt.Sprintf("Scope: %s", used))
			}
			fmt.Print(out)
			return nil
		case "Stop":
			ok, err := input.AskYesNo(fmt.Sprintf("Stop unit %s?", pickedUnit), false)
			if err != nil {
				return validation.HandleInputError(err)
			}
			if !ok {
				ui.PrintInfo("Dibatalkan")
				return nil
			}
			out, used, err := Stop(ctx, selectedScope, pickedUnit)
			if err != nil {
				return err
			}
			if used != "" && used != usedScope {
				ui.PrintInfo(fmt.Sprintf("Scope: %s", used))
			}
			fmt.Print(out)
			ui.PrintSuccess("Stop command dikirim")
			return nil
		case "Remove":
			ok, err := input.AskYesNo(fmt.Sprintf("Hapus job %s?", pickedUnit), false)
			if err != nil {
				return validation.HandleInputError(err)
			}
			if !ok {
				ui.PrintInfo("Dibatalkan")
				return nil
			}

			purge := false
			if selectedScope == schedulerutil.ScopeSystem {
				// Purge butuh root untuk hapus file unit.
				if isRoot {
					purge, err = input.AskYesNo("Sekalian hapus unit file (/etc/systemd/system) bila ada?", false)
					if err != nil {
						return validation.HandleInputError(err)
					}
				} else {
					ui.PrintInfo("Tip: jalankan dengan sudo untuk --purge")
				}
			}

			out, used, err := Remove(ctx, pickedUnit, RemoveOptions{Scope: selectedScope, Purge: purge})
			if err != nil {
				return err
			}
			if used != "" && used != usedScope {
				ui.PrintInfo(fmt.Sprintf("Scope: %s", used))
			}
			if out != "" {
				fmt.Print(out)
			}
			ui.PrintSuccess("Remove selesai")
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

func collectUnits(ctx context.Context, scope schedulerutil.Scope) ([]unitInfo, schedulerutil.Scope, error) {
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
