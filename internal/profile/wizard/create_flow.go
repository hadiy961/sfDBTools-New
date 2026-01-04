// File : internal/profile/wizard/create_flow.go
// Deskripsi : Flow wizard untuk pembuatan profile
// Author : Hadiyatna Muflihun
// Tanggal : 4 Januari 2026
// Last Modified : 4 Januari 2026

package wizard

import (
	"strings"

	"sfDBTools/internal/types"
	"sfDBTools/pkg/consts"
	"sfDBTools/pkg/helper"
	"sfDBTools/pkg/ui"
)

func (r *Runner) runCreateFlow(mode string) error {
	if r.ProfileInfo == nil {
		r.ProfileInfo = &types.ProfileInfo{}
	}

	// 1. Config Name (skip jika sudah diberikan via flag/env)
	if strings.TrimSpace(r.ProfileInfo.Name) == "" {
		if err := r.promptDBConfigName(mode); err != nil {
			return err
		}
	} else {
		r.ProfileInfo.Name = helper.TrimProfileSuffix(r.ProfileInfo.Name)
		if r.CheckConfigurationNameUnique != nil {
			if err := r.CheckConfigurationNameUnique(mode); err != nil {
				ui.PrintError(err.Error())
				// Jika nama dari flag/env ternyata bentrok, minta user input nama baru.
				if err2 := r.promptDBConfigName(mode); err2 != nil {
					return err2
				}
			} else {
				ui.PrintInfo(consts.ProfileMsgConfigWillBeSavedAsPrefix + buildFileName(r.ProfileInfo.Name))
			}
		}
	}

	// 2. Profile Details (skip field yang sudah ada agar user tidak input dua kali)
	return r.promptProfileInfo()
}
