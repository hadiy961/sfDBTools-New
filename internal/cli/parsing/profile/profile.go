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
