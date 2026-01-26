// File : internal/app/profile/helpers/resolver/conflict_resolver.go
// Deskripsi : Conflict resolution helpers untuk import (auto-rename, dll)
// Author : Hadiyatna Muflihun
// Tanggal : 25 Januari 2026
// Last Modified : 25 Januari 2026

package resolver

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	profileconn "sfdbtools/internal/app/profile/connection"
	"sfdbtools/internal/app/profile/merger"
	profilemodel "sfdbtools/internal/app/profile/model"
	applog "sfdbtools/internal/services/log"
	"sfdbtools/internal/shared/fsops"
	sharedvalidation "sfdbtools/internal/shared/validation"
	"sfdbtools/internal/ui/prompt"
)

// AutoRename generates new unique name untuk profile yang konflik
// Mencoba suffix _2, _3, dst. hingga menemukan nama yang belum ada
func AutoRename(baseDir string, name string, planned map[string]bool) string {
	base := profileconn.TrimProfileSuffix(strings.TrimSpace(name))
	if base == "" {
		base = "profile"
	}

	// Start from _2
	for i := 2; i < 10000; i++ {
		candidate := fmt.Sprintf("%s_%d", base, i)
		key := strings.ToLower(candidate)
		if planned[key] {
			continue
		}
		abs := filepath.Join(baseDir, merger.BuildProfileFileName(candidate))
		if fsops.PathExists(abs) {
			continue
		}
		return candidate
	}

	// Fallback: timestamp suffix jika loop habis
	return fmt.Sprintf("%s_%d", base, time.Now().Unix())
}

// ResolveConflict menangani konflik nama profile (interactive atau auto)
func ResolveConflict(log applog.Logger, r *profilemodel.ImportRow, opts *profilemodel.ProfileImportOptions, onConflict, baseDir string, planned map[string]bool) (profilemodel.ImportRow, error) {
	safeName := func(name string) string {
		v := strings.TrimSpace(name)
		if v == "" {
			return "(unknown)"
		}
		return v
	}

	log.Infof("[profile-import] Konflik terdeteksi untuk '%s' (row %d)", safeName(r.PlannedName), r.RowNum)

	var action string
	if !opts.SkipConfirm {
		choice, _, err := prompt.SelectOne(
			fmt.Sprintf("Profile '%s' (row %d) sudah ada. Pilih aksi:", safeName(r.PlannedName), r.RowNum),
			[]string{"Rename", "Overwrite", "Skip", "Batalkan"},
			0,
		)
		if err != nil {
			return *r, sharedvalidation.HandleInputError(err)
		}
		switch choice {
		case "Rename":
			action = profilemodel.ImportConflictRename
		case "Overwrite":
			action = profilemodel.ImportConflictOverwrite
		case "Skip":
			action = profilemodel.ImportConflictSkip
		default:
			return *r, sharedvalidation.ErrUserCancelled
		}
	} else {
		action = onConflict
	}

	log.Infof("[profile-import] Aksi konflik '%s' (row %d): %s", safeName(r.PlannedName), r.RowNum, action)

	switch action {
	case profilemodel.ImportConflictFail:
		return *r, fmt.Errorf("konflik file untuk profile '%s' (row %d)", safeName(r.PlannedName), r.RowNum)
	case profilemodel.ImportConflictSkip:
		r.Skip = true
		r.SkipReason = profilemodel.ImportSkipReasonConflict
		r.PlanAction = profilemodel.ImportPlanSkip
	case profilemodel.ImportConflictOverwrite:
		r.PlanAction = profilemodel.ImportPlanOverwrite
	case profilemodel.ImportConflictRename:
		r.PlanAction = profilemodel.ImportPlanRename
		r.RenamedFrom = r.PlannedName
		if !opts.SkipConfirm {
			newName, err := prompt.AskText("Masukkan nama profile baru:")
			if err != nil {
				return *r, sharedvalidation.HandleInputError(err)
			}
			newName = profileconn.TrimProfileSuffix(strings.TrimSpace(newName))
			if err := sharedvalidation.ValidateProfileName(newName); err != nil {
				return *r, err
			}
			// Re-check conflict
			abs2 := filepath.Join(baseDir, merger.BuildProfileFileName(newName))
			if fsops.PathExists(abs2) || planned[strings.ToLower(newName)] {
				return *r, fmt.Errorf("nama profile '%s' masih konflik (row %d)", safeName(newName), r.RowNum)
			}
			r.PlannedName = newName
		} else {
			r.PlannedName = AutoRename(baseDir, r.PlannedName, planned)
			log.Infof("[profile-import] Auto-rename row %d: %s -> %s", r.RowNum, safeName(r.RenamedFrom), safeName(r.PlannedName))
		}
	}

	return *r, nil
}
