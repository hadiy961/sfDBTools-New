// File : internal/app/profile/wizard/helpers.go
// Deskripsi : Helper kecil untuk wizard profile (selection & prompt yang dipakai ulang)
// Author : Hadiyatna Muflihun
// Tanggal : 9 Januari 2026
// Last Modified : 9 Januari 2026

package wizard

import (
	"fmt"
	"strings"

	"sfdbtools/internal/shared/consts"
	"sfdbtools/internal/shared/validation"
	"sfdbtools/internal/ui/print"
	"sfdbtools/internal/ui/prompt"

	"github.com/AlecAivazis/survey/v2"
)

func selectedSetFromIdxs(fields []string, idxs []int) map[string]bool {
	selected := make(map[string]bool, len(idxs))
	for _, i := range idxs {
		// idxs dari prompt.SelectMany adalah 0-based
		if i >= 0 && i < len(fields) {
			selected[fields[i]] = true
		}
	}
	return selected
}

func selectManyFieldsOrCancel(promptText string, fields []string) (map[string]bool, error) {
	_, idxs, err := prompt.SelectMany(promptText, fields, nil)
	if err != nil {
		return nil, validation.HandleInputError(err)
	}
	if len(idxs) == 0 {
		print.PrintWarning(consts.ProfileMsgNoFieldsSelected)
		return nil, validation.ErrUserCancelled
	}
	return selectedSetFromIdxs(fields, idxs), nil
}

func promptNewEncryptionKeyConfirmed() (string, error) {
	newKey, err := prompt.AskPassword(
		consts.ProfilePromptNewEncryptionKey,
		prompt.ComposeValidators(
			validateNotBlank(consts.ProfileFieldEncryptionKey),
			validateNoControlChars(consts.ProfileFieldEncryptionKey),
			validateOptionalNoLeadingTrailingSpaces(consts.ProfileFieldEncryptionKey),
		),
	)
	if err != nil {
		return "", validation.HandleInputError(err)
	}
	confirmKey, err := prompt.AskPassword(
		consts.ProfilePromptConfirmNewEncryptionKey,
		prompt.ComposeValidators(
			validateNotBlank(consts.ProfileFieldEncryptionKey),
			validateNoControlChars(consts.ProfileFieldEncryptionKey),
			validateOptionalNoLeadingTrailingSpaces(consts.ProfileFieldEncryptionKey),
		),
	)
	if err != nil {
		return "", validation.HandleInputError(err)
	}
	if strings.TrimSpace(newKey) != strings.TrimSpace(confirmKey) {
		return "", validation.HandleInputError(fmt.Errorf(consts.ProfileErrNewEncryptionKeyMismatch))
	}
	return strings.TrimSpace(newKey), nil
}

func (r *Runner) promptDBHostRequired(defaultValue string) error {
	validator := prompt.ComposeValidators(
		validateNotBlank(consts.ProfileLabelDBHost),
		validateNoControlChars(consts.ProfileLabelDBHost),
		validateNoLeadingTrailingSpaces(consts.ProfileLabelDBHost),
		validateNoSpaces(consts.ProfileLabelDBHost),
	)
	v, err := prompt.AskText(consts.ProfileLabelDBHost, prompt.WithDefault(defaultValue), prompt.WithValidator(validator))
	if err != nil {
		return validation.HandleInputError(err)
	}
	r.State.ProfileInfo.DBInfo.Host = strings.TrimSpace(v)
	return nil
}

func (r *Runner) promptDBPortRequired(defaultValue int) error {
	validator := prompt.ComposeValidators(
		survey.Required,
		validatePortRange(1, 65535, false, consts.ProfileLabelDBPort),
	)
	v, err := prompt.AskInt(consts.ProfileLabelDBPort, defaultValue, validator)
	if err != nil {
		return validation.HandleInputError(err)
	}
	r.State.ProfileInfo.DBInfo.Port = v
	return nil
}

func (r *Runner) promptDBUserRequired(defaultValue string) error {
	validator := prompt.ComposeValidators(
		validateNotBlank(consts.ProfileLabelDBUser),
		validateNoControlChars(consts.ProfileLabelDBUser),
		validateNoLeadingTrailingSpaces(consts.ProfileLabelDBUser),
		validateNoSpaces(consts.ProfileLabelDBUser),
	)
	v, err := prompt.AskText(consts.ProfileLabelDBUser, prompt.WithDefault(defaultValue), prompt.WithValidator(validator))
	if err != nil {
		return validation.HandleInputError(err)
	}
	r.State.ProfileInfo.DBInfo.User = strings.TrimSpace(v)
	return nil
}

func (r *Runner) promptDBPasswordKeepCurrent(existing string) error {
	print.PrintInfo(consts.ProfileTipKeepCurrentDBPassword)
	pw, err := prompt.AskPassword(
		consts.ProfileLabelDBPassword,
		prompt.ComposeValidators(
			validateOptionalNoControlChars(consts.ProfileLabelDBPassword),
			validateOptionalNoLeadingTrailingSpaces(consts.ProfileLabelDBPassword),
		),
	)
	if err != nil {
		return validation.HandleInputError(err)
	}
	if strings.TrimSpace(pw) == "" {
		r.State.ProfileInfo.DBInfo.Password = existing
	} else {
		r.State.ProfileInfo.DBInfo.Password = pw
	}
	return nil
}
