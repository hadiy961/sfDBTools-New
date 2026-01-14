// File : internal/app/profile/wizard/create_flow.go
// Deskripsi : Flow wizard untuk pembuatan profile
// Author : Hadiyatna Muflihun
// Tanggal : 4 Januari 2026
// Last Modified : 14 Januari 2026
package wizard

import (
	"strings"

	"sfdbtools/internal/app/profile/shared"
	"sfdbtools/internal/domain"
	"sfdbtools/internal/shared/consts"
	"sfdbtools/internal/ui/print"
)

func (r *Runner) runCreateFlow(mode string) error {
	if r.State.ProfileInfo == nil {
		r.State.ProfileInfo = &domain.ProfileInfo{}
	}

	// 1. Config Name (skip jika sudah diberikan via flag/env)
	if strings.TrimSpace(r.State.ProfileInfo.Name) == "" {
		if err := r.promptDBConfigName(mode); err != nil {
			return err
		}
	} else {
		r.State.ProfileInfo.Name = shared.TrimProfileSuffix(r.State.ProfileInfo.Name)
		if r.Validator != nil {
			if err := r.Validator.CheckNameUnique(mode); err != nil {
				print.PrintError(err.Error())
				// Jika nama dari flag/env ternyata bentrok, minta user input nama baru.
				if err2 := r.promptDBConfigName(mode); err2 != nil {
					return err2
				}
			} else {
				print.PrintInfo(consts.ProfileMsgConfigWillBeSavedAsPrefix + shared.BuildProfileFileName(r.State.ProfileInfo.Name))
			}
		}
	}

	// 2. Profile Details (skip field yang sudah ada agar user tidak input dua kali)
	return r.promptProfileInfo()
}
