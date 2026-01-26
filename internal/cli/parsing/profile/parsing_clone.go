package profile

import (
	profilemodel "sfdbtools/internal/app/profile/model"
	parsingcommon "sfdbtools/internal/cli/parsing/common"
	resolver "sfdbtools/internal/cli/resolver"
	"sfdbtools/internal/shared/consts"
	"strings"

	"github.com/spf13/cobra"
)

// ParsingCloneProfile parses flags for the profile clone command and returns ProfileCloneOptions
func ParsingCloneProfile(cmd *cobra.Command) (*profilemodel.ProfileCloneOptions, error) {
	sourceProfile := resolver.GetStringFlagOrEnv(cmd, "source", "")
	targetName := resolver.GetStringFlagOrEnv(cmd, "name", "")
	targetHost := resolver.GetStringFlagOrEnv(cmd, "host", "")
	targetPort := resolver.GetIntFlagOrEnv(cmd, "port", "")
	outputDir := resolver.GetStringFlagOrEnv(cmd, "output-dir", "")

	key, _, err := parsingcommon.ResolveEncryptionKey(cmd, consts.ENV_SOURCE_PROFILE_KEY, consts.ENV_TARGET_PROFILE_KEY)
	if err != nil {
		return nil, err
	}

	newKey, err := resolver.GetSecretStringFlagOrEnv(cmd, "new-profile-key", "")
	if err != nil {
		return nil, err
	}

	interactive := parsingcommon.IsInteractiveMode()

	if !interactive {
		missing := make([]string, 0, 3)
		if strings.TrimSpace(sourceProfile) == "" {
			missing = append(missing, "--source")
		}
		if strings.TrimSpace(targetName) == "" {
			missing = append(missing, "--name")
		}
		if strings.TrimSpace(key) == "" {
			missing = append(missing, "--profile-key / ENV "+consts.ENV_SOURCE_PROFILE_KEY+" atau "+consts.ENV_TARGET_PROFILE_KEY)
		}

		if err := parsingcommon.ValidateNonInteractive(interactive, missing,
			"Contoh: sfdbtools profile clone --quiet --source <profile-sumber> --name <nama-baru> --profile-key <key> [--host <host>] [--port <port>]"); err != nil {
			return nil, err
		}
	}

	return &profilemodel.ProfileCloneOptions{
		SourceProfile: sourceProfile,
		TargetName:    targetName,
		TargetHost:    targetHost,
		TargetPort:    targetPort,
		ProfileKey:    key,
		NewProfileKey: newKey,
		OutputDir:     outputDir,
		Interactive:   interactive,
	}, nil
}
