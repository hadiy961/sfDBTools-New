// File : internal/app/profile/wizard/prompts.go
// Deskripsi : Prompt wizard untuk nama/config, DB info, dan SSH tunnel
// Author : Hadiyatna Muflihun
// Tanggal : 4 Januari 2026
// Last Modified : 9 Januari 2026

package wizard

import (
	"fmt"
	"os"
	"strings"

	"sfdbtools/internal/app/profile/shared"
	"sfdbtools/internal/shared/consts"
	"sfdbtools/internal/shared/validation"
	"sfdbtools/internal/ui/print"
	"sfdbtools/internal/ui/prompt"

	"github.com/AlecAivazis/survey/v2"
)

func (r *Runner) promptDBConfigName(mode string) error {
	print.PrintSubHeader(consts.ProfileWizardSubHeaderConfigName)

	for {
		nameValidator := prompt.ComposeValidators(
			validateNotBlank(consts.ProfileWizardLabelConfigName),
			validateNoControlChars(consts.ProfileWizardLabelConfigName),
			validateNoLeadingTrailingSpaces(consts.ProfileWizardLabelConfigName),
			prompt.ValidateFilename,
		)
		def := strings.TrimSpace(r.State.ProfileInfo.Name)
		if def == "" {
			def = consts.ProfileWizardDefaultConfigName
		}
		configName, err := prompt.AskText(consts.ProfileWizardLabelConfigName, prompt.WithDefault(def), prompt.WithValidator(nameValidator))
		if err != nil {
			return validation.HandleInputError(err)
		}

		r.State.ProfileInfo.Name = strings.TrimSpace(configName)
		if r.CheckNameUnique != nil {
			if err = r.CheckNameUnique(mode); err != nil {
				print.PrintError(err.Error())
				continue
			}
		}
		break
	}

	r.State.ProfileInfo.Name = strings.TrimSpace(r.State.ProfileInfo.Name)
	print.PrintInfo(consts.ProfileMsgConfigWillBeSavedAsPrefix + shared.BuildProfileFileName(r.State.ProfileInfo.Name))
	return nil
}

func (r *Runner) promptProfileInfo() error {
	print.PrintSubHeader(consts.ProfileWizardSubHeaderProfileDetails)

	// Host
	if strings.TrimSpace(r.State.ProfileInfo.DBInfo.Host) == "" {
		validator := prompt.ComposeValidators(
			validateNotBlank(consts.ProfileLabelDBHost),
			validateNoControlChars(consts.ProfileLabelDBHost),
			validateNoLeadingTrailingSpaces(consts.ProfileLabelDBHost),
			validateNoSpaces(consts.ProfileLabelDBHost),
		)
		v, err := prompt.AskText(consts.ProfileLabelDBHost, prompt.WithDefault("localhost"), prompt.WithValidator(validator))
		if err != nil {
			return validation.HandleInputError(err)
		}
		r.State.ProfileInfo.DBInfo.Host = strings.TrimSpace(v)
	}

	// Port
	if r.State.ProfileInfo.DBInfo.Port == 0 {
		validator := prompt.ComposeValidators(
			survey.Required,
			validatePortRange(1, 65535, false, consts.ProfileLabelDBPort),
		)
		v, err := prompt.AskInt(consts.ProfileLabelDBPort, 3306, validator)
		if err != nil {
			return validation.HandleInputError(err)
		}
		r.State.ProfileInfo.DBInfo.Port = v
	}

	// User
	if strings.TrimSpace(r.State.ProfileInfo.DBInfo.User) == "" {
		validator := prompt.ComposeValidators(
			validateNotBlank(consts.ProfileLabelDBUser),
			validateNoControlChars(consts.ProfileLabelDBUser),
			validateNoLeadingTrailingSpaces(consts.ProfileLabelDBUser),
			validateNoSpaces(consts.ProfileLabelDBUser),
		)
		v, err := prompt.AskText(consts.ProfileLabelDBUser, prompt.WithDefault("root"), prompt.WithValidator(validator))
		if err != nil {
			return validation.HandleInputError(err)
		}
		r.State.ProfileInfo.DBInfo.User = strings.TrimSpace(v)
	}

	// Password
	isEditFlow := r.State.OriginalProfileInfo != nil || r.State.OriginalProfileName != ""
	if strings.TrimSpace(r.State.ProfileInfo.DBInfo.Password) == "" {
		envPassword := os.Getenv(consts.ENV_TARGET_DB_PASSWORD)
		if !isEditFlow && envPassword != "" {
			r.State.ProfileInfo.DBInfo.Password = envPassword
		} else {
			if !isEditFlow {
				print.PrintWarning(fmt.Sprintf(consts.ProfileWarnEnvVarMissingOrEmptyFmt, consts.ENV_TARGET_DB_PASSWORD, consts.ENV_TARGET_DB_PASSWORD))
			}
			if isEditFlow {
				print.PrintInfo(consts.ProfileTipKeepCurrentDBPasswordUpdate)
			}
			pw, err := prompt.AskPassword(
				consts.ProfileLabelDBPassword,
				prompt.ComposeValidators(
					validateNotBlank(consts.ProfileLabelDBPassword),
					validateNoControlChars(consts.ProfileLabelDBPassword),
					validateOptionalNoLeadingTrailingSpaces(consts.ProfileLabelDBPassword),
				),
			)
			if err != nil {
				return validation.HandleInputError(err)
			}
			r.State.ProfileInfo.DBInfo.Password = pw
		}
	}

	// SSH tunnel
	return r.promptSSHTunnelDetailsIfEnabledOrAsk()
}

