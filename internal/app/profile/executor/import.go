// File : internal/app/profile/executor/import.go
// Deskripsi : Import bulk profile dari XLSX lokal atau Google Spreadsheet (executor layer)
// Author : Hadiyatna Muflihun
// Tanggal : 25 Januari 2026
// Last Modified : 26 Januari 2026

package executor

import (
	"fmt"
	"strings"

	importdisplay "sfdbtools/internal/app/profile/display"
	"sfdbtools/internal/app/profile/helpers/importer"
	profilemodel "sfdbtools/internal/app/profile/model"
	"sfdbtools/internal/app/profile/wizard"
	appconfig "sfdbtools/internal/services/config"
	"sfdbtools/internal/shared/consts"
	sharedvalidation "sfdbtools/internal/shared/validation"
	"sfdbtools/internal/ui/prompt"
)

func (e *Executor) ImportProfiles() error {
	if e == nil || e.State == nil {
		return fmt.Errorf("internal error: executor state tidak tersedia")
	}
	opts, ok := e.State.ImportOptions()
	if !ok || opts == nil {
		return fmt.Errorf("internal error: import options tidak tersedia")
	}

	// Validasi on-conflict strategy
	onConflict := strings.ToLower(strings.TrimSpace(opts.OnConflict))
	if onConflict == "" {
		onConflict = profilemodel.ImportConflictFail
	}
	switch onConflict {
	case profilemodel.ImportConflictFail, profilemodel.ImportConflictSkip,
		profilemodel.ImportConflictOverwrite, profilemodel.ImportConflictRename:
		// ok
	default:
		return fmt.Errorf("nilai --on-conflict tidak valid: %s (pilih: fail|skip|overwrite|rename)", opts.OnConflict)
	}

	// PHASE 1-3: Run import wizard untuk get planned rows
	cfg, ok := e.Config.(*appconfig.Config)
	if !ok {
		return fmt.Errorf("internal error: config type assertion failed")
	}
	wiz := wizard.NewImportWizard(e.Log, cfg, e.ConfigDir)
	planned, err := wiz.Run(opts)
	if err != nil {
		return err
	}

	// Extract source label untuk display
	srcLabel := "import"
	if strings.TrimSpace(opts.Input) != "" {
		srcLabel = "XLSX"
	} else if strings.TrimSpace(opts.GSheetURL) != "" {
		srcLabel = "Google Spreadsheet"
	}

	// PHASE 4: Display summary + confirmation
	if !opts.SkipConfirm {
		importdisplay.PrintImportPlanSummary(planned, srcLabel)
		msg := fmt.Sprintf("Lanjut simpan %d profile dari %s?", importer.CountPlanned(planned), srcLabel)
		ok, askErr := prompt.Confirm(msg, false)
		if askErr != nil {
			return sharedvalidation.HandleInputError(askErr)
		}
		if !ok {
			return sharedvalidation.ErrUserCancelled
		}
	}

	// PHASE 5: Save profiles
	if !opts.SkipConnTest {
		opts.ConnTestDone = true
	}

	e.Log.Infof("[profile-import] Memulai fase simpan | total_akan_simpan=%d", importer.CountPlanned(planned))
	saved := 0
	failed := 0

	for _, r := range planned {
		if r.Skip {
			continue
		}

		info := importer.BuildProfileInfoFromRow(r)
		e.State.ProfileInfo = &info

		if err := e.SaveProfile(consts.ProfileSaveModeCreate); err != nil {
			failed++
			if opts.ContinueOnError {
				e.Log.Warnf("Gagal menyimpan profile '%s' (row %d): %v", importer.SafeName(r.PlannedName), r.RowNum, err)
				continue
			}
			return err
		}
		saved++
	}

	if failed > 0 {
		e.Log.Warnf("Import selesai dengan error: saved=%d failed=%d", saved, failed)
	} else {
		e.Log.Infof("[profile-import] Import selesai: saved=%d", saved)
	}

	return nil
}
