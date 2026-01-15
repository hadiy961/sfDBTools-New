// File : internal/app/profile/wizard/edit_flow.go
// Deskripsi : Flow wizard untuk edit profile (honor flag overrides)
// Author : Hadiyatna Muflihun
// Tanggal : 4 Januari 2026
// Last Modified : 15 Januari 2026
package wizard

import (
	"fmt"
	"strings"

	profiledisplay "sfdbtools/internal/app/profile/display"
	profilehelper "sfdbtools/internal/app/profile/helpers"
	"sfdbtools/internal/app/profile/merger"
	profilemodel "sfdbtools/internal/app/profile/model"
	"sfdbtools/internal/domain"
	"sfdbtools/internal/shared/consts"
	"sfdbtools/internal/shared/fsops"
	"sfdbtools/internal/shared/validation"
	"sfdbtools/internal/ui/prompt"
)

func (r *Runner) runEditFlow() error {
	// Simpan override dari flag/env sebelum load snapshot.
	overrideDB := domain.DBInfo{}
	overrideSSH := domain.SSHTunnelConfig{}
	if r.State.ProfileInfo != nil {
		overrideDB = r.State.ProfileInfo.DBInfo
		overrideSSH = r.State.ProfileInfo.SSHTunnel
	}

	// Jika user memberikan --profile, jangan minta pilih lagi.
	target := ""
	if r.State.ProfileInfo != nil {
		target = strings.TrimSpace(r.State.ProfileInfo.Path)
	}
	if target == "" {
		target = strings.TrimSpace(r.State.OriginalProfileName)
	}

	if target != "" {
		absPath, name, err := profilehelper.ResolveConfigPathInDir(r.ConfigDir, target)
		if err != nil {
			return err
		}
		if !fsops.PathExists(absPath) {
			return fmt.Errorf(consts.ProfileErrConfigFileNotFoundFmt, absPath)
		}
		if r.Loader == nil {
			return fmt.Errorf(consts.ProfileErrLoadSnapshotUnavailable)
		}
		snap, err := r.Loader.LoadSnapshot(absPath)
		if err != nil {
			return err
		}
		r.State.OriginalProfileInfo = snap
		if r.State.ProfileInfo != nil {
			r.State.ProfileInfo.Path = absPath
			r.State.ProfileInfo.Name = name
		}
		r.State.OriginalProfileName = name
	} else {
		// Tidak ada --profile: pilih secara interaktif.
		if err := r.promptSelectExistingConfig(); err != nil {
			return err
		}
	}

	// Jadikan snapshot sebagai baseline, lalu apply override dari flag/env.
	merger.ApplySnapshotAsBaseline(r.State.ProfileInfo, r.State.OriginalProfileInfo)
	merger.ApplyDBOverrides(r.State.ProfileInfo, overrideDB)
	merger.ApplySSHOverrides(r.State.ProfileInfo, overrideSSH)

	// Tampilkan isi profil terlebih dahulu (seperti profile show), lalu beri opsi ubah/batal.
	// Ini tetap dijalankan walaupun ada override flag/env, supaya user selalu lihat kondisi awal.
	prevOpts := r.State.Options
	r.State.Options = &profilemodel.ProfileShowOptions{}
	profiledisplay.DisplayProfileDetails(r.ConfigDir, r.State)
	r.State.Options = prevOpts

	action, _, err := prompt.SelectOne(consts.ProfilePromptAction, []string{consts.ProfileActionEditData, consts.ProfileActionCancel}, -1)
	if err != nil {
		return validation.HandleInputError(err)
	}
	if action == consts.ProfileActionCancel {
		return validation.ErrUserCancelled
	}

	// Setelah user memilih "Ubah data", langsung minta multi-select field yang ingin diubah.
	if err := r.promptEditSelectedFields(); err != nil {
		return err
	}

	// Jika SSH tunnel aktif, pastikan minimal field penting sudah terisi tanpa memaksa input opsional.
	if err := r.ensureSSHTunnelMinimumIfEnabled(); err != nil {
		return err
	}
	return nil
}

// ensureSSHTunnelMinimumIfEnabled memastikan kebutuhan minimal SSH tunnel pada edit flow.
// Pada edit flow, kita hindari prompt untuk field opsional (identity/local port/password) kecuali user memilihnya manual.
// Catatan: SSH Host wajib jika tunnel aktif, dan SSH Port akan di-default ke 22 jika kosong.
func (r *Runner) ensureSSHTunnelMinimumIfEnabled() error {
	if r.State.ProfileInfo == nil {
		return nil
	}
	if !r.State.ProfileInfo.SSHTunnel.Enabled {
		return nil
	}
	// Host wajib jika tunnel aktif.
	if strings.TrimSpace(r.State.ProfileInfo.SSHTunnel.Host) == "" {
		validator := prompt.ComposeValidators(
			validateNotBlank(consts.ProfileLabelSSHHost),
			validateNoControlChars(consts.ProfileLabelSSHHost),
			validateNoLeadingTrailingSpaces(consts.ProfileLabelSSHHost),
			validateNoSpaces(consts.ProfileLabelSSHHost),
		)
		v, err := prompt.AskText(consts.ProfilePromptSSHHost, prompt.WithDefault(""), prompt.WithValidator(validator))
		if err != nil {
			return validation.HandleInputError(err)
		}
		r.State.ProfileInfo.SSHTunnel.Host = strings.TrimSpace(v)
	}
	// Port default 22 untuk menghindari prompt tambahan.
	if r.State.ProfileInfo.SSHTunnel.Port == 0 {
		r.State.ProfileInfo.SSHTunnel.Port = 22
	}
	return nil
}
