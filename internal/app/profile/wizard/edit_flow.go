// File : internal/app/profile/wizard/edit_flow.go
// Deskripsi : Flow wizard untuk edit profile (honor flag overrides)
// Author : Hadiyatna Muflihun
// Tanggal : 4 Januari 2026
// Last Modified : 6 Januari 2026
package wizard

import (
	"fmt"
	"strings"

	profilemodel "sfdbtools/internal/app/profile/model"
	"sfdbtools/internal/app/profile/shared"
	"sfdbtools/internal/domain"
	"sfdbtools/internal/ui/print"
	"sfdbtools/internal/ui/prompt"
	"sfdbtools/pkg/consts"
	"sfdbtools/pkg/fsops"
	"sfdbtools/pkg/helper"
	"sfdbtools/pkg/validation"

	"github.com/AlecAivazis/survey/v2"
)

func (r *Runner) runEditFlow() error {
	// Simpan override dari flag/env sebelum load snapshot.
	overrideDB := domain.DBInfo{}
	overrideSSH := domain.SSHTunnelConfig{}
	if r.ProfileInfo != nil {
		overrideDB = r.ProfileInfo.DBInfo
		overrideSSH = r.ProfileInfo.SSHTunnel
	}

	// Jika user memberikan --profile, jangan minta pilih lagi.
	target := ""
	if r.ProfileInfo != nil {
		target = strings.TrimSpace(r.ProfileInfo.Path)
	}
	if target == "" {
		target = strings.TrimSpace(r.OriginalProfileName)
	}

	if target != "" {
		absPath, name, err := helper.ResolveConfigPath(target)
		if err != nil {
			return err
		}
		if !fsops.PathExists(absPath) {
			return fmt.Errorf(consts.ProfileErrConfigFileNotFoundFmt, absPath)
		}
		if r.LoadSnapshotFromPath == nil {
			return fmt.Errorf(consts.ProfileErrLoadSnapshotUnavailable)
		}
		snap, err := r.LoadSnapshotFromPath(absPath)
		if err != nil {
			return err
		}
		r.OriginalProfileInfo = snap
		if r.ProfileInfo != nil {
			r.ProfileInfo.Path = absPath
			r.ProfileInfo.Name = name
		}
		r.OriginalProfileName = name
	} else {
		// Tidak ada --profile: pilih secara interaktif.
		if err := r.promptSelectExistingConfig(); err != nil {
			return err
		}
	}

	// Jadikan snapshot sebagai baseline, lalu apply override dari flag/env.
	shared.ApplySnapshotAsBaseline(r.ProfileInfo, r.OriginalProfileInfo)
	shared.ApplyDBOverrides(r.ProfileInfo, overrideDB)
	shared.ApplySSHOverrides(r.ProfileInfo, overrideSSH)

	// Tampilkan isi profil terlebih dahulu (seperti profile show), lalu beri opsi ubah/batal.
	// Ini tetap dijalankan walaupun ada override flag/env, supaya user selalu lihat kondisi awal.
	prevShow := r.ProfileShow
	r.ProfileShow = &profilemodel.ProfileShowOptions{}
	if r.DisplayProfileDetails != nil {
		r.DisplayProfileDetails()
	}
	r.ProfileShow = prevShow

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

	print.PrintSubHeader(consts.ProfileMsgChangeSummaryPrefix + r.ProfileInfo.Name)
	return nil
}

// ensureSSHTunnelMinimumIfEnabled memastikan kebutuhan minimal SSH tunnel pada edit flow.
// Pada edit flow, kita hindari prompt untuk field opsional (identity/local port/password) kecuali user memilihnya manual.
// Catatan: SSH Host wajib jika tunnel aktif, dan SSH Port akan di-default ke 22 jika kosong.
func (r *Runner) ensureSSHTunnelMinimumIfEnabled() error {
	if r.ProfileInfo == nil {
		return nil
	}
	if !r.ProfileInfo.SSHTunnel.Enabled {
		return nil
	}
	// Host wajib jika tunnel aktif.
	if strings.TrimSpace(r.ProfileInfo.SSHTunnel.Host) == "" {
		v, err := prompt.AskText(consts.ProfilePromptSSHHost, prompt.WithDefault(""), prompt.WithValidator(survey.Required))
		if err != nil {
			return validation.HandleInputError(err)
		}
		r.ProfileInfo.SSHTunnel.Host = v
	}
	// Port default 22 untuk menghindari prompt tambahan.
	if r.ProfileInfo.SSHTunnel.Port == 0 {
		r.ProfileInfo.SSHTunnel.Port = 22
	}
	return nil
}
