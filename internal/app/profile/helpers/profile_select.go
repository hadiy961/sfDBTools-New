// File : internal/app/profile/helpers/profile_select.go
// Deskripsi : Helper untuk pilih profile (interactive) + snapshot (untuk show/edit)
// Author : Hadiyatna Muflihun
// Tanggal : 14 Januari 2026
// Last Modified : 14 Januari 2026

package helpers

import (
	"sfdbtools/internal/app/profile/helpers/loader"
	"sfdbtools/internal/app/profile/helpers/selection"
	"sfdbtools/internal/domain"
)

func SelectExistingDBConfig(configDir, purpose string) (domain.ProfileInfo, error) {
	return selection.SelectExistingDBConfig(configDir, purpose)
}

// SelectExistingDBConfigWithSnapshot memilih profile secara interaktif lalu mengembalikan:
// - info (hasil load+parse profile)
// - originalName (nama profile yang dipilih)
// - snapshot (salinan info untuk baseline/original display)
func SelectExistingDBConfigWithSnapshot(configDir string, prompt string) (info *domain.ProfileInfo, originalName string, snapshot *domain.ProfileInfo, err error) {
	return loader.SelectExistingDBConfigWithSnapshot(configDir, prompt)
}
