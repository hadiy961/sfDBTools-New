// File : internal/profile/executor/retry_helper.go
// Deskripsi : Helper DRY untuk flow retry saat koneksi DB gagal ketika save profile
// Author : Hadiyatna Muflihun
// Tanggal : 4 Januari 2026
// Last Modified : 14 Januari 2026

package executor

import (
	"sfdbtools/internal/shared/consts"
	"sfdbtools/internal/shared/validation"
	"sfdbtools/internal/ui/print"
	"sfdbtools/internal/ui/prompt"
)

func (e *Executor) handleConnectionFailedRetry(retryWarningMsg string, cancelInfoMsg string) (bool, error) {
	retryInput, askErr := prompt.Confirm(consts.ProfilePromptRetryInputConfig, true)
	if askErr != nil {
		return false, validation.HandleInputError(askErr)
	}
	if retryInput {
		print.PrintWarning(retryWarningMsg)
		return true, nil
	}
	print.PrintInfo(cancelInfoMsg)
	return false, validation.ErrUserCancelled
}

// handleConnectionFailedRetryIfNeeded memusatkan flow retry untuk error koneksi DB.
// Return:
// - shouldRetry=true  => caller bisa `continue`
// - err!=nil          => error final (termasuk ErrUserCancelled)
func (e *Executor) handleConnectionFailedRetryIfNeeded(opErr error, retryWarningMsg string, cancelInfoMsg string) (shouldRetry bool, err error) {
	if opErr == nil {
		return false, nil
	}
	if opErr != validation.ErrConnectionFailedRetry {
		return false, opErr
	}

	retry, err := e.handleConnectionFailedRetry(retryWarningMsg, cancelInfoMsg)
	if err != nil {
		return false, err
	}
	return retry, nil
}