// promptSSHTunnelDetailsIfEnabledOrAsk akan:
// - jika sudah Enabled (mis. via flag --ssh), langsung prompt detail yang belum ada
// - jika belum Enabled, akan tanya Yes/No dulu
func (r *Runner) promptSSHTunnelDetailsIfEnabledOrAsk() error {
	enabled := r.State.ProfileInfo.SSHTunnel.Enabled
	if !enabled {
		v, err := prompt.Confirm(consts.ProfilePromptUseSSHTunnel, false)
		if err != nil {
			return validation.HandleInputError(err)
		}
		enabled = v
	}
	if !enabled {
		r.State.ProfileInfo.SSHTunnel.Enabled = false
		return nil
	}
	// enabled
	r.State.ProfileInfo.SSHTunnel.Enabled = true
	return r.promptSSHTunnelDetailsIfEnabled()
}

// promptSSHTunnelDetailsIfEnabled meminta detail SSH hanya jika Enabled dan field penting belum tersedia.
func (r *Runner) promptSSHTunnelDetailsIfEnabled() error {
	if !r.State.ProfileInfo.SSHTunnel.Enabled {
		return nil
	}

	ssh := &r.State.ProfileInfo.SSHTunnel
	// Host wajib
	if strings.TrimSpace(ssh.Host) == "" {
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
		ssh.Host = strings.TrimSpace(v)
	}
	// Port default 22
	if ssh.Port == 0 {
		validator := prompt.ComposeValidators(
			survey.Required,
			validatePortRange(1, 65535, false, consts.ProfileLabelSSHPort),
		)
		v, err := prompt.AskInt(consts.ProfileLabelSSHPort, 22, validator)
		if err != nil {
			return validation.HandleInputError(err)
		}
		ssh.Port = v
	}
	// User optional
	if strings.TrimSpace(ssh.User) == "" {
		validator := prompt.ComposeValidators(
			validateNoControlChars(consts.ProfileLabelSSHUser),
			validateNoLeadingTrailingSpaces(consts.ProfileLabelSSHUser),
			validateNoSpaces(consts.ProfileLabelSSHUser),
		)
		v, err := prompt.AskText(consts.ProfilePromptSSHUser, prompt.WithDefault(""), prompt.WithValidator(validator))
		if err != nil {
			return validation.HandleInputError(err)
		}
		ssh.User = strings.TrimSpace(v)
	}

	// Password opsional
	if strings.TrimSpace(ssh.Password) == "" {
		print.PrintInfo(consts.ProfileTipSSHPasswordOptional)
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
		if strings.TrimSpace(pw) != "" {
			ssh.Password = pw
		}
	}

	// Identity file opsional
	if strings.TrimSpace(ssh.IdentityFile) == "" {
		validator := prompt.ComposeValidators(
			validateOptionalNoControlChars(consts.ProfileLabelSSHIdentityFile),
			validateNoLeadingTrailingSpaces(consts.ProfileLabelSSHIdentityFile),
			validateOptionalExistingFilePath(consts.ProfileLabelSSHIdentityFile),
		)
		v, err := prompt.AskText(consts.ProfilePromptSSHIdentityFileOptional, prompt.WithDefault(""), prompt.WithValidator(validator))
		if err != nil {
			return validation.HandleInputError(err)
		}
		ssh.IdentityFile = strings.TrimSpace(v)
	}

	// Local port opsional
	if ssh.LocalPort == 0 {
		validator := validatePortRange(1, 65535, true, consts.ProfileLabelSSHLocalPort)
		v, err := prompt.AskInt(consts.ProfilePromptSSHLocalPort, 0, validator)
		if err != nil {
			return validation.HandleInputError(err)
		}
		ssh.LocalPort = v
	}

	return nil
}
