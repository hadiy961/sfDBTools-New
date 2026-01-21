// File : internal/profile/wizard/runner.go
// Deskripsi : Runner wizard interaktif untuk create/edit profile
// Author : Hadiyatna Muflihun
// Tanggal : 4 Januari 2026
// Last Modified : 21 Januari 2026
package wizard

import (
	"fmt"

	profiledisplay "sfdbtools/internal/app/profile/display"
	"sfdbtools/internal/app/profile/helpers/keys"
	profilemodel "sfdbtools/internal/app/profile/model"
	profilevalidation "sfdbtools/internal/app/profile/validation"
	"sfdbtools/internal/domain"
	applog "sfdbtools/internal/services/log"
	"sfdbtools/internal/shared/consts"
	"sfdbtools/internal/shared/validation"
	"sfdbtools/internal/ui/print"
	"sfdbtools/internal/ui/prompt"
)

// Dependency interfaces (ISP-style): wizard adalah consumer.
// Catatan: signature mengikuti kebutuhan saat ini (minim perubahan behavior).
type NameValidator interface {
	CheckNameUnique(mode string) error
}

type SnapshotLoader interface {
	LoadSnapshot(absPath string) (*domain.ProfileInfo, error)
}

type Runner struct {
	Log       applog.Logger
	ConfigDir string
	State     *profilemodel.ProfileState // Shared state pointer, tidak perlu sync

	Validator NameValidator
	Loader    SnapshotLoader
}

// New creates a new Runner instance
func New(log applog.Logger, configDir string, state *profilemodel.ProfileState, validator NameValidator, loader SnapshotLoader) *Runner {
	if log == nil {
		log = applog.NullLogger()
	}
	return &Runner{
		Log:       log,
		ConfigDir: configDir,
		State:     state,
		Validator: validator,
		Loader:    loader,
	}
}

func (r *Runner) Run(mode string) error {
	r.Log.Info(consts.ProfileWizardLogStarted)

	for {
		switch mode {
		case consts.ProfileModeEdit:
			if err := r.runEditFlow(); err != nil {
				return err
			}
		case consts.ProfileModeClone:
			// Clone mode: state sudah pre-fill dari source, tampilkan ringkasan + action menu.
			if err := r.runCloneFlow(); err != nil {
				return err
			}
		default:
			if err := r.runCreateFlow(mode); err != nil {
				return err
			}
		}

		if r.State.ProfileInfo == nil || r.State.ProfileInfo.Name == "" {
			return fmt.Errorf(consts.ProfileErrProfileNameEmpty)
		}

		// Review (clone sudah menampilkan ringkasan di action menu)
		if mode != consts.ProfileModeClone {
			profiledisplay.DisplayProfileDetails(r.ConfigDir, r.State)
		}

		// Edit no-op: jika benar-benar tidak ada perubahan, batalkan tanpa save.
		if mode == consts.ProfileModeEdit && r.State != nil && !r.State.HasMeaningfulChanges() {
			return validation.ErrUserCancelled
		}

		// Comprehensive validation sebelum user confirm save
		// Catch validation errors early, before expensive connection test
		if err := profilevalidation.ValidateProfileInfo(r.State.ProfileInfo); err != nil {
			r.Log.Errorf("Validasi profile gagal: %v", err)
			print.PrintError(fmt.Sprintf("❌ Validasi gagal: %v", err))

			// Tanya user apakah mau retry atau cancel
			confirmRetry, retryErr := prompt.Confirm("Apakah Anda ingin memperbaiki input?", true)
			if retryErr != nil {
				return validation.HandleInputError(retryErr)
			}
			if confirmRetry {
				print.PrintWarning(consts.ProfileWizardMsgRestart)
				continue
			}
			return err
		}

		// Clone: sudah ada pilihan eksplisit "Simpan Clone" di action menu.
		if mode == consts.ProfileModeClone {
			break
		}

		confirmSave, err := prompt.Confirm(consts.ProfilePromptConfirmSave, true)
		if err != nil {
			return validation.HandleInputError(err)
		}
		if confirmSave {
			break
		}

		confirmRetry, err := prompt.Confirm(consts.ProfilePromptConfirmRetry, false)
		if err != nil {
			return validation.HandleInputError(err)
		}
		if confirmRetry {
			print.PrintWarning(consts.ProfileWizardMsgRestart)
			continue
		}
		return validation.ErrUserCancelled
	}

	r.Log.Info(consts.ProfileWizardMsgConfirmAccepted)

	key, source, err := keys.ResolveProfileEncryptionKey(r.State.ProfileInfo.EncryptionKey, true)
	if err != nil {
		return fmt.Errorf(consts.ProfileErrGetEncryptionPasswordFailedFmt, err)
	}
	r.Log.WithField("Sumber Kunci", source).Debug(consts.ProfileLogEncryptionKeyObtained)
	r.State.ProfileInfo.EncryptionKey = key
	return nil
}
