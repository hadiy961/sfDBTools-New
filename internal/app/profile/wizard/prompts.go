// File : internal/app/profile/wizard/prompts.go
// Deskripsi : Prompt wizard untuk nama/config, DB info, dan SSH tunnel
// Author : Hadiyatna Muflihun
// Tanggal : 4 Januari 2026
// Last Modified : 5 Januari 2026

package wizard

import (
	"fmt"
	"os"
	"strings"

	"sfDBTools/internal/app/profile/shared"
	"sfDBTools/pkg/consts"
	"sfDBTools/pkg/input"
	"sfDBTools/pkg/ui"
	"sfDBTools/pkg/validation"

	"github.com/AlecAivazis/survey/v2"
)

func (r *Runner) promptDBConfigName(mode string) error {
	ui.PrintSubHeader(consts.ProfileWizardSubHeaderConfigName)

	for {
		nameValidator := input.ComposeValidators(survey.Required, input.ValidateFilename)
		def := strings.TrimSpace(r.ProfileInfo.Name)
		if def == "" {
			def = consts.ProfileWizardDefaultConfigName
		}
		configName, err := input.AskString(consts.ProfileWizardLabelConfigName, def, nameValidator)
		if err != nil {
			return validation.HandleInputError(err)
		}

		r.ProfileInfo.Name = strings.TrimSpace(configName)
		if r.CheckConfigurationNameUnique != nil {
			if err = r.CheckConfigurationNameUnique(mode); err != nil {
				ui.PrintError(err.Error())
				continue
			}
		}
		break
	}

	r.ProfileInfo.Name = strings.TrimSpace(r.ProfileInfo.Name)
	ui.PrintInfo(consts.ProfileMsgConfigWillBeSavedAsPrefix + shared.BuildProfileFileName(r.ProfileInfo.Name))
	return nil
}

func (r *Runner) promptProfileInfo() error {
	ui.PrintSubHeader(consts.ProfileWizardSubHeaderProfileDetails)

	// Host
	if strings.TrimSpace(r.ProfileInfo.DBInfo.Host) == "" {
		v, err := input.AskString(consts.ProfileLabelDBHost, "localhost", survey.Required)
		if err != nil {
			return validation.HandleInputError(err)
		}
		r.ProfileInfo.DBInfo.Host = v
	}

	// Port
	if r.ProfileInfo.DBInfo.Port == 0 {
		v, err := input.AskInt(consts.ProfileLabelDBPort, 3306, survey.Required)
		if err != nil {
			return validation.HandleInputError(err)
		}
		r.ProfileInfo.DBInfo.Port = v
	}

	// User
	if strings.TrimSpace(r.ProfileInfo.DBInfo.User) == "" {
		v, err := input.AskString(consts.ProfileLabelDBUser, "root", survey.Required)
		if err != nil {
			return validation.HandleInputError(err)
		}
		r.ProfileInfo.DBInfo.User = v
	}

	// Password
	isEditFlow := r.OriginalProfileInfo != nil || r.OriginalProfileName != ""
	if strings.TrimSpace(r.ProfileInfo.DBInfo.Password) == "" {
		envPassword := os.Getenv(consts.ENV_TARGET_DB_PASSWORD)
		if !isEditFlow && envPassword != "" {
			r.ProfileInfo.DBInfo.Password = envPassword
		} else {
			if !isEditFlow {
				ui.PrintWarning(fmt.Sprintf(consts.ProfileWarnEnvVarMissingOrEmptyFmt, consts.ENV_TARGET_DB_PASSWORD, consts.ENV_TARGET_DB_PASSWORD))
			}
			if isEditFlow {
				ui.PrintInfo(consts.ProfileTipKeepCurrentDBPasswordUpdate)
			}
			pw, err := input.AskPassword(consts.ProfileLabelDBPassword, survey.Required)
			if err != nil {
				return validation.HandleInputError(err)
			}
			r.ProfileInfo.DBInfo.Password = pw
		}
	}

	// SSH tunnel
	return r.promptSSHTunnelDetailsIfEnabledOrAsk()
}

// promptSSHTunnelDetailsIfEnabledOrAsk akan:
// - jika sudah Enabled (mis. via flag --ssh), langsung prompt detail yang belum ada
// - jika belum Enabled, akan tanya Yes/No dulu
func (r *Runner) promptSSHTunnelDetailsIfEnabledOrAsk() error {
	enabled := r.ProfileInfo.SSHTunnel.Enabled
	if !enabled {
		v, err := input.AskYesNo(consts.ProfilePromptUseSSHTunnel, false)
		if err != nil {
			return validation.HandleInputError(err)
		}
		enabled = v
	}
	if !enabled {
		r.ProfileInfo.SSHTunnel.Enabled = false
		return nil
	}
	// enabled
	r.ProfileInfo.SSHTunnel.Enabled = true
	return r.promptSSHTunnelDetailsIfEnabled()
}

// promptSSHTunnelDetailsIfEnabled meminta detail SSH hanya jika Enabled dan field penting belum tersedia.
func (r *Runner) promptSSHTunnelDetailsIfEnabled() error {
	if !r.ProfileInfo.SSHTunnel.Enabled {
		return nil
	}

	ssh := &r.ProfileInfo.SSHTunnel
	// Host wajib
	if strings.TrimSpace(ssh.Host) == "" {
		v, err := input.AskString(consts.ProfilePromptSSHHost, "", survey.Required)
		if err != nil {
			return validation.HandleInputError(err)
		}
		ssh.Host = v
	}
	// Port default 22
	if ssh.Port == 0 {
		v, err := input.AskInt(consts.ProfileLabelSSHPort, 22, survey.Required)
		if err != nil {
			return validation.HandleInputError(err)
		}
		ssh.Port = v
	}
	// User optional
	if strings.TrimSpace(ssh.User) == "" {
		v, err := input.AskString(consts.ProfilePromptSSHUser, "", nil)
		if err != nil {
			return validation.HandleInputError(err)
		}
		ssh.User = v
	}

	// Password opsional
	if strings.TrimSpace(ssh.Password) == "" {
		ui.PrintInfo(consts.ProfileTipSSHPasswordOptional)
		pw, err := input.AskPassword(consts.ProfilePromptSSHPasswordOptional, nil)
		if err != nil {
			return validation.HandleInputError(err)
		}
		if pw != "" {
			ssh.Password = pw
		}
	}

	// Identity file opsional
	if strings.TrimSpace(ssh.IdentityFile) == "" {
		v, err := input.AskString(consts.ProfilePromptSSHIdentityFileOptional, "", nil)
		if err != nil {
			return validation.HandleInputError(err)
		}
		ssh.IdentityFile = v
	}

	// Local port opsional
	if ssh.LocalPort == 0 {
		v, err := input.AskInt(consts.ProfilePromptSSHLocalPort, 0, nil)
		if err != nil {
			return validation.HandleInputError(err)
		}
		ssh.LocalPort = v
	}

	return nil
}
