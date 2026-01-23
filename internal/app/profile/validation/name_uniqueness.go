// File : internal/app/profile/validation/name_uniqueness.go
// Deskripsi : Name uniqueness validation untuk profile (dipindahkan dari validation.go)
// Author : Hadiyatna Muflihun
// Tanggal : 15 Januari 2026
// Last Modified : 15 Januari 2026
package validation

import (
	"fmt"
	"path/filepath"
	"strings"

	"sfdbtools/internal/app/profile/connection"
	profileerrors "sfdbtools/internal/app/profile/errors"
	profilemodel "sfdbtools/internal/app/profile/model"
	"sfdbtools/internal/domain"
	appconfig "sfdbtools/internal/services/config"
	"sfdbtools/internal/shared/consts"
	"sfdbtools/internal/shared/fsops"
	"sfdbtools/internal/shared/validation"
)

// deriveProfileName mendapatkan nama profile yang valid dari ProfileInfo
func deriveProfileName(profileInfo *domain.ProfileInfo, originalProfileInfo *domain.ProfileInfo) string {
	if profileInfo != nil {
		if v := strings.TrimSpace(connection.TrimProfileSuffix(profileInfo.Name)); v != "" {
			return v
		}
	}
	if originalProfileInfo != nil {
		if v := strings.TrimSpace(connection.TrimProfileSuffix(originalProfileInfo.Name)); v != "" {
			return v
		}
	}
	if profileInfo != nil {
		if p := strings.TrimSpace(profileInfo.Path); p != "" {
			return strings.TrimSpace(connection.TrimProfileSuffix(filepath.Base(p)))
		}
	}
	return ""
}

// deriveOriginalProfileName mendapatkan nama profile original yang valid
func deriveOriginalProfileName(originalProfileName string, profileInfo *domain.ProfileInfo, originalProfileInfo *domain.ProfileInfo) string {
	if v := strings.TrimSpace(connection.TrimProfileSuffix(originalProfileName)); v != "" {
		return v
	}
	if originalProfileInfo != nil {
		if v := strings.TrimSpace(connection.TrimProfileSuffix(originalProfileInfo.Name)); v != "" {
			return v
		}
	}
	if profileInfo != nil {
		if p := strings.TrimSpace(profileInfo.Path); p != "" {
			return strings.TrimSpace(connection.TrimProfileSuffix(filepath.Base(p)))
		}
	}
	return ""
}

// filePathInConfigDir membangun absolute path di dalam config dir untuk nama file konfigurasi yang diberikan.
func filePathInConfigDir(configDir string, name string) string {
	return filepath.Join(configDir, validation.ProfileExt(name))
}

// CheckConfigurationNameUnique memvalidasi apakah nama konfigurasi unik.
// Digunakan untuk memastikan tidak ada konflik saat create atau edit profile.
func CheckConfigurationNameUnique(cfg *appconfig.Config, state *profilemodel.ProfileState, mode string) error {
	if state == nil || state.ProfileInfo == nil {
		return profileerrors.ErrProfileNil
	}

	state.ProfileInfo.Name = connection.TrimProfileSuffix(state.ProfileInfo.Name)

	switch mode {
	case consts.ProfileModeCreate, consts.ProfileModeClone:
		return checkCreateModeUniqueness(cfg, state)
	case consts.ProfileModeEdit:
		return checkEditModeUniqueness(cfg, state)
	default:
		return nil
	}
}

// checkCreateModeUniqueness memvalidasi uniqueness untuk mode create
func checkCreateModeUniqueness(cfg *appconfig.Config, state *profilemodel.ProfileState) error {
	abs := filePathInConfigDir(cfg.ConfigDir.DatabaseProfile, state.ProfileInfo.Name)
	if fsops.PathExists(abs) {
		return fmt.Errorf(consts.ProfileErrConfigNameExistsFmt, state.ProfileInfo.Name)
	}
	return nil
}

// checkEditModeUniqueness memvalidasi uniqueness untuk mode edit
func checkEditModeUniqueness(cfg *appconfig.Config, state *profilemodel.ProfileState) error {
	newName := deriveProfileName(state.ProfileInfo, state.OriginalProfileInfo)
	if newName == "" {
		return fmt.Errorf(consts.ProfileErrProfileNameEmptyAlt)
	}
	state.ProfileInfo.Name = newName

	original := deriveOriginalProfileName(state.OriginalProfileName, state.ProfileInfo, state.OriginalProfileInfo)
	if original != "" {
		state.OriginalProfileName = original
	}

	baseDir := cfg.ConfigDir.DatabaseProfile
	if state.ProfileInfo.Path != "" && filepath.IsAbs(state.ProfileInfo.Path) {
		baseDir = filepath.Dir(state.ProfileInfo.Path)
	}

	if original == "" {
		// Fallback terakhir: cek file yang di-derive dari newName.
		targetAbs := filepath.Join(baseDir, validation.ProfileExt(newName))
		if !fsops.PathExists(targetAbs) {
			return fmt.Errorf(consts.ProfileErrConfigFileNotFoundChooseOtherFmt, newName)
		}
		return nil
	}

	if original == newName {
		origAbs := filepath.Join(baseDir, validation.ProfileExt(original))
		// Jika kita tahu absolute path asli, itu lebih reliable.
		if state.ProfileInfo.Path != "" && filepath.IsAbs(state.ProfileInfo.Path) {
			origAbs = state.ProfileInfo.Path
		}
		if !fsops.PathExists(origAbs) {
			return fmt.Errorf(consts.ProfileErrOriginalConfigFileNotFoundFmt, original)
		}
		return nil
	}

	newAbs := filepath.Join(baseDir, validation.ProfileExt(newName))
	if fsops.PathExists(newAbs) {
		return fmt.Errorf(consts.ProfileErrTargetConfigNameExistsFmt, newName)
	}
	origAbs := filepath.Join(baseDir, validation.ProfileExt(original))
	if !fsops.PathExists(origAbs) {
		return fmt.Errorf(consts.ProfileErrOriginalConfigFileNotFoundFmt, original)
	}
	return nil
}
