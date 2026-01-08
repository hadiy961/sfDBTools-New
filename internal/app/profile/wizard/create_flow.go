// File : internal/app/profile/wizard/create_flow.go
// Deskripsi : Flow wizard untuk pembuatan profile
// Author : Hadiyatna Muflihun
// Tanggal : 4 Januari 2026
// Last Modified : 5 Januari 2026
package wizard

import (
	"strings"

	profilehelper "sfdbtools/internal/app/profile/helpers"
	"sfdbtools/internal/app/profile/shared"
	"sfdbtools/internal/domain"
	"sfdbtools/internal/ui/print"
	"sfdbtools/pkg/consts"
)

func (r *Runner) runCreateFlow(mode string) error {
	if r.ProfileInfo == nil {
		r.ProfileInfo = &domain.ProfileInfo{}
	}

	// 1. Config Name (skip jika sudah diberikan via flag/env)
	if strings.TrimSpace(r.ProfileInfo.Name) == "" {
		if err := r.promptDBConfigName(mode); err != nil {
			return err
		}
	} else {
		r.ProfileInfo.Name = profilehelper.TrimProfileSuffix(r.ProfileInfo.Name)
		if r.CheckConfigurationNameUnique != nil {
			if err := r.CheckConfigurationNameUnique(mode); err != nil {
				print.PrintError(err.Error())
				// Jika nama dari flag/env ternyata bentrok, minta user input nama baru.
				if err2 := r.promptDBConfigName(mode); err2 != nil {
					return err2
				}
			} else {
				print.PrintInfo(consts.ProfileMsgConfigWillBeSavedAsPrefix + shared.BuildProfileFileName(r.ProfileInfo.Name))
			}
		}
	}

	// 2. Profile Details (skip field yang sudah ada agar user tidak input dua kali)
	return r.promptProfileInfo()
}
