// File : internal/profile/wizard/edit_fields.go
// Deskripsi : Prompt edit field secara interaktif (multi-select)
// Author : Hadiyatna Muflihun
// Tanggal : 4 Januari 2026
// Last Modified : 5 Januari 2026

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

	if selected[consts.ProfileLabelDBHost] {
		v, err := input.AskString(consts.ProfileLabelDBHost, r.ProfileInfo.DBInfo.Host, survey.Required)
		if err != nil {
			return validation.HandleInputError(err)
		}
		r.ProfileInfo.DBInfo.Host = v
	}

	if selected[consts.ProfileLabelDBPort] {
		v, err := input.AskInt(consts.ProfileLabelDBPort, r.ProfileInfo.DBInfo.Port, survey.Required)
		if err != nil {
			return validation.HandleInputError(err)
		}
		r.ProfileInfo.DBInfo.Port = v
	}

	if selected[consts.ProfileLabelDBUser] {
		v, err := input.AskString(consts.ProfileLabelDBUser, r.ProfileInfo.DBInfo.User, survey.Required)
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
		ui.PrintInfo(consts.ProfileTipKeepCurrentDBPassword)
		pw, err := input.AskPassword(consts.ProfileLabelDBPassword, nil)
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

	if selected[consts.ProfileLabelSSHHost] {
		v, err := input.AskString(consts.ProfilePromptSSHHost, r.ProfileInfo.SSHTunnel.Host, sshRequired)
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
		v, err := input.AskInt(consts.ProfileLabelSSHPort, def, sshRequired)
		if err != nil {
			return validation.HandleInputError(err)
		}
		r.ProfileInfo.SSHTunnel.Port = v
	}

	if selected[consts.ProfileLabelSSHUser] {
		v, err := input.AskString(consts.ProfilePromptSSHUser, r.ProfileInfo.SSHTunnel.User, nil)
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

	if selected[consts.ProfileLabelSSHIdentityFile] {
		v, err := input.AskString(consts.ProfilePromptSSHIdentityFileOptional, r.ProfileInfo.SSHTunnel.IdentityFile, nil)
		if err != nil {
			return validation.HandleInputError(err)
		}
		r.ProfileInfo.SSHTunnel.IdentityFile = v
	}

	if selected[consts.ProfileLabelSSHLocalPort] {
		v, err := input.AskInt(consts.ProfilePromptSSHLocalPort, r.ProfileInfo.SSHTunnel.LocalPort, nil)
		if err != nil {
			return validation.HandleInputError(err)
		}
		r.ProfileInfo.SSHTunnel.LocalPort = v
	}

	return nil
}
