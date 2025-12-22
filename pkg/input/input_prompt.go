// File : pkg/input/input_prompt.go
// Deskripsi : Fungsi utilitas untuk input interaktif dari user
// Author : Hadiyatna Muflihun
// Tanggal : 2024-10-03
// Last Modified : 2024-10-03
package input

import (
	"fmt"
	"strconv"

	"github.com/AlecAivazis/survey/v2"
)

// AskPassword prompts user for a password with input masking.
func AskPassword(message string, validator survey.Validator) (string, error) {
	answer := ""
	prompt := &survey.Password{Message: message}

	var opts []survey.AskOpt
	if validator != nil {
		opts = append(opts, survey.WithValidator(validator))
	}

	err := survey.AskOne(prompt, &answer, opts...)
	return answer, err
}

func AskInt(message string, defaultValue int, validator survey.Validator) (int, error) {
	var answer string
	prompt := &survey.Input{
		Message: message,
		Default: fmt.Sprintf("%d", defaultValue),
	}

	var opts []survey.AskOpt
	if validator != nil {
		opts = append(opts, survey.WithValidator(validator))
	}

	if err := survey.AskOne(prompt, &answer, opts...); err != nil {
		return 0, err
	}
	val, _ := strconv.Atoi(answer)
	return val, nil
}

func AskString(message, defaultValue string, validator survey.Validator) (string, error) {
	var answer string
	prompt := &survey.Input{
		Message: message,
		Default: defaultValue,
	}

	var opts []survey.AskOpt
	if validator != nil {
		opts = append(opts, survey.WithValidator(validator))
	}

	err := survey.AskOne(prompt, &answer, opts...)
	return answer, err
}

func AskYesNo(question string, defaultValue bool) (bool, error) {
	var response bool
	prompt := &survey.Confirm{
		Message: question,
		Default: defaultValue,
	}
	return response, survey.AskOne(prompt, &response)
}

// PromptString meminta input string dari user.
func PromptString(message string) (string, error) {
	return AskString(message, "", nil)
}

// PromptPassword meminta input password dari user.
func PromptPassword(message string) (string, error) {
	return AskPassword(message, nil)
}

// PromptConfirm meminta konfirmasi yes/no dari user.
func PromptConfirm(message string) (bool, error) {
	return AskYesNo(message, true)
}

// SelectSingleFromList menampilkan list dan meminta user memilih satu item.
func SelectSingleFromList(items []string, message string) (string, error) {
	var selected string
	prompt := &survey.Select{
		Message: message,
		Options: items,
	}
	err := survey.AskOne(prompt, &selected)
	return selected, err
}
