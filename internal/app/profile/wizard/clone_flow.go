// File : internal/app/profile/wizard/clone_flow.go
// Deskripsi : Flow wizard untuk clone profile (edit-like: tampilkan profil + action menu)
// Author : Hadiyatna Muflihun
// Tanggal : 21 Januari 2026
// Last Modified : 21 Januari 2026

package wizard

import (
	"fmt"

	profiledisplay "sfdbtools/internal/app/profile/display"
	profilemodel "sfdbtools/internal/app/profile/model"
	"sfdbtools/internal/shared/consts"
	"sfdbtools/internal/shared/validation"
	"sfdbtools/internal/ui/prompt"
)

func (r *Runner) runCloneFlow() error {
	if r == nil || r.State == nil || r.State.ProfileInfo == nil {
		return fmt.Errorf("state clone tidak tersedia")
	}

	// Flow clone mirip edit:
	// 1) tampilkan profil (target yang sudah pre-fill)
	// 2) aksi: Ubah data / Simpan Clone / Batalkan
	// 3) jika Ubah data, multi-select fields, lalu kembali ke menu aksi
	for {
		prevOpts := r.State.Options
		r.State.Options = &profilemodel.ProfileShowOptions{}
		profiledisplay.DisplayProfileDetails(r.ConfigDir, r.State)
		r.State.Options = prevOpts

		action, _, err := prompt.SelectOne(
			consts.ProfilePromptAction,
			[]string{consts.ProfileActionEditData, consts.ProfileActionSaveClone, consts.ProfileActionCancel},
			-1,
		)
		if err != nil {
			return validation.HandleInputError(err)
		}

		switch action {
		case consts.ProfileActionCancel:
			return validation.ErrUserCancelled
		case consts.ProfileActionSaveClone:
			return nil
		case consts.ProfileActionEditData:
			if err := r.promptCloneSelectedFields(); err != nil {
				return err
			}
			if err := r.ensureSSHTunnelMinimumIfEnabled(); err != nil {
				return err
			}
			continue
		default:
			return fmt.Errorf("aksi tidak dikenali: %s", action)
		}
	}
}
