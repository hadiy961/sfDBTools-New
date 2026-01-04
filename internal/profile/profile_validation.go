// File : internal/profile/profile_validation.go
// Deskripsi : Validasi dan pengecekan unik untuk profile
// Author : Hadiyatna Muflihun
// Tanggal : 4 Januari 2026
// Last Modified : 4 Januari 2026

package profile

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"sfDBTools/internal/types"
	"sfDBTools/pkg/consts"
	"sfDBTools/pkg/fsops"
	"sfDBTools/pkg/helper"
	"sfDBTools/pkg/input"
	"sfDBTools/pkg/runtimecfg"
	"sfDBTools/pkg/ui"
	"sfDBTools/pkg/validation"

	"github.com/AlecAivazis/survey/v2"
	"github.com/mattn/go-isatty"
)

// CheckConfigurationNameUnique memvalidasi apakah nama konfigurasi unik.
func (s *Service) CheckConfigurationNameUnique(mode string) error {
	s.ProfileInfo.Name = helper.TrimProfileSuffix(s.ProfileInfo.Name)
	switch mode {
	case consts.ProfileModeCreate:
		abs := s.filePathInConfigDir(s.ProfileInfo.Name)
		if fsops.PathExists(abs) {
			return fmt.Errorf(consts.ProfileErrConfigNameExistsFmt, s.ProfileInfo.Name)
		}
		return nil
	case consts.ProfileModeEdit:
		original := helper.TrimProfileSuffix(s.OriginalProfileName)
		newName := helper.TrimProfileSuffix(s.ProfileInfo.Name)

		baseDir := s.Config.ConfigDir.DatabaseProfile
		if s.ProfileInfo.Path != "" && filepath.IsAbs(s.ProfileInfo.Path) {
			baseDir = filepath.Dir(s.ProfileInfo.Path)
		}

		if original == "" {
			targetAbs := filepath.Join(baseDir, validation.ProfileExt(newName))
			if !fsops.PathExists(targetAbs) {
				return fmt.Errorf(consts.ProfileErrConfigFileNotFoundChooseOtherFmt, newName)
			}
			return nil
		}

		if original == newName {
			origAbs := filepath.Join(baseDir, validation.ProfileExt(original))
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

func ValidateProfileInfo(p *types.ProfileInfo) error {
	if p == nil {
		return fmt.Errorf(consts.ProfileErrProfileInfoNil)
	}
	if p.Name == "" {
		return fmt.Errorf(consts.ProfileErrProfileNameEmptyAlt)
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

func ValidateDBInfo(db *types.DBInfo) error {
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
		ui.PrintWarning(consts.ProfileWarnDBPasswordPrompting)
		pw, err := input.AskPassword(fmt.Sprintf(consts.ProfilePromptDBPasswordForUserFmt, db.User), survey.Required)
		if err != nil {
			return validation.HandleInputError(err)
		}
		db.Password = pw
	}
	return nil
}
