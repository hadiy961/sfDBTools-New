// File : internal/app/profile/wizard/prompt_helpers.go
// Deskripsi : Helper methods untuk mengurangi boilerplate error handling di wizard prompts
// Author : Hadiyatna Muflihun
// Tanggal : 15 Januari 2026
// Last Modified : 15 Januari 2026

package wizard

import (
	"strings"

	"sfdbtools/internal/shared/validation"
	"sfdbtools/internal/ui/prompt"

	"github.com/AlecAivazis/survey/v2"
)

// askAndAssignText is a helper that combines prompt.AskText with automatic trimming and error handling.
// Reduces boilerplate from 5 lines to 1 line per prompt.
func (r *Runner) askAndAssignText(
	target *string,
	label string,
	opts ...prompt.TextOption,
) error {
	v, err := prompt.AskText(label, opts...)
	if err != nil {
		return validation.HandleInputError(err)
	}
	*target = strings.TrimSpace(v)
	return nil
}

// askAndAssignInt is a helper that combines prompt.AskInt with error handling.
// Reduces boilerplate for integer prompts.
func (r *Runner) askAndAssignInt(
	target *int,
	label string,
	defaultVal int,
	validator survey.Validator,
) error {
	v, err := prompt.AskInt(label, defaultVal, validator)
	if err != nil {
		return validation.HandleInputError(err)
	}
	*target = v
	return nil
}

// askAndAssignPassword is a helper that combines prompt.AskPassword with error handling.
// Reduces boilerplate for password prompts.
func (r *Runner) askAndAssignPassword(
	target *string,
	label string,
	validator survey.Validator,
) error {
	v, err := prompt.AskPassword(label, validator)
	if err != nil {
		return validation.HandleInputError(err)
	}
	*target = v
	return nil
}

// askAndAssignBool is a helper that combines prompt.Confirm with error handling.
// Reduces boilerplate for boolean prompts.
func (r *Runner) askAndAssignBool(
	target *bool,
	label string,
	defaultVal bool,
) error {
	v, err := prompt.Confirm(label, defaultVal)
	if err != nil {
		return validation.HandleInputError(err)
	}
	*target = v
	return nil
}
