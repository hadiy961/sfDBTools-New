// File : internal/app/profile/validation.go
// Deskripsi : Validasi dan pengecekan unik untuk profile
// Author : Hadiyatna Muflihun
// Tanggal : 4 Januari 2026
// Last Modified : 15 Januari 2026
package profile

import (
	"fmt"
	"path/filepath"
	"sfdbtools/internal/app/profile/connection"
	profileerrors "sfdbtools/internal/app/profile/errors"
	"sfdbtools/internal/domain"
	"sfdbtools/internal/shared/consts"
	"sfdbtools/internal/shared/fsops"
	"sfdbtools/internal/shared/validation"
	"strings"
)

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

// CheckConfigurationNameUnique memvalidasi apakah nama konfigurasi unik.
func (s *executorOps) checkConfigurationNameUnique(mode string) error {
	if s.State.ProfileInfo == nil {
		return profileerrors.ErrProfileNil
	}
	s.State.ProfileInfo.Name = connection.TrimProfileSuffix(s.State.ProfileInfo.Name)
	switch mode {
	case consts.ProfileModeCreate:
		abs := s.filePathInConfigDir(s.State.ProfileInfo.Name)
		if fsops.PathExists(abs) {
			return fmt.Errorf(consts.ProfileErrConfigNameExistsFmt, s.State.ProfileInfo.Name)
		}
		return nil
	case consts.ProfileModeEdit:
		newName := deriveProfileName(s.State.ProfileInfo, s.State.OriginalProfileInfo)
		if newName == "" {
			return fmt.Errorf(consts.ProfileErrProfileNameEmptyAlt)
		}
		s.State.ProfileInfo.Name = newName

		original := deriveOriginalProfileName(s.State.OriginalProfileName, s.State.ProfileInfo, s.State.OriginalProfileInfo)
		if original != "" {
			s.State.OriginalProfileName = original
		}

		baseDir := s.Config.ConfigDir.DatabaseProfile
		if s.State.ProfileInfo.Path != "" && filepath.IsAbs(s.State.ProfileInfo.Path) {
			baseDir = filepath.Dir(s.State.ProfileInfo.Path)
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
			if s.State.ProfileInfo.Path != "" && filepath.IsAbs(s.State.ProfileInfo.Path) {
				origAbs = s.State.ProfileInfo.Path
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
	return nil
}
