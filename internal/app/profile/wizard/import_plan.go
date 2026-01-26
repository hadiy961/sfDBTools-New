// File : internal/app/profile/wizard/import_plan.go
// Deskripsi : Import planning wizard (conflict + connection test orchestration)
// Author : Hadiyatna Muflihun
// Tanggal : 26 Januari 2026
// Last Modified : 26 Januari 2026

package wizard

import (
	"fmt"
	"path/filepath"
	"strings"

	profileconn "sfdbtools/internal/app/profile/connection"
	"sfdbtools/internal/app/profile/helpers/importer"
	"sfdbtools/internal/app/profile/merger"
	profilemodel "sfdbtools/internal/app/profile/model"
	"sfdbtools/internal/shared/fsops"
)

// resolveAndPlan handles conflict resolution dan connection testing
// Returns planned rows siap untuk disave
func (w *ImportWizard) resolveAndPlan(rows []profilemodel.ImportRow, opts *profilemodel.ProfileImportOptions) ([]profilemodel.ImportRow, error) {
	baseDir := w.ConfigDir
	if strings.TrimSpace(baseDir) == "" {
		return nil, fmt.Errorf("config dir profile kosong")
	}

	onConflict := strings.ToLower(strings.TrimSpace(opts.OnConflict))
	if onConflict == "" {
		onConflict = profilemodel.ImportConflictFail
	}

	w.Log.Infof("[profile-import] Pre-save checks (3/3): conflict + conn-test | on_conflict=%s | skip_conn_test=%v",
		onConflict, opts.SkipConnTest)

	// Global conflict strategy prompt (interactive mode)
	conflictMode, err := w.handleConflicts(rows, baseDir, opts)
	if err != nil {
		return nil, err
	}

	// Global connection test prompt (interactive mode)
	if err := w.promptConnectionTest(opts); err != nil {
		return nil, err
	}

	// Process each row
	plannedNames := map[string]bool{}
	planned := make([]profilemodel.ImportRow, 0, len(rows))
	connTestTotal := importer.CountPlanned(rows)
	connTestIndex := 0

	for i := range rows {
		r := rows[i]

		// Skip already marked rows
		if r.Skip {
			r.PlanAction = profilemodel.ImportPlanSkip
			if strings.TrimSpace(r.SkipReason) == "" {
				r.SkipReason = profilemodel.ImportSkipReasonUnknown
			}
			planned = append(planned, r)
			continue
		}

		// Normalize name
		name := profileconn.TrimProfileSuffix(strings.TrimSpace(r.Name))
		r.PlannedName = name
		r.PlanAction = profilemodel.ImportPlanCreate

		// Check conflict
		absTarget := filepath.Join(baseDir, merger.BuildProfileFileName(r.PlannedName))
		conflict := fsops.PathExists(absTarget) || plannedNames[strings.ToLower(r.PlannedName)]
		if conflict {
			resolved, mode, err := w.resolveConflictForRow(r, baseDir, plannedNames, conflictMode, opts)
			if err != nil {
				return nil, err
			}
			conflictMode = mode // Update mode jika user pilih "apply to all"
			r = resolved
			if r.Skip {
				planned = append(planned, r)
				continue
			}
		}

		plannedNames[strings.ToLower(r.PlannedName)] = true

		// Connection test (skip jika SkipConnTest=true)
		if !opts.SkipConnTest {
			connTestIndex++
			tested, err := w.testConnection(r, connTestIndex, connTestTotal, opts)
			if err != nil {
				return nil, err
			}
			if tested.Skip {
				planned = append(planned, tested)
				continue
			}
			r = tested
		}

		planned = append(planned, r)
	}

	return planned, nil
}
