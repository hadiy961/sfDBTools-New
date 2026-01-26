// File : internal/app/profile/wizard/import_conflict.go
// Deskripsi : Import conflict resolution wizard
// Author : Hadiyatna Muflihun
// Tanggal : 26 Januari 2026
// Last Modified : 26 Januari 2026

package wizard

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	profileconn "sfdbtools/internal/app/profile/connection"
	importdisplay "sfdbtools/internal/app/profile/display"
	"sfdbtools/internal/app/profile/helpers/resolver"
	"sfdbtools/internal/app/profile/merger"
	profilemodel "sfdbtools/internal/app/profile/model"
	"sfdbtools/internal/shared/fsops"
	sharedvalidation "sfdbtools/internal/shared/validation"
	"sfdbtools/internal/ui/prompt"
)

// handleConflicts handles file name conflicts untuk import
// Returns final conflict mode yang digunakan (untuk apply to all)
func (w *ImportWizard) handleConflicts(
	rows []profilemodel.ImportRow,
	baseDir string,
	opts *profilemodel.ProfileImportOptions,
) (string, error) {
	conflictMode := strings.ToLower(strings.TrimSpace(opts.OnConflict))
	if conflictMode == "" {
		conflictMode = profilemodel.ImportConflictFail
	}

	// Automation mode: gunakan flag --on-conflict tanpa prompt
	if opts.SkipConfirm {
		return conflictMode, nil
	}

	// Interactive mode: pre-scan conflicts + prompt global strategy
	conflicts := w.findConflictsPreview(rows, baseDir)
	if len(conflicts) == 0 {
		return conflictMode, nil
	}

	// Tampilkan ringkasan konflik
	importdisplay.PrintImportConflictSummary(conflicts, baseDir, 20)

	items := []string{"Overwrite semua", "Rename semua", "Skip semua", "Tanya per row", "Batalkan"}
	defaultIdx := 0
	switch conflictMode {
	case profilemodel.ImportConflictRename:
		defaultIdx = 1
	case profilemodel.ImportConflictSkip:
		defaultIdx = 2
	case profilemodel.ImportConflictFail:
		defaultIdx = 3
	}

	choice, _, askErr := prompt.SelectOne("Ada konflik nama profile. Pilih aksi default:", items, defaultIdx)
	if askErr != nil {
		return "", sharedvalidation.HandleInputError(askErr)
	}

	switch choice {
	case "Overwrite semua":
		conflictMode = profilemodel.ImportConflictOverwrite
	case "Rename semua":
		conflictMode = profilemodel.ImportConflictRename
	case "Skip semua":
		conflictMode = profilemodel.ImportConflictSkip
	case "Tanya per row":
		conflictMode = "ask" // per-row interactive mode
	default:
		return "", sharedvalidation.ErrUserCancelled
	}

	return conflictMode, nil
}

// resolveConflictForRow resolves conflict untuk satu row
// Returns updated row dan potentially updated conflict mode (jika user pilih "apply to all")
func (w *ImportWizard) resolveConflictForRow(
	r profilemodel.ImportRow,
	baseDir string,
	planned map[string]bool,
	mode string,
	opts *profilemodel.ProfileImportOptions,
) (profilemodel.ImportRow, string, error) {
	// Automation mode: gunakan resolver non-interactive
	if opts.SkipConfirm {
		resolved, err := resolver.ResolveConflict(w.Log, &r, opts, mode, baseDir, planned)
		return resolved, mode, err
	}

	// Interactive mode
	mode = strings.ToLower(strings.TrimSpace(mode))
	if mode == "" {
		mode = profilemodel.ImportConflictFail
	}

	// Apply action helper
	applyAction := func(action string) (profilemodel.ImportRow, error) {
		switch action {
		case profilemodel.ImportConflictOverwrite:
			r.PlanAction = profilemodel.ImportPlanOverwrite
			return r, nil
		case profilemodel.ImportConflictSkip:
			r.Skip = true
			r.SkipReason = profilemodel.ImportSkipReasonConflict
			r.PlanAction = profilemodel.ImportPlanSkip
			return r, nil
		case profilemodel.ImportConflictRename:
			r.PlanAction = profilemodel.ImportPlanRename
			r.RenamedFrom = r.PlannedName
			r.PlannedName = resolver.AutoRename(baseDir, r.PlannedName, planned)
			return r, nil
		default:
			return r, fmt.Errorf("konflik file untuk profile '%s' (row %d)", safeName(r.PlannedName), r.RowNum)
		}
	}

	// Non-ask mode: apply automatically
	if mode != "ask" {
		out, err := applyAction(mode)
		return out, mode, err
	}

	// Ask mode: prompt per row
	choice, _, askErr := prompt.SelectOne(
		fmt.Sprintf("Profile '%s' (row %d) sudah ada. Pilih aksi:", safeName(r.PlannedName), r.RowNum),
		[]string{"Overwrite", "Rename", "Skip", "Batalkan"},
		0,
	)
	if askErr != nil {
		return r, mode, sharedvalidation.HandleInputError(askErr)
	}

	action := ""
	switch choice {
	case "Overwrite":
		action = profilemodel.ImportConflictOverwrite
	case "Rename":
		action = profilemodel.ImportConflictRename
	case "Skip":
		action = profilemodel.ImportConflictSkip
	default:
		return r, mode, sharedvalidation.ErrUserCancelled
	}

	// Optional: apply to all remaining conflicts
	applyAll, applyErr := prompt.Confirm("Terapkan aksi ini untuk semua konflik berikutnya?", false)
	if applyErr != nil {
		return r, mode, sharedvalidation.HandleInputError(applyErr)
	}

	modeAfter := mode
	if applyAll {
		modeAfter = action
	}

	out, err := applyAction(action)
	return out, modeAfter, err
}

// findConflictsPreview scans untuk conflicts sebelum processing
func (w *ImportWizard) findConflictsPreview(rows []profilemodel.ImportRow, baseDir string) []profilemodel.ImportRow {
	planned := map[string]bool{}
	conflicts := make([]profilemodel.ImportRow, 0)

	for i := range rows {
		r := rows[i]
		if r.Skip {
			continue
		}
		name := profileconn.TrimProfileSuffix(strings.TrimSpace(r.Name))
		if strings.TrimSpace(name) == "" {
			continue
		}
		absTarget := filepath.Join(baseDir, merger.BuildProfileFileName(name))
		key := strings.ToLower(name)
		if fsops.PathExists(absTarget) || planned[key] {
			r.PlannedName = name
			conflicts = append(conflicts, r)
		}
		planned[key] = true
	}

	sort.Slice(conflicts, func(i, j int) bool { return conflicts[i].RowNum < conflicts[j].RowNum })
	return conflicts
}

// safeName returns nama profile yang aman untuk ditampilkan
func safeName(name string) string {
	s := strings.TrimSpace(name)
	if s == "" {
		return "(unknown)"
	}
	return s
}
