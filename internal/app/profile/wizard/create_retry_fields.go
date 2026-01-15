// File : internal/app/profile/wizard/create_retry_fields.go
// Deskripsi : Prompt retry create profile (pilih field yang ingin diubah)
// Author : Hadiyatna Muflihun
// Tanggal : 9 Januari 2026
// Last Modified : 9 Januari 2026

package wizard

import (
	"strings"

	profileerrors "sfdbtools/internal/app/profile/errors"
	"sfdbtools/internal/shared/consts"
	"sfdbtools/internal/shared/validation"
	"sfdbtools/internal/ui/prompt"
)

// PromptCreateRetrySelectedFields menampilkan opsi seperti profile edit saat user memilih mengulang input
// setelah koneksi database gagal ketika proses save (mode create).
//
// UX requirement:
// - Tampilkan multi-select field (7 item) seperti screenshot: nama, kunci enkripsi, DB host/port/user/password, dan toggle SSH.
// - Hanya re-prompt field yang dipilih.
func (r *Runner) PromptCreateRetrySelectedFields() error {
	if r.State.ProfileInfo == nil {
		return profileerrors.ErrProfileNil
	}

	fields := []string{
		consts.ProfileFieldName,
		consts.ProfileFieldEncryptionKey,
		consts.ProfileLabelDBHost,
		consts.ProfileLabelDBPort,
		consts.ProfileLabelDBUser,
		consts.ProfileLabelDBPassword,
		consts.ProfileFieldSSHTunnelToggle,
	}

	selected, err := selectManyFieldsOrCancel(consts.ProfilePromptSelectFieldsToChange, fields)
	if err != nil {
		return err
	}

	if selected[consts.ProfileFieldName] {
		if err := r.promptDBConfigName(consts.ProfileModeCreate); err != nil {
			return err
		}
	}

	if selected[consts.ProfileFieldEncryptionKey] {
		newKey, err := promptNewEncryptionKeyConfirmed()
		if err != nil {
			return err
		}
		r.State.ProfileInfo.EncryptionKey = strings.TrimSpace(newKey)
		r.State.ProfileInfo.EncryptionSource = "prompt"
	}

	if selected[consts.ProfileLabelDBHost] {
		def := strings.TrimSpace(r.State.ProfileInfo.DBInfo.Host)
		if def == "" {
			def = "localhost"
		}
		if err := r.promptDBHostRequired(def); err != nil {
			return err
		}
	}

	if selected[consts.ProfileLabelDBPort] {
		def := r.State.ProfileInfo.DBInfo.Port
		if def == 0 {
			def = 3306
		}
		if err := r.promptDBPortRequired(def); err != nil {
			return err
		}
	}

	if selected[consts.ProfileLabelDBUser] {
		def := strings.TrimSpace(r.State.ProfileInfo.DBInfo.User)
		if def == "" {
			def = "root"
		}
		if err := r.promptDBUserRequired(def); err != nil {
			return err
		}
	}

	if selected[consts.ProfileLabelDBPassword] {
		existing := ""
		if r.State.ProfileInfo != nil {
			existing = r.State.ProfileInfo.DBInfo.Password
		}
		if err := r.promptDBPasswordKeepCurrent(existing); err != nil {
			return err
		}
	}

	if selected[consts.ProfileFieldSSHTunnelToggle] {
		enable, err := prompt.Confirm(consts.ProfilePromptUseSSHTunnel, r.State.ProfileInfo.SSHTunnel.Enabled)
		if err != nil {
			return validation.HandleInputError(err)
		}
		r.State.ProfileInfo.SSHTunnel.Enabled = enable
		if enable {
			// Jika user mengaktifkan SSH tunnel, prompt minimal field yang diperlukan.
			if err := r.promptSSHTunnelDetailsIfEnabled(); err != nil {
				return err
			}
		}
	}

	return nil
}
