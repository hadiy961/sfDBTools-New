package validation

import (
	"errors"

	"github.com/AlecAivazis/survey/v2/terminal"
)

var ErrUserCancelled = errors.New("user_cancelled")
var ErrConnectionFailedRetry = errors.New("connection_failed_retry")

// HandleInputError menangani error dari input/survey dan mengubahnya menjadi ErrUserCancelled jika perlu.
func HandleInputError(err error) error {
	if err == terminal.InterruptErr {
		return ErrUserCancelled
	}
	return err
}
