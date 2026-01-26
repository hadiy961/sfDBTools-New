package profile

import (
	"fmt"
	"os"
	profilemodel "sfdbtools/internal/app/profile/model"
	parsingcommon "sfdbtools/internal/cli/parsing/common"
	resolver "sfdbtools/internal/cli/resolver"
	"sfdbtools/internal/domain"
	applog "sfdbtools/internal/services/log"
	"sfdbtools/internal/shared/consts"
	"strings"

	"github.com/spf13/cobra"
)

// ParsingCreateProfile melakukan parsing opsi untuk create profile
func ParsingCreateProfile(cmd *cobra.Command, logger applog.Logger) (*profilemodel.ProfileCreateOptions, error) {
	name := resolver.GetStringFlagOrEnv(cmd, "profile", "")
	outputDir := resolver.GetStringFlagOrEnv(cmd, "output-dir", "")

	key, _, err := parsingcommon.ResolveEncryptionKey(cmd, consts.ENV_TARGET_PROFILE_KEY, consts.ENV_SOURCE_PROFILE_KEY)
	if err != nil {
		return nil, err
	}

	interactive := parsingcommon.IsInteractiveMode()
	dbConfig := parsingcommon.ParseDBConfig(cmd)
	sshConfig := parsingcommon.ParseSSHConfig(cmd)

	// Port default hanya untuk mode non-interaktif (create). Untuk mode interaktif,
	// biarkan port=0 agar wizard akan prompt dengan default 3306.
	if !interactive {
		portExplicit := cmd.Flags().Changed("port") || strings.TrimSpace(os.Getenv(consts.ENV_TARGET_DB_PORT)) != ""
		if dbConfig.Port == 0 {
			if portExplicit {
				return nil, fmt.Errorf("port database tidak valid: 0 (gunakan 1-65535 atau omit --port/ENV %s untuk default 3306)", consts.ENV_TARGET_DB_PORT)
			}
			dbConfig.Port = 3306
		}
		if dbConfig.Port < 0 || dbConfig.Port > 65535 {
			return nil, fmt.Errorf("port database tidak valid: %d (range 1-65535)", dbConfig.Port)
		}
	}

	if !interactive {
		missing := make([]string, 0, 5)
		if strings.TrimSpace(name) == "" {
			missing = append(missing, "--profile")
		}
		if strings.TrimSpace(dbConfig.Host) == "" {
			missing = append(missing, "--host / ENV "+consts.ENV_TARGET_DB_HOST)
		}
		if strings.TrimSpace(dbConfig.User) == "" {
			missing = append(missing, "--user / ENV "+consts.ENV_TARGET_DB_USER)
		}
		if strings.TrimSpace(dbConfig.Password) == "" {
			missing = append(missing, "--password / ENV "+consts.ENV_TARGET_DB_PASSWORD)
		}
		if strings.TrimSpace(key) == "" {
			missing = append(missing, "--profile-key / ENV "+consts.ENV_TARGET_PROFILE_KEY+" atau "+consts.ENV_SOURCE_PROFILE_KEY)
		}

		if err := parsingcommon.ValidateNonInteractive(interactive, missing,
			"Contoh: sfdbtools profile create --quiet --profile <nama> --host <host> --user <user> --password <pw> --profile-key <key> [--port 3306]"); err != nil {
			return nil, err
		}
	}

	return &profilemodel.ProfileCreateOptions{
		ProfileInfo: domain.ProfileInfo{
			Name:          name,
			EncryptionKey: key,
			DBInfo:        dbConfig,
			SSHTunnel:     sshConfig,
		},
		OutputDir:   outputDir,
		Interactive: interactive,
	}, nil
}
