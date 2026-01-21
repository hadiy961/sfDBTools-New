// File : internal/ui/prompt/prompt.go
// Deskripsi : Prompt/input interaktif (select/confirm/password/text)
// Author : Hadiyatna Muflihun
// Tanggal : 5 Januari 2026
// Last Modified : 9 Januari 2026

package prompt

import (
	"fmt"
	"os"
	"sfdbtools/internal/ui/input"

	"sfdbtools/internal/shared/runtimecfg"

	"github.com/AlecAivazis/survey/v2"
	"github.com/mattn/go-isatty"
)

type textOptions struct {
	defaultValue string
	validator    survey.Validator
}

type TextOption func(*textOptions)

func WithDefault(v string) TextOption {
	return func(o *textOptions) { o.defaultValue = v }
}

func WithValidator(v survey.Validator) TextOption {
	return func(o *textOptions) { o.validator = v }
}

func AskText(label string, opts ...TextOption) (string, error) {
	cfg := &textOptions{}
	for _, opt := range opts {
		opt(cfg)
	}
	return input.AskString(label, cfg.defaultValue, cfg.validator)
}

func AskTicket(action string) (string, error) {
	return input.AskTicket(action)
}

func ValidateFilename(ans interface{}) error {
	return input.ValidateFilename(ans)
}

func ComposeValidators(validators ...survey.Validator) survey.Validator {
	return input.ComposeValidators(validators...)
}

func AskPassword(label string, validator survey.Validator) (string, error) {
	return input.AskPassword(label, validator)
}

func AskInt(label string, defaultValue int, validator survey.Validator) (int, error) {
	return input.AskInt(label, defaultValue, validator)
}

func Confirm(label string, defaultYes bool) (bool, error) {
	return input.AskYesNo(label, defaultYes)
}

// AskConfirm adalah alias untuk Confirm untuk backward compatibility.
func AskConfirm(label string, defaultYes bool) (bool, error) {
	return Confirm(label, defaultYes)
}

func PromptPassword(message string) (string, error) {
	return input.PromptPassword(message)
}

func PromptConfirm(message string) (bool, error) {
	return input.PromptConfirm(message)
}

func SelectOne(label string, items []string, defaultIndex int) (string, int, error) {
	if defaultIndex >= 0 && defaultIndex < len(items) {
		selected, err := input.SelectSingleFromListWithDefault(items, label, items[defaultIndex])
		if err != nil {
			return "", -1, err
		}
		idx := indexOf(items, selected)
		return selected, idx, nil
	}
	selected, err := input.SelectSingleFromList(items, label)
	if err != nil {
		return "", -1, err
	}
	idx := indexOf(items, selected)
	return selected, idx, nil
}

func SelectMany(label string, items []string, defaults []int) ([]string, []int, error) {
	_ = defaults
	idxs1, err := input.ShowMultiSelect(label, items)
	if err != nil {
		return nil, nil, err
	}

	selectedItems := make([]string, 0, len(idxs1))
	selectedIdxs := make([]int, 0, len(idxs1))
	for _, idx1 := range idxs1 {
		idx := idx1 - 1
		if idx >= 0 && idx < len(items) {
			selectedItems = append(selectedItems, items[idx])
			selectedIdxs = append(selectedIdxs, idx)
		}
	}
	return selectedItems, selectedIdxs, nil
}

func SelectFile(directory string, label string, extensions []string) (string, error) {
	return input.SelectFileInteractive(directory, label, extensions)
}

func WaitForEnter(message ...string) {
	// Jangan pernah block di mode non-interaktif (mis. saat output dipipe atau --quiet).
	if runtimecfg.IsQuiet() || runtimecfg.IsDaemon() {
		return
	}
	if !isatty.IsTerminal(os.Stdin.Fd()) || !isatty.IsTerminal(os.Stdout.Fd()) {
		return
	}

	msg := "Tekan Enter untuk melanjutkan..."
	if len(message) > 0 && message[0] != "" {
		msg = message[0]
	}
	fmt.Print(msg)
	_, _ = fmt.Scanln()
}

func indexOf(items []string, v string) int {
	for i, s := range items {
		if s == v {
			return i
		}
	}
	return -1
}
