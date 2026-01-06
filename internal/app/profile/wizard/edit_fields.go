// File : internal/profile/wizard/edit_fields.go
// Deskripsi : Prompt edit field secara interaktif (multi-select)
// Author : Hadiyatna Muflihun
// Tanggal : 4 Januari 2026
// Last Modified : 6 Januari 2026

package wizard

import (
	"sfDBTools/internal/ui/print"
	"sfDBTools/internal/ui/prompt"
	"sfDBTools/pkg/consts"
	"sfDBTools/pkg/validation"

	"github.com/AlecAivazis/survey/v2"
)

// promptEditSelectedFields meminta user memilih field yang ingin diubah via multi-select.
func (r *Runner) promptEditSelectedFields() error {
	fields := []string{
		consts.ProfileFieldName,
		consts.ProfileLabelDBHost,
		consts.ProfileLabelDBPort,
		consts.ProfileLabelDBUser,
		consts.ProfileLabelDBPassword,
		consts.ProfileFieldSSHTunnelToggle,
		consts.ProfileLabelSSHHost,
		consts.ProfileLabelSSHPort,
		consts.ProfileLabelSSHUser,
		consts.ProfileLabelSSHPassword,
		consts.ProfileLabelSSHIdentityFile,
		consts.ProfileLabelSSHLocalPort,
	}

	_, idxs, err := prompt.SelectMany(consts.ProfilePromptSelectFieldsToChange, fields, nil)
	if err != nil {
		return validation.HandleInputError(err)
	}
	if len(idxs) == 0 {
		print.PrintWarning(consts.ProfileMsgNoFieldsSelected)
		return validation.ErrUserCancelled
	}

	selected := make(map[string]bool, len(idxs))
	for _, i := range idxs {
		// idxs dari prompt.SelectMany adalah 0-based
		if i >= 0 && i < len(fields) {
			selected[fields[i]] = true
		}
	}

	if selected[consts.ProfileFieldName] {
		if err := r.promptDBConfigName(consts.ProfileModeEdit); err != nil {
			return err
		}
	}

	if selected[consts.ProfileLabelDBHost] {
		v, err := prompt.AskText(consts.ProfileLabelDBHost, prompt.WithDefault(r.ProfileInfo.DBInfo.Host), prompt.WithValidator(survey.Required))
		if err != nil {
			return validation.HandleInputError(err)
		}
		r.ProfileInfo.DBInfo.Host = v
	}

	if selected[consts.ProfileLabelDBPort] {
		v, err := prompt.AskInt(consts.ProfileLabelDBPort, r.ProfileInfo.DBInfo.Port, survey.Required)
		if err != nil {
			return validation.HandleInputError(err)
		}
		r.ProfileInfo.DBInfo.Port = v
	}

	if selected[consts.ProfileLabelDBUser] {
		v, err := prompt.AskText(consts.ProfileLabelDBUser, prompt.WithDefault(r.ProfileInfo.DBInfo.User), prompt.WithValidator(survey.Required))
		if err != nil {
			return validation.HandleInputError(err)
		}
		r.ProfileInfo.DBInfo.User = v
	}

	if selected[consts.ProfileLabelDBPassword] {
		existing := ""
		if r.ProfileInfo != nil {
			existing = r.ProfileInfo.DBInfo.Password
		}
		print.PrintInfo(consts.ProfileTipKeepCurrentDBPassword)
		pw, err := prompt.AskPassword(consts.ProfileLabelDBPassword, nil)
		if err != nil {
			return validation.HandleInputError(err)
		}
		if pw == "" {
			r.ProfileInfo.DBInfo.Password = existing
		} else {
			r.ProfileInfo.DBInfo.Password = pw
		}
	}

	if selected[consts.ProfileFieldSSHTunnelToggle] {
		enable, err := prompt.Confirm(consts.ProfilePromptUseSSHTunnel, r.ProfileInfo.SSHTunnel.Enabled)
		if err != nil {
			return validation.HandleInputError(err)
		}
		r.ProfileInfo.SSHTunnel.Enabled = enable
	}

	sshRequired := survey.Validator(nil)
	if r.ProfileInfo.SSHTunnel.Enabled {
		sshRequired = survey.Required
	}

	if selected[consts.ProfileLabelSSHHost] {
		v, err := prompt.AskText(consts.ProfilePromptSSHHost, prompt.WithDefault(r.ProfileInfo.SSHTunnel.Host), prompt.WithValidator(sshRequired))
		if err != nil {
			return validation.HandleInputError(err)
		}
		r.ProfileInfo.SSHTunnel.Host = v
	}

	if selected[consts.ProfileLabelSSHPort] {
		def := r.ProfileInfo.SSHTunnel.Port
		if def == 0 {
			def = 22
		}
		v, err := prompt.AskInt(consts.ProfileLabelSSHPort, def, sshRequired)
		if err != nil {
			return validation.HandleInputError(err)
		}
		r.ProfileInfo.SSHTunnel.Port = v
	}

	if selected[consts.ProfileLabelSSHUser] {
		v, err := prompt.AskText(consts.ProfilePromptSSHUser, prompt.WithDefault(r.ProfileInfo.SSHTunnel.User))
		if err != nil {
			return validation.HandleInputError(err)
		}
		r.ProfileInfo.SSHTunnel.User = v
	}

	if selected[consts.ProfileLabelSSHPassword] {
		existing := ""
		if r.ProfileInfo != nil {
			existing = r.ProfileInfo.SSHTunnel.Password
		}
		print.PrintInfo(consts.ProfileTipKeepCurrentSSHPassword)
		pw, err := prompt.AskPassword(consts.ProfilePromptSSHPasswordOptional, nil)
		if err != nil {
			return validation.HandleInputError(err)
		}
		if pw == "" {
			r.ProfileInfo.SSHTunnel.Password = existing
		} else {
			r.ProfileInfo.SSHTunnel.Password = pw
		}
	}

	if selected[consts.ProfileLabelSSHIdentityFile] {
		v, err := prompt.AskText(consts.ProfilePromptSSHIdentityFileOptional, prompt.WithDefault(r.ProfileInfo.SSHTunnel.IdentityFile))
		if err != nil {
			return validation.HandleInputError(err)
		}
		r.ProfileInfo.SSHTunnel.IdentityFile = v
	}

	if selected[consts.ProfileLabelSSHLocalPort] {
		v, err := prompt.AskInt(consts.ProfilePromptSSHLocalPort, r.ProfileInfo.SSHTunnel.LocalPort, nil)
		if err != nil {
			return validation.HandleInputError(err)
		}
		r.ProfileInfo.SSHTunnel.LocalPort = v
	}

	return nil
}
