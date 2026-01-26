package profile

import (
	profilemodel "sfdbtools/internal/app/profile/model"
	parsingcommon "sfdbtools/internal/cli/parsing/common"
	resolver "sfdbtools/internal/cli/resolver"
	"sfdbtools/internal/domain"
	"sfdbtools/internal/shared/consts"

	"github.com/spf13/cobra"
)

// ParsingDeleteProfile parses flags for the profile delete command and returns ProfileDeleteOptions
func ParsingDeleteProfile(cmd *cobra.Command) (*profilemodel.ProfileDeleteOptions, error) {
	profilePaths := resolver.GetStringSliceFlagOrEnv(cmd, "profile", consts.ENV_SOURCE_PROFILE)
	force := resolver.GetBoolFlagOrEnv(cmd, "force", "")
	interactive := parsingcommon.IsInteractiveMode()

	if !interactive {
		missing := make([]string, 0, 2)
		if len(profilePaths) == 0 {
			missing = append(missing, "--profile")
		}
		if !force {
			missing = append(missing, "--force")
		}

		if err := parsingcommon.ValidateNonInteractive(interactive, missing,
			"Contoh: sfdbtools profile delete --quiet --force --profile <nama-file> [-f <nama-file-lain>]"); err != nil {
			return nil, err
		}
	}

	return &profilemodel.ProfileDeleteOptions{
		ProfileInfo: domain.ProfileInfo{
			EncryptionKey: "",
		},
		Profiles:    profilePaths,
		Force:       force,
		Interactive: interactive,
	}, nil
}
