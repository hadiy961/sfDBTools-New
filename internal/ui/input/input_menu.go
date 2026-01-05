package input

import (
	"os"

	"github.com/AlecAivazis/survey/v2"
)

// ShowMenu displays a menu and returns the selected index (1-based).
func ShowMenu(title string, options []string) (int, error) {
	if err := ensureInteractiveAllowed(); err != nil {
		return 0, err
	}
	var selectedIndex int

	prompt := &survey.Select{
		Message: title,
		Options: options,
	}

	err := survey.AskOne(prompt, &selectedIndex, survey.WithStdio(
		os.Stdin,
		os.Stdout,
		os.Stderr,
	))
	if err != nil {
		return 0, err
	}

	// survey returns 0-based index; convert to 1-based.
	return selectedIndex + 1, nil
}

// ShowMultiSelect displays a multi-select menu and returns selected indices (1-based).
func ShowMultiSelect(title string, options []string) ([]int, error) {
	if err := ensureInteractiveAllowed(); err != nil {
		return nil, err
	}
	var selected []string

	prompt := &survey.MultiSelect{
		Message: title,
		Options: options,
	}

	err := survey.AskOne(prompt, &selected, survey.WithStdio(
		os.Stdin,
		os.Stdout,
		os.Stderr,
	))
	if err != nil {
		return nil, err
	}

	idxs := make([]int, 0, len(selected))
	for _, sel := range selected {
		for i, opt := range options {
			if opt == sel {
				idxs = append(idxs, i+1)
				break
			}
		}
	}
	return idxs, nil
}
