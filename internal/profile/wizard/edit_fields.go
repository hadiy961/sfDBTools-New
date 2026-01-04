// File : internal/profile/wizard/edit_fields.go
// Deskripsi : Prompt edit field secara interaktif (multi-select)
// Author : Hadiyatna Muflihun
// Tanggal : 4 Januari 2026
// Last Modified : 4 Januari 2026

package wizard

import (
	"sfDBTools/pkg/consts"
	"sfDBTools/pkg/input"
	"sfDBTools/pkg/ui"
	"sfDBTools/pkg/validation"

	"github.com/AlecAivazis/survey/v2"
)

func (r *Runner) promptEditSelectedFields() error {
	fields := []string{
		consts.ProfileFieldName,
		consts.ProfileFieldDBHost,
		consts.ProfileFieldDBPort,
		consts.ProfileFieldDBUser,
		consts.ProfileFieldDBPassword,
		consts.ProfileFieldSSHTunnelToggle,
		consts.ProfileFieldSSHHost,
		consts.ProfileFieldSSHPort,
		consts.ProfileFieldSSHUser,
		consts.ProfileFieldSSHPassword,
		consts.ProfileFieldSSHIdentityFile,
		consts.ProfileFieldSSHLocalPort,
	}

	idxs, err := input.ShowMultiSelect(consts.ProfilePromptSelectFieldsToChange, fields)
	if err != nil {
		return validation.HandleInputError(err)
	}
	if len(idxs) == 0 {
		ui.PrintWarning(consts.ProfileMsgNoFieldsSelected)
		return validation.ErrUserCancelled
	}

	selected := make(map[string]bool, len(idxs))
	for _, i := range idxs {
		if i >= 1 && i <= len(fields) {
			selected[fields[i-1]] = true
		}
	}

	if selected[consts.ProfileFieldName] {
		if err := r.promptDBConfigName(consts.ProfileModeEdit); err != nil {
			return err
		}
	}

	if selected[consts.ProfileFieldDBHost] {
		v, err := input.AskString(consts.ProfilePromptDBHost, r.ProfileInfo.DBInfo.Host, survey.Required)
		if err != nil {
			return validation.HandleInputError(err)
		}
		r.ProfileInfo.DBInfo.Host = v
	}

	if selected[consts.ProfileFieldDBPort] {
		v, err := input.AskInt(consts.ProfilePromptDBPort, r.ProfileInfo.DBInfo.Port, survey.Required)
		if err != nil {
			return validation.HandleInputError(err)
		}
		r.ProfileInfo.DBInfo.Port = v
	}

	if selected[consts.ProfileFieldDBUser] {
		v, err := input.AskString(consts.ProfilePromptDBUser, r.ProfileInfo.DBInfo.User, survey.Required)
		if err != nil {
			return validation.HandleInputError(err)
		}
		r.ProfileInfo.DBInfo.User = v
	}

	if selected[consts.ProfileFieldDBPassword] {
		existing := ""
		if r.ProfileInfo != nil {
			existing = r.ProfileInfo.DBInfo.Password
		}
		ui.PrintInfo(consts.ProfileTipKeepCurrentDBPassword)
		pw, err := input.AskPassword(consts.ProfilePromptDBPassword, nil)
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
		enable, err := input.AskYesNo(consts.ProfilePromptUseSSHTunnel, r.ProfileInfo.SSHTunnel.Enabled)
		if err != nil {
			return validation.HandleInputError(err)
		}
		r.ProfileInfo.SSHTunnel.Enabled = enable
	}

	sshRequired := survey.Validator(nil)
	if r.ProfileInfo.SSHTunnel.Enabled {
		sshRequired = survey.Required
	}

	if selected[consts.ProfileFieldSSHHost] {
		v, err := input.AskString(consts.ProfilePromptSSHHost, r.ProfileInfo.SSHTunnel.Host, sshRequired)
		if err != nil {
			return validation.HandleInputError(err)
		}
		r.ProfileInfo.SSHTunnel.Host = v
	}

	if selected[consts.ProfileFieldSSHPort] {
		def := r.ProfileInfo.SSHTunnel.Port
		if def == 0 {
			def = 22
		}
		v, err := input.AskInt(consts.ProfilePromptSSHPort, def, sshRequired)
		if err != nil {
			return validation.HandleInputError(err)
		}
		r.ProfileInfo.SSHTunnel.Port = v
	}

	if selected[consts.ProfileFieldSSHUser] {
		v, err := input.AskString(consts.ProfilePromptSSHUser, r.ProfileInfo.SSHTunnel.User, nil)
		if err != nil {
			return validation.HandleInputError(err)
		}
		r.ProfileInfo.SSHTunnel.User = v
	}

	if selected[consts.ProfileFieldSSHPassword] {
		existing := ""
		if r.ProfileInfo != nil {
			existing = r.ProfileInfo.SSHTunnel.Password
		}
		ui.PrintInfo(consts.ProfileTipKeepCurrentSSHPassword)
		pw, err := input.AskPassword(consts.ProfilePromptSSHPasswordOptional, nil)
		if err != nil {
			return validation.HandleInputError(err)
		}
		if pw == "" {
			r.ProfileInfo.SSHTunnel.Password = existing
		} else {
			r.ProfileInfo.SSHTunnel.Password = pw
		}
	}

	if selected[consts.ProfileFieldSSHIdentityFile] {
		v, err := input.AskString(consts.ProfilePromptSSHIdentityFileOptional, r.ProfileInfo.SSHTunnel.IdentityFile, nil)
		if err != nil {
			return validation.HandleInputError(err)
		}
		r.ProfileInfo.SSHTunnel.IdentityFile = v
	}

	if selected[consts.ProfileFieldSSHLocalPort] {
		v, err := input.AskInt(consts.ProfilePromptSSHLocalPort, r.ProfileInfo.SSHTunnel.LocalPort, nil)
		if err != nil {
			return validation.HandleInputError(err)
		}
		r.ProfileInfo.SSHTunnel.LocalPort = v
	}

	return nil
}
