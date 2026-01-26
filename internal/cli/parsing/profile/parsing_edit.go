package profile

import (
	"fmt"
	"os"
	profilemodel "sfdbtools/internal/app/profile/model"
	parsingcommon "sfdbtools/internal/cli/parsing/common"
	resolver "sfdbtools/internal/cli/resolver"
	"sfdbtools/internal/domain"
	"sfdbtools/internal/shared/consts"
	"strings"

	"github.com/spf13/cobra"
)

// ParsingEditProfile parses flags for the profile edit command and returns ProfileEditOptions
func ParsingEditProfile(cmd *cobra.Command) (*profilemodel.ProfileEditOptions, error) {
	filePath := resolver.GetStringFlagOrEnv(cmd, "profile", consts.ENV_SOURCE_PROFILE)
	newName := resolver.GetStringFlagOrEnv(cmd, "new-name", "")
	name := resolver.GetStringFlagOrEnv(cmd, "profile", "")

	key, keySource, err := parsingcommon.ResolveEncryptionKey(cmd, consts.ENV_TARGET_PROFILE_KEY, consts.ENV_SOURCE_PROFILE_KEY)
	if err != nil {
		return nil, err
	}

	newKey, err := resolver.GetSecretStringFlagOrEnv(cmd, "new-profile-key", "")
	if err != nil {
		return nil, err
	}
	newKeySource := ""
	if strings.TrimSpace(newKey) != "" {
		newKeySource = "flag"
	}

	interactive := parsingcommon.IsInteractiveMode()
	dbConfig := parsingcommon.ParseDBConfig(cmd)
	sshConfig := parsingcommon.ParseSSHConfig(cmd)

	// Jangan default port saat edit: port=0 berarti tidak override snapshot.
	// Namun jika user/env secara eksplisit memberi nilai di luar range, fail-fast.
	portExplicit := cmd.Flags().Changed("port") || strings.TrimSpace(os.Getenv(consts.ENV_TARGET_DB_PORT)) != ""
	if portExplicit {
		if dbConfig.Port <= 0 || dbConfig.Port > 65535 {
			return nil, fmt.Errorf("port database tidak valid: %d (range 1-65535)", dbConfig.Port)
		}
	}

	if !interactive {
		missing := make([]string, 0, 2)
		if strings.TrimSpace(filePath) == "" {
			missing = append(missing, "--profile")
		}
		if strings.TrimSpace(key) == "" {
			missing = append(missing, "--profile-key / ENV "+consts.ENV_TARGET_PROFILE_KEY+" atau "+consts.ENV_SOURCE_PROFILE_KEY)
		}

		if err := parsingcommon.ValidateNonInteractive(interactive, missing,
			"Contoh: sfdbtools profile edit --quiet --profile <nama-file> --profile-key <key> [--host <host>] [--port 3306] [--user <user>] [--password <pw>]"); err != nil {
			return nil, err
		}
	}

	return &profilemodel.ProfileEditOptions{
		ProfileInfo: domain.ProfileInfo{
			Path:             filePath,
			Name:             name,
			EncryptionKey:    key,
			EncryptionSource: keySource,
			DBInfo:           dbConfig,
			SSHTunnel:        sshConfig,
		},
		NewName:             newName,
		Interactive:         interactive,
		NewProfileKey:       newKey,
		NewProfileKeySource: newKeySource,
	}, nil
}
