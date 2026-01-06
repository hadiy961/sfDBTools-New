// File : internal/profile/profile_validation.go
// Deskripsi : Validasi dan pengecekan unik untuk profile
// Author : Hadiyatna Muflihun
// Tanggal : 4 Januari 2026
// Last Modified : 6 Januari 2026
package profile

import (
	"fmt"
	"os"
	"path/filepath"
	"sfdbtools/pkg/consts"
	"sfdbtools/pkg/fsops"
	"sfdbtools/pkg/helper"
	"sfdbtools/pkg/runtimecfg"
	"sfdbtools/pkg/validation"
	"strings"

	"sfdbtools/internal/domain"
	"sfdbtools/internal/ui/print"
	"sfdbtools/internal/ui/prompt"

	"github.com/AlecAivazis/survey/v2"
	"github.com/mattn/go-isatty"
)

func deriveProfileName(profileInfo *domain.ProfileInfo, originalProfileInfo *domain.ProfileInfo) string {
	if profileInfo != nil {
		if v := strings.TrimSpace(helper.TrimProfileSuffix(profileInfo.Name)); v != "" {
			return v
		}
	}
	if originalProfileInfo != nil {
		if v := strings.TrimSpace(helper.TrimProfileSuffix(originalProfileInfo.Name)); v != "" {
			return v
		}
	}
	if profileInfo != nil {
		if p := strings.TrimSpace(profileInfo.Path); p != "" {
			return strings.TrimSpace(helper.TrimProfileSuffix(filepath.Base(p)))
		}
	}
	return ""
}

func deriveOriginalProfileName(originalProfileName string, profileInfo *domain.ProfileInfo, originalProfileInfo *domain.ProfileInfo) string {
	if v := strings.TrimSpace(helper.TrimProfileSuffix(originalProfileName)); v != "" {
		return v
	}
	if originalProfileInfo != nil {
		if v := strings.TrimSpace(helper.TrimProfileSuffix(originalProfileInfo.Name)); v != "" {
			return v
		}
	}
	if profileInfo != nil {
		if p := strings.TrimSpace(profileInfo.Path); p != "" {
			return strings.TrimSpace(helper.TrimProfileSuffix(filepath.Base(p)))
		}
	}
	return ""
}

// CheckConfigurationNameUnique memvalidasi apakah nama konfigurasi unik.
func (s *Service) CheckConfigurationNameUnique(mode string) error {
	if s.ProfileInfo == nil {
		return fmt.Errorf(consts.ProfileErrProfileInfoNil)
	}
	s.ProfileInfo.Name = helper.TrimProfileSuffix(s.ProfileInfo.Name)
	switch mode {
	case consts.ProfileModeCreate:
		abs := s.filePathInConfigDir(s.ProfileInfo.Name)
		if fsops.PathExists(abs) {
			return fmt.Errorf(consts.ProfileErrConfigNameExistsFmt, s.ProfileInfo.Name)
		}
		return nil
	case consts.ProfileModeEdit:
		newName := deriveProfileName(s.ProfileInfo, s.OriginalProfileInfo)
		if newName == "" {
			return fmt.Errorf(consts.ProfileErrProfileNameEmptyAlt)
		}
		s.ProfileInfo.Name = newName

		original := deriveOriginalProfileName(s.OriginalProfileName, s.ProfileInfo, s.OriginalProfileInfo)
		if original != "" {
			s.OriginalProfileName = original
		}

		baseDir := s.Config.ConfigDir.DatabaseProfile
		if s.ProfileInfo.Path != "" && filepath.IsAbs(s.ProfileInfo.Path) {
			baseDir = filepath.Dir(s.ProfileInfo.Path)
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
			if s.ProfileInfo.Path != "" && filepath.IsAbs(s.ProfileInfo.Path) {
				origAbs = s.ProfileInfo.Path
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

func ValidateProfileInfo(p *domain.ProfileInfo) error {
	if p == nil {
		return fmt.Errorf(consts.ProfileErrProfileInfoNil)
	}
	if p.Name == "" {
		return fmt.Errorf(consts.ProfileErrProfileNameEmptyAlt)
	}
	if err := validation.ValidateProfileName(p.Name); err != nil {
		return err
	}
	if err := ValidateDBInfo(&p.DBInfo); err != nil {
		return fmt.Errorf(consts.ProfileErrValidateDBInfoFailedFmt, err)
	}
	if p.SSHTunnel.Enabled {
		if strings.TrimSpace(p.SSHTunnel.Host) == "" {
			return fmt.Errorf(consts.ProfileErrSSHTunnelHostEmpty)
		}
		if p.SSHTunnel.Port == 0 {
			p.SSHTunnel.Port = 22
		}
	}
	return nil
}

func ValidateDBInfo(db *domain.DBInfo) error {
	if db == nil {
		return fmt.Errorf(consts.ProfileErrDBInfoNil)
	}
	if db.Host == "" {
		return fmt.Errorf(consts.ProfileErrDBHostEmpty)
	}
	if db.Port <= 0 || db.Port > 65535 {
		return fmt.Errorf(consts.ProfileErrDBPortInvalidFmt, db.Port)
	}
	if db.User == "" {
		return fmt.Errorf(consts.ProfileErrDBUserEmpty)
	}
	if db.Password == "" {
		// Jangan pernah prompt saat mode non-interaktif (--quiet/--daemon) atau saat bukan TTY.
		if runtimecfg.IsQuiet() || runtimecfg.IsDaemon() || !isatty.IsTerminal(os.Stdin.Fd()) || !isatty.IsTerminal(os.Stdout.Fd()) {
			return fmt.Errorf(
				consts.ProfileErrDBPasswordRequiredNonInteractiveFmt,
				consts.ENV_TARGET_DB_PASSWORD,
				validation.ErrNonInteractive,
			)
		}
		print.PrintWarning(consts.ProfileWarnDBPasswordPrompting)
		pw, err := prompt.AskPassword(fmt.Sprintf(consts.ProfilePromptDBPasswordForUserFmt, db.User), survey.Required)
		if err != nil {
			return validation.HandleInputError(err)
		}
		db.Password = pw
	}
	return nil
}
