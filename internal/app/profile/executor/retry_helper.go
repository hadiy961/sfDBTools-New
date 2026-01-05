// File : internal/profile/executor/retry_helper.go
// Deskripsi : Helper DRY untuk flow retry saat koneksi DB gagal ketika save profile
// Author : Hadiyatna Muflihun
// Tanggal : 4 Januari 2026
// Last Modified : 4 Januari 2026

package executor

import (
	"sfDBTools/pkg/consts"
	"sfDBTools/pkg/input"
	"sfDBTools/pkg/ui"
	"sfDBTools/pkg/validation"
)

func (e *Executor) handleConnectionFailedRetry(retryWarningMsg string, cancelInfoMsg string) (bool, error) {
	retryInput, askErr := input.AskYesNo(consts.ProfilePromptRetryInputConfig, true)
	if askErr != nil {
		return false, validation.HandleInputError(askErr)
	}
	if retryInput {
		ui.PrintWarning(retryWarningMsg)
		return true, nil
	}
	ui.PrintInfo(cancelInfoMsg)
	return false, validation.ErrUserCancelled
}
