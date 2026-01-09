// File : internal/profile/wizard/edit_fields.go
// Deskripsi : Prompt edit field secara interaktif (multi-select)
// Author : Hadiyatna Muflihun
// Tanggal : 4 Januari 2026
// Last Modified : 9 Januari 2026

package wizard

import (
	"strings"

	"sfdbtools/internal/shared/consts"
	"sfdbtools/internal/shared/validation"
	"sfdbtools/internal/ui/print"
	"sfdbtools/internal/ui/prompt"

	"github.com/AlecAivazis/survey/v2"
)

// promptEditSelectedFields meminta user memilih field yang ingin diubah via multi-select.
func (r *Runner) promptEditSelectedFields() error {
	fields := []string{
		consts.ProfileFieldName,
		consts.ProfileFieldEncryptionKey,
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

	selected, err := selectManyFieldsOrCancel(consts.ProfilePromptSelectFieldsToChange, fields)
	if err != nil {
		return err
	}

	if selected[consts.ProfileFieldName] {
		if err := r.promptDBConfigName(consts.ProfileModeEdit); err != nil {
			return err
		}
	}

	if selected[consts.ProfileFieldEncryptionKey] {
		// Rotasi encryption key untuk file profil (decrypt tetap pakai key lama yang sudah dipakai saat load snapshot).
		newKey, err := promptNewEncryptionKeyConfirmed()
		if err != nil {
			return err
		}
		if r.ProfileEdit != nil {
			r.ProfileEdit.NewProfileKey = strings.TrimSpace(newKey)
			r.ProfileEdit.NewProfileKeySource = "prompt"
		}
	}

	if selected[consts.ProfileLabelDBHost] {
		def := strings.TrimSpace(r.ProfileInfo.DBInfo.Host)
		if def == "" {
			def = "localhost"
		}
		if err := r.promptDBHostRequired(def); err != nil {
			return err
		}
	}

	if selected[consts.ProfileLabelDBPort] {
		def := r.ProfileInfo.DBInfo.Port
		if def == 0 {
			def = 3306
		}
		if err := r.promptDBPortRequired(def); err != nil {
			return err
		}
	}

	if selected[consts.ProfileLabelDBUser] {
		def := strings.TrimSpace(r.ProfileInfo.DBInfo.User)
		if def == "" {
			def = "root"
		}
		if err := r.promptDBUserRequired(def); err != nil {
			return err
		}
	}

	if selected[consts.ProfileLabelDBPassword] {
		existing := ""
		if r.ProfileInfo != nil {
			existing = r.ProfileInfo.DBInfo.Password
		}
		if err := r.promptDBPasswordKeepCurrent(existing); err != nil {
			return err
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
		validator := prompt.ComposeValidators(
			sshRequired,
			validateNoControlChars(consts.ProfileLabelSSHHost),
			validateNoLeadingTrailingSpaces(consts.ProfileLabelSSHHost),
			validateNoSpaces(consts.ProfileLabelSSHHost),
		)
		v, err := prompt.AskText(consts.ProfilePromptSSHHost, prompt.WithDefault(r.ProfileInfo.SSHTunnel.Host), prompt.WithValidator(validator))
		if err != nil {
			return validation.HandleInputError(err)
		}
		r.ProfileInfo.SSHTunnel.Host = strings.TrimSpace(v)
	}

	if selected[consts.ProfileLabelSSHPort] {
		def := r.ProfileInfo.SSHTunnel.Port
		if def == 0 {
			def = 22
		}
		validator := prompt.ComposeValidators(
			sshRequired,
			validatePortRange(1, 65535, false, consts.ProfileLabelSSHPort),
		)
		v, err := prompt.AskInt(consts.ProfileLabelSSHPort, def, validator)
		if err != nil {
			return validation.HandleInputError(err)
		}
		r.ProfileInfo.SSHTunnel.Port = v
	}

	if selected[consts.ProfileLabelSSHUser] {
		validator := prompt.ComposeValidators(
			validateNoControlChars(consts.ProfileLabelSSHUser),
			validateNoLeadingTrailingSpaces(consts.ProfileLabelSSHUser),
			validateNoSpaces(consts.ProfileLabelSSHUser),
		)
		v, err := prompt.AskText(consts.ProfilePromptSSHUser, prompt.WithDefault(r.ProfileInfo.SSHTunnel.User), prompt.WithValidator(validator))
		if err != nil {
			return validation.HandleInputError(err)
		}
		r.ProfileInfo.SSHTunnel.User = strings.TrimSpace(v)
	}

	if selected[consts.ProfileLabelSSHPassword] {
		existing := ""
		if r.ProfileInfo != nil {
			existing = r.ProfileInfo.SSHTunnel.Password
		}
		print.PrintInfo(consts.ProfileTipKeepCurrentSSHPassword)
		pw, err := prompt.AskPassword(
			consts.ProfilePromptSSHPasswordOptional,
			prompt.ComposeValidators(
				validateOptionalNoControlChars(consts.ProfileLabelSSHPassword),
				validateOptionalNoLeadingTrailingSpaces(consts.ProfileLabelSSHPassword),
			),
		)
		if err != nil {
			return validation.HandleInputError(err)
		}
		if strings.TrimSpace(pw) == "" {
			r.ProfileInfo.SSHTunnel.Password = existing
		} else {
			r.ProfileInfo.SSHTunnel.Password = pw
		}
	}

	if selected[consts.ProfileLabelSSHIdentityFile] {
		validator := prompt.ComposeValidators(
			validateOptionalNoControlChars(consts.ProfileLabelSSHIdentityFile),
			validateNoLeadingTrailingSpaces(consts.ProfileLabelSSHIdentityFile),
			validateOptionalExistingFilePath(consts.ProfileLabelSSHIdentityFile),
		)
		v, err := prompt.AskText(consts.ProfilePromptSSHIdentityFileOptional, prompt.WithDefault(r.ProfileInfo.SSHTunnel.IdentityFile), prompt.WithValidator(validator))
		if err != nil {
			return validation.HandleInputError(err)
		}
		r.ProfileInfo.SSHTunnel.IdentityFile = strings.TrimSpace(v)
	}

	if selected[consts.ProfileLabelSSHLocalPort] {
		validator := validatePortRange(1, 65535, true, consts.ProfileLabelSSHLocalPort)
		v, err := prompt.AskInt(consts.ProfilePromptSSHLocalPort, r.ProfileInfo.SSHTunnel.LocalPort, validator)
		if err != nil {
			return validation.HandleInputError(err)
		}
		r.ProfileInfo.SSHTunnel.LocalPort = v
	}

	return nil
}
