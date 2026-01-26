// File : internal/app/profile/wizard/import_source.go
// Deskripsi : Import source selection wizard (XLSX/GSheet)
// Author : Hadiyatna Muflihun
// Tanggal : 26 Januari 2026
// Last Modified : 26 Januari 2026

package wizard

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"sfdbtools/internal/app/profile/helpers/reader"
	profilemodel "sfdbtools/internal/app/profile/model"
	sharedvalidation "sfdbtools/internal/shared/validation"
	"sfdbtools/internal/ui/prompt"

	"github.com/AlecAivazis/survey/v2"
)

// readSource membaca data dari XLSX lokal atau Google Spreadsheet
// Handles interactive source selection, sheet/tab selection
func (w *ImportWizard) readSource(opts *profilemodel.ProfileImportOptions) ([]string, [][]string, []string, string, error) {
	// Pilih Source (full-interaktif) jika kedua source kosong
	if strings.TrimSpace(opts.Input) == "" && strings.TrimSpace(opts.GSheetURL) == "" {
		if opts.SkipConfirm {
			return nil, nil, nil, "(unknown)", fmt.Errorf("sumber import kosong: gunakan --input atau --gsheet (automation wajib --skip-confirm)")
		}
		if err := w.promptSourceType(opts); err != nil {
			return nil, nil, nil, "(unknown)", err
		}
	}

	// Route ke XLSX atau GSheet based on what's set
	if strings.TrimSpace(opts.Input) != "" {
		return w.readXLSXSource(opts)
	}
	if strings.TrimSpace(opts.GSheetURL) != "" {
		return w.readGSheetSource(opts)
	}

	return nil, nil, nil, "(unknown)", fmt.Errorf("sumber import kosong")
}

// promptSourceType prompts user untuk memilih source type
func (w *ImportWizard) promptSourceType(opts *profilemodel.ProfileImportOptions) error {
	choice, _, err := prompt.SelectOne(
		"Pilih sumber import profile:",
		[]string{"XLSX lokal", "Google Spreadsheet", "Batal"},
		0,
	)
	if err != nil {
		return sharedvalidation.HandleInputError(err)
	}

	switch choice {
	case "XLSX lokal":
		selected, selErr := prompt.SelectFile(".", "Pilih file XLSX untuk import profile", []string{".xlsx"})
		if selErr != nil {
			return sharedvalidation.HandleInputError(selErr)
		}
		opts.Input = strings.TrimSpace(selected)
		opts.GSheetURL = ""

	case "Google Spreadsheet":
		urlStr, err := prompt.AskText("Masukkan URL Google Spreadsheet:", prompt.WithValidator(func(ans interface{}) error {
			v := strings.TrimSpace(fmt.Sprintf("%v", ans))
			if v == "" {
				return fmt.Errorf("url tidak boleh kosong")
			}
			return nil
		}))
		if err != nil {
			return sharedvalidation.HandleInputError(err)
		}
		opts.GSheetURL = strings.TrimSpace(urlStr)
		opts.Input = ""

	default:
		return sharedvalidation.ErrUserCancelled
	}

	return nil
}

// readXLSXSource reads from local XLSX file (dengan interactive sheet selection)
func (w *ImportWizard) readXLSXSource(opts *profilemodel.ProfileImportOptions) ([]string, [][]string, []string, string, error) {
	// Pilih Sheet (XLSX) - hanya jika --sheet kosong
	if strings.TrimSpace(opts.Sheet) == "" && !opts.SkipConfirm {
		sheets, err := reader.ListXLSXSheets(opts.Input)
		if err != nil {
			return nil, nil, nil, "XLSX", err
		}
		if len(sheets) == 1 {
			opts.Sheet = sheets[0]
			w.Log.Infof("[profile-import] XLSX hanya punya 1 sheet, auto pilih: %s", opts.Sheet)
		} else if len(sheets) > 1 {
			sel, _, selErr := prompt.SelectOne("Pilih sheet XLSX:", sheets, 0)
			if selErr != nil {
				return nil, nil, nil, "XLSX", sharedvalidation.HandleInputError(selErr)
			}
			opts.Sheet = strings.TrimSpace(sel)
		}
	}

	w.Log.Infof("[profile-import] Membaca XLSX lokal: %s (sheet=%s)", opts.Input, opts.Sheet)
	res, err := reader.ReadXLSX(opts.Input, opts.Sheet)
	if err != nil {
		return nil, nil, nil, "XLSX", err
	}
	return res.Headers, res.DataRows, res.Warnings, res.Source, nil
}

