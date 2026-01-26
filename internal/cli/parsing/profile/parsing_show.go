package profile

import (
	profilemodel "sfdbtools/internal/app/profile/model"
	parsingcommon "sfdbtools/internal/cli/parsing/common"
	resolver "sfdbtools/internal/cli/resolver"
	"sfdbtools/internal/domain"
	"sfdbtools/internal/shared/consts"
	"strings"

	"github.com/spf13/cobra"
)

// ParsingShowProfile parses flags untuk profile show command
func ParsingShowProfile(cmd *cobra.Command) (*profilemodel.ProfileShowOptions, error) {
	filePath := resolver.GetStringFlagOrEnv(cmd, "profile", "")
	key, _, err := parsingcommon.ResolveEncryptionKey(cmd, consts.ENV_TARGET_PROFILE_KEY, consts.ENV_SOURCE_PROFILE_KEY)
	if err != nil {
		return nil, err
	}

	RevealPassword := resolver.GetBoolFlagOrEnv(cmd, "reveal-password", "")
	interactive := parsingcommon.IsInteractiveMode()

	if !interactive {
		missing := make([]string, 0, 2)
		if strings.TrimSpace(filePath) == "" {
			missing = append(missing, "--profile")
		}
		if strings.TrimSpace(key) == "" {
			missing = append(missing, "--profile-key / ENV "+consts.ENV_TARGET_PROFILE_KEY+" atau "+consts.ENV_SOURCE_PROFILE_KEY)
		}

		if err := parsingcommon.ValidateNonInteractive(interactive, missing,
			"Contoh: sfdbtools profile show --quiet --profile <nama-file> --profile-key <key> [--reveal-password]"); err != nil {
			return nil, err
		}
	}

	return &profilemodel.ProfileShowOptions{
		RevealPassword: RevealPassword,
		Interactive:    interactive,
		ProfileInfo: domain.ProfileInfo{
			Path:          filePath,
			EncryptionKey: key,
		},
	}, nil
}
