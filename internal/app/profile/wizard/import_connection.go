// File : internal/app/profile/wizard/import_connection.go
// Deskripsi : Import connection test wizard
// Author : Hadiyatna Muflihun
// Tanggal : 26 Januari 2026
// Last Modified : 26 Januari 2026

package wizard

import (
	"fmt"
	"strings"

	profileconn "sfdbtools/internal/app/profile/connection"
	"sfdbtools/internal/app/profile/helpers/importer"
	profilemodel "sfdbtools/internal/app/profile/model"
	"sfdbtools/internal/shared/consts"
	sharedvalidation "sfdbtools/internal/shared/validation"
	"sfdbtools/internal/ui/prompt"
)

// promptConnectionTest prompts user untuk connection test (interactive mode only)
func (w *ImportWizard) promptConnectionTest(opts *profilemodel.ProfileImportOptions) error {
	if opts.SkipConfirm || opts.SkipConnTest {
		return nil
	}

	ok, askErr := prompt.Confirm("Tes koneksi sekarang?", true)
	if askErr != nil {
		return sharedvalidation.HandleInputError(askErr)
	}
	if !ok {
		opts.SkipConnTest = true
		w.Log.Warn("[profile-import] Conn-test di-skip untuk run ini (user memilih tidak tes sekarang)")
	}
	return nil
}

// testConnection melakukan connection test untuk satu row
// Handles interactive prompt saat conn-test gagal
func (w *ImportWizard) testConnection(
	r profilemodel.ImportRow,
	idx, total int,
	opts *profilemodel.ProfileImportOptions,
) (profilemodel.ImportRow, error) {
	w.Log.Infof("[profile-import] %d/%d tes koneksi '%s' (row %d) host=%s",
		idx, total, importer.SafeName(r.PlannedName), r.RowNum, strings.TrimSpace(r.Host))

	info := importer.BuildProfileInfoFromRow(r)
	conn, err := profileconn.ConnectWithProfile(w.Config, &info, consts.DefaultInitialDatabase)
	if err != nil {
		w.Log.Warnf("[profile-import] %d/%d koneksi gagal '%s' (row %d): %v",
			idx, total, importer.SafeName(r.PlannedName), r.RowNum, err)

		// Automation mode: skip jika --continue-on-error, error jika tidak
		if opts.SkipConfirm {
			if opts.ContinueOnError {
				w.Log.Warnf("Conn-test gagal untuk '%s' (row %d); skip karena --continue-on-error",
					importer.SafeName(r.PlannedName), r.RowNum)
				r.Skip = true
				r.SkipReason = profilemodel.ImportSkipReasonConnTest
				r.PlanAction = profilemodel.ImportPlanSkip
				return r, nil
			}
			return r, fmt.Errorf("conn-test gagal untuk profile '%s' (row %d): %w",
				importer.SafeName(r.PlannedName), r.RowNum, err)
		}

		// Interactive mode: tanya user
		choice, _, askErr := prompt.SelectOne(
			fmt.Sprintf("Conn-test gagal untuk '%s' (row %d). Pilih aksi:", importer.SafeName(r.PlannedName), r.RowNum),
			[]string{"Tetap simpan", "Skip", "Batalkan"},
			0,
		)
		if askErr != nil {
			return r, sharedvalidation.HandleInputError(askErr)
		}

		switch choice {
		case "Tetap simpan":
			w.Log.Infof("[profile-import] Hasil row %d: tetap simpan meski conn-test gagal", r.RowNum)
		case "Skip":
			r.Skip = true
			r.SkipReason = profilemodel.ImportSkipReasonConnTest
			r.PlanAction = profilemodel.ImportPlanSkip
			w.Log.Infof("[profile-import] Hasil row %d: skip (%s)", r.RowNum, profilemodel.ImportSkipReasonConnTest)
		default:
			return r, sharedvalidation.ErrUserCancelled
		}
	} else {
		conn.Close()
		w.Log.Infof("[profile-import] %d/%d koneksi OK '%s' (row %d)",
			idx, total, importer.SafeName(r.PlannedName), r.RowNum)
	}

	return r, nil
}