// readGSheetSource reads from Google Spreadsheet (dengan interactive tab selection)
func (w *ImportWizard) readGSheetSource(opts *profilemodel.ProfileImportOptions) ([]string, [][]string, []string, string, error) {
	// Pilih Sheet (Google Spreadsheet)
	if opts.GID < 0 && !opts.SkipConfirm {
		tabs, err := reader.ListGoogleSheetTabsBestEffort(opts.GSheetURL)
		if err == nil && len(tabs) > 0 {
			if err := w.promptGSheetTab(tabs, opts); err != nil {
				return nil, nil, nil, "Google Spreadsheet", err
			}
		} else {
			if err := w.promptGSheetGID(opts); err != nil {
				return nil, nil, nil, "Google Spreadsheet", err
			}
		}
	}

	w.Log.Infof("[profile-import] Membaca Google Spreadsheet: %s (gid=%d)", opts.GSheetURL, opts.GID)
	res, err := reader.ReadGoogleSheetAsCSV(opts.GSheetURL, opts.GID)
	if err != nil {
		return nil, nil, nil, "Google Spreadsheet", err
	}
	return res.Headers, res.DataRows, res.Warnings, res.Source, nil
}

// promptGSheetTab prompts user untuk memilih tab dari list (jika tab info tersedia)
func (w *ImportWizard) promptGSheetTab(tabs []reader.GoogleSheetTab, opts *profilemodel.ProfileImportOptions) error {
	items := make([]string, 0, len(tabs))
	for _, t := range tabs {
		name := strings.TrimSpace(t.Name)
		if name == "" {
			name = fmt.Sprintf("(gid=%d)", t.GID)
		} else {
			name = fmt.Sprintf("%s (gid=%d)", name, t.GID)
		}
		items = append(items, name)
	}
	// Stabil: sort by gid agar tidak random
	sort.SliceStable(items, func(i, j int) bool { return items[i] < items[j] })

	if len(items) == 1 {
		// parse gid dari string
		gidVal := extractGIDFromLabel(items[0])
		if gidVal >= 0 {
			opts.GID = gidVal
		}
		return nil
	}

	sel, _, selErr := prompt.SelectOne("Pilih tab Google Sheet:", items, 0)
	if selErr != nil {
		return sharedvalidation.HandleInputError(selErr)
	}
	gidVal := extractGIDFromLabel(sel)
	if gidVal < 0 {
		return fmt.Errorf("gagal membaca gid dari pilihan: %s", sel)
	}
	opts.GID = gidVal
	return nil
}

// promptGSheetGID prompts user untuk input gid manual
func (w *ImportWizard) promptGSheetGID(opts *profilemodel.ProfileImportOptions) error {
	gidVal, askErr := prompt.AskInt(
		"Masukkan gid tab Google Sheet (default: 0):",
		0,
		survey.Validator(func(ans interface{}) error {
			v := strings.TrimSpace(fmt.Sprintf("%v", ans))
			if v == "" {
				return nil
			}
			n, err := strconv.Atoi(v)
			if err != nil {
				return fmt.Errorf("gid harus angka")
			}
			if n < 0 {
				return fmt.Errorf("gid tidak boleh negatif")
			}
			return nil
		}),
	)
	if askErr != nil {
		return sharedvalidation.HandleInputError(askErr)
	}
	opts.GID = gidVal
	return nil
}

// extractGIDFromLabel extracts gid dari label format "<name> (gid=<n>)"
func extractGIDFromLabel(s string) int {
	open := strings.LastIndex(s, "gid=")
	if open < 0 {
		return -1
	}
	part := s[open+len("gid="):]
	part = strings.TrimSpace(part)
	part = strings.TrimSuffix(part, ")")
	n, err := strconv.Atoi(part)
	if err != nil {
		return -1
	}
	return n
}
