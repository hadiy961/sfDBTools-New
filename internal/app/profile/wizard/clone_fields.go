// File : internal/app/profile/wizard/clone_fields.go
// Deskripsi : Prompt edit field untuk clone (multi-select; tanpa ubah nama)
// Author : Hadiyatna Muflihun
// Tanggal : 21 Januari 2026
// Last Modified : 21 Januari 2026

package wizard

import (
	"strings"

	"sfdbtools/internal/shared/consts"
	"sfdbtools/internal/shared/validation"
	"sfdbtools/internal/ui/print"
	"sfdbtools/internal/ui/prompt"

	"github.com/AlecAivazis/survey/v2"
)

// promptCloneSelectedFields meminta user memilih field yang ingin diubah via multi-select.
// Catatan: Nama profile TIDAK ditawarkan untuk diubah, sesuai UX clone (nama auto: <source>-clone).
func (r *Runner) promptCloneSelectedFields() error {
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
		if err := r.promptDBConfigName(consts.ProfileModeClone); err != nil {
			return err
		}
		if cloneOpts, ok := r.State.CloneOptions(); ok && cloneOpts != nil {
			cloneOpts.TargetName = strings.TrimSpace(r.State.ProfileInfo.Name)
		}
	}

	if selected[consts.ProfileFieldEncryptionKey] {
		newKey, err := promptNewEncryptionKeyConfirmed()
		if err != nil {
			return err
		}
		newKey = strings.TrimSpace(newKey)
		if r.State != nil && r.State.ProfileInfo != nil {
			r.State.ProfileInfo.EncryptionKey = newKey
		}
		if cloneOpts, ok := r.State.CloneOptions(); ok && cloneOpts != nil {
			cloneOpts.NewProfileKey = newKey
		}
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
	}

	sshRequired := survey.Validator(nil)
	if r.State.ProfileInfo.SSHTunnel.Enabled {
		sshRequired = survey.Required
	}

	if selected[consts.ProfileLabelSSHHost] {
		validator := prompt.ComposeValidators(
			sshRequired,
			validateNoControlChars(consts.ProfileLabelSSHHost),
			validateNoLeadingTrailingSpaces(consts.ProfileLabelSSHHost),
			validateNoSpaces(consts.ProfileLabelSSHHost),
		)
		v, err := prompt.AskText(consts.ProfilePromptSSHHost, prompt.WithDefault(r.State.ProfileInfo.SSHTunnel.Host), prompt.WithValidator(validator))
		if err != nil {
			return validation.HandleInputError(err)
		}
		r.State.ProfileInfo.SSHTunnel.Host = strings.TrimSpace(v)
	}

	if selected[consts.ProfileLabelSSHPort] {
		def := r.State.ProfileInfo.SSHTunnel.Port
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
		r.State.ProfileInfo.SSHTunnel.Port = v
	}

	if selected[consts.ProfileLabelSSHUser] {
		validator := prompt.ComposeValidators(
			validateNoControlChars(consts.ProfileLabelSSHUser),
			validateNoLeadingTrailingSpaces(consts.ProfileLabelSSHUser),
			validateNoSpaces(consts.ProfileLabelSSHUser),
		)
		v, err := prompt.AskText(consts.ProfilePromptSSHUser, prompt.WithDefault(r.State.ProfileInfo.SSHTunnel.User), prompt.WithValidator(validator))
		if err != nil {
			return validation.HandleInputError(err)
		}
		r.State.ProfileInfo.SSHTunnel.User = strings.TrimSpace(v)
	}

	if selected[consts.ProfileLabelSSHPassword] {
		existing := ""
		if r.State.ProfileInfo != nil {
			existing = r.State.ProfileInfo.SSHTunnel.Password
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
			r.State.ProfileInfo.SSHTunnel.Password = existing
		} else {
			r.State.ProfileInfo.SSHTunnel.Password = pw
		}
	}

	if selected[consts.ProfileLabelSSHIdentityFile] {
		validator := prompt.ComposeValidators(
			validateOptionalNoControlChars(consts.ProfileLabelSSHIdentityFile),
			validateNoLeadingTrailingSpaces(consts.ProfileLabelSSHIdentityFile),
			validateOptionalExistingFilePath(consts.ProfileLabelSSHIdentityFile),
		)
		v, err := prompt.AskText(consts.ProfilePromptSSHIdentityFileOptional, prompt.WithDefault(r.State.ProfileInfo.SSHTunnel.IdentityFile), prompt.WithValidator(validator))
		if err != nil {
			return validation.HandleInputError(err)
		}
		r.State.ProfileInfo.SSHTunnel.IdentityFile = strings.TrimSpace(v)
	}

	if selected[consts.ProfileLabelSSHLocalPort] {
		validator := validatePortRange(1, 65535, true, consts.ProfileLabelSSHLocalPort)
		v, err := prompt.AskInt(consts.ProfilePromptSSHLocalPort, r.State.ProfileInfo.SSHTunnel.LocalPort, validator)
		if err != nil {
			return validation.HandleInputError(err)
		}
		r.State.ProfileInfo.SSHTunnel.LocalPort = v
	}

	return nil
}
