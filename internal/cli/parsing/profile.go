package parsing

import (
	"fmt"
	"os"
	profilemodel "sfdbtools/internal/app/profile/model"
	"sfdbtools/internal/domain"
	applog "sfdbtools/internal/services/log"
	"sfdbtools/pkg/consts"
	"sfdbtools/pkg/helper"
	"sfdbtools/pkg/runtimecfg"
	"sfdbtools/pkg/validation"
	"strings"

	"github.com/mattn/go-isatty"
	"github.com/spf13/cobra"
)

// ParsingProfile
func ParsingCreateProfile(cmd *cobra.Command, logger applog.Logger) (*profilemodel.ProfileCreateOptions, error) {
	host := helper.GetStringFlagOrEnv(cmd, "host", consts.ENV_TARGET_DB_HOST)
	port := helper.GetIntFlagOrEnv(cmd, "port", consts.ENV_TARGET_DB_PORT)
	user := helper.GetStringFlagOrEnv(cmd, "user", consts.ENV_TARGET_DB_USER)
	password := helper.GetStringFlagOrEnv(cmd, "password", consts.ENV_TARGET_DB_PASSWORD)
	name := helper.GetStringFlagOrEnv(cmd, "profile", "")
	outputDir := helper.GetStringFlagOrEnv(cmd, "output-dir", "")
	key, err := helper.GetSecretStringFlagOrEnv(cmd, "profile-key", consts.ENV_TARGET_PROFILE_KEY)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(key) == "" {
		key, err = helper.GetSecretStringFlagOrEnv(cmd, "profile-key", consts.ENV_SOURCE_PROFILE_KEY)
		if err != nil {
			return nil, err
		}
	}
	interactive := !(runtimecfg.IsQuiet() || runtimecfg.IsDaemon()) &&
		isatty.IsTerminal(os.Stdin.Fd()) && isatty.IsTerminal(os.Stdout.Fd())

	sshEnabled := helper.GetBoolFlagOrEnv(cmd, "ssh", "")
	sshHost := helper.GetStringFlagOrEnv(cmd, "ssh-host", "")
	sshPort := helper.GetIntFlagOrEnv(cmd, "ssh-port", "")
	sshUser := helper.GetStringFlagOrEnv(cmd, "ssh-user", "")
	sshPassword := helper.GetStringFlagOrEnv(cmd, "ssh-password", "")
	sshIdentityFile := helper.GetStringFlagOrEnv(cmd, "ssh-identity-file", "")
	sshLocalPort := helper.GetIntFlagOrEnv(cmd, "ssh-local-port", "")
	if sshPort == 0 {
		sshPort = 22
	}

	if !interactive {
		missing := make([]string, 0, 4)
		if strings.TrimSpace(name) == "" {
			missing = append(missing, "--profile")
		}
		if strings.TrimSpace(host) == "" {
			missing = append(missing, "--host / ENV "+consts.ENV_TARGET_DB_HOST)
		}
		if strings.TrimSpace(user) == "" {
			missing = append(missing, "--user / ENV "+consts.ENV_TARGET_DB_USER)
		}
		if strings.TrimSpace(password) == "" {
			missing = append(missing, "--password / ENV "+consts.ENV_TARGET_DB_PASSWORD)
		}
		if strings.TrimSpace(key) == "" {
			missing = append(missing, "--profile-key / ENV "+consts.ENV_TARGET_PROFILE_KEY+" atau "+consts.ENV_SOURCE_PROFILE_KEY)
		}
		if len(missing) > 0 {
			return nil, fmt.Errorf(
				"mode non-interaktif (--quiet): parameter wajib belum diisi (%s). Contoh: sfdbtools profile create --quiet --profile <nama> --host <host> --user <user> --password <pw> --profile-key <key> [--port 3306]: %w",
				strings.Join(missing, ", "),
				validation.ErrNonInteractive,
			)
		}
		// Port boleh default.
		if port == 0 {
			port = 3306
		}
	} else {
		// Interactive: biarkan nilai kosong agar wizard bisa menanyakan (tanpa duplikasi input).
		// Default nilai akan diberikan di layer wizard saat prompt.
	}

	profileOptions := &profilemodel.ProfileCreateOptions{
		ProfileInfo: domain.ProfileInfo{
			Name:          name,
			EncryptionKey: key,
			DBInfo: domain.DBInfo{
				Host:     host,
				Port:     port,
				User:     user,
				Password: password,
			},
			SSHTunnel: domain.SSHTunnelConfig{
				Enabled:      sshEnabled,
				Host:         sshHost,
				Port:         sshPort,
				User:         sshUser,
				Password:     sshPassword,
				IdentityFile: sshIdentityFile,
				LocalPort:    sshLocalPort,
			},
		},
		OutputDir:   outputDir,
		Interactive: interactive,
	}

	return profileOptions, nil
}

// ParsingEditProfile parses flags for the profile edit command and returns ProfileEditOptions
func ParsingEditProfile(cmd *cobra.Command) (*profilemodel.ProfileEditOptions, error) {
	filePath := helper.GetStringFlagOrEnv(cmd, "profile", consts.ENV_SOURCE_PROFILE)
	newName := helper.GetStringFlagOrEnv(cmd, "new-name", "")
	host := helper.GetStringFlagOrEnv(cmd, "host", consts.ENV_TARGET_DB_HOST)
	port := helper.GetIntFlagOrEnv(cmd, "port", consts.ENV_TARGET_DB_PORT)
	user := helper.GetStringFlagOrEnv(cmd, "user", consts.ENV_TARGET_DB_USER)
	password := helper.GetStringFlagOrEnv(cmd, "password", consts.ENV_TARGET_DB_PASSWORD)
	name := helper.GetStringFlagOrEnv(cmd, "profile", "")
	key, err := helper.GetSecretStringFlagOrEnv(cmd, "profile-key", consts.ENV_TARGET_PROFILE_KEY)
	if err != nil {
		return nil, err
	}
	keySource := ""
	if strings.TrimSpace(key) != "" {
		if cmd.Flags().Changed("profile-key") {
			keySource = "flag"
		} else {
			keySource = "env"
		}
	}
	if strings.TrimSpace(key) == "" {
		key, err = helper.GetSecretStringFlagOrEnv(cmd, "profile-key", consts.ENV_SOURCE_PROFILE_KEY)
		if err != nil {
			return nil, err
		}
		if strings.TrimSpace(key) != "" {
			if cmd.Flags().Changed("profile-key") {
				keySource = "flag"
			} else {
				keySource = "env"
			}
		}
	}
	interactive := !(runtimecfg.IsQuiet() || runtimecfg.IsDaemon()) &&
		isatty.IsTerminal(os.Stdin.Fd()) && isatty.IsTerminal(os.Stdout.Fd())

	if !interactive {
		missing := make([]string, 0, 2)
		if strings.TrimSpace(filePath) == "" {
			missing = append(missing, "--profile")
		}
		if strings.TrimSpace(key) == "" {
			missing = append(missing, "--profile-key / ENV "+consts.ENV_TARGET_PROFILE_KEY+" atau "+consts.ENV_SOURCE_PROFILE_KEY)
		}
		if len(missing) > 0 {
			return nil, fmt.Errorf(
				"mode non-interaktif (--quiet): parameter wajib belum diisi (%s). Contoh: sfdbtools profile edit --quiet --profile <nama-file> --profile-key <key> [--host <host>] [--port 3306] [--user <user>] [--password <pw>]: %w",
				strings.Join(missing, ", "),
				validation.ErrNonInteractive,
			)
		}
		if port == 0 {
			port = 3306
		}
	}

	sshEnabled := helper.GetBoolFlagOrEnv(cmd, "ssh", "")
	sshHost := helper.GetStringFlagOrEnv(cmd, "ssh-host", "")
	sshPort := helper.GetIntFlagOrEnv(cmd, "ssh-port", "")
	sshUser := helper.GetStringFlagOrEnv(cmd, "ssh-user", "")
	sshPassword := helper.GetStringFlagOrEnv(cmd, "ssh-password", "")
	sshIdentityFile := helper.GetStringFlagOrEnv(cmd, "ssh-identity-file", "")
	sshLocalPort := helper.GetIntFlagOrEnv(cmd, "ssh-local-port", "")
	if sshPort == 0 {
		sshPort = 22
	}

	profileOptions := &profilemodel.ProfileEditOptions{
		ProfileInfo: domain.ProfileInfo{
			Path:             filePath,
			Name:             name,
			EncryptionKey:    key,
			EncryptionSource: keySource,
			DBInfo: domain.DBInfo{
				Host:     host,
				Port:     port,
				User:     user,
				Password: password,
			},
			SSHTunnel: domain.SSHTunnelConfig{
				Enabled:      sshEnabled,
				Host:         sshHost,
				Port:         sshPort,
				User:         sshUser,
				Password:     sshPassword,
				IdentityFile: sshIdentityFile,
				LocalPort:    sshLocalPort,
			},
		},
		NewName:     newName,
		Interactive: interactive,
	}

	return profileOptions, nil
}

// ParsingShowProfile
func ParsingShowProfile(cmd *cobra.Command) (*profilemodel.ProfileShowOptions, error) {
	filePath := helper.GetStringFlagOrEnv(cmd, "profile", "")
	key, err := helper.GetSecretStringFlagOrEnv(cmd, "profile-key", consts.ENV_TARGET_PROFILE_KEY)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(key) == "" {
		key, err = helper.GetSecretStringFlagOrEnv(cmd, "profile-key", consts.ENV_SOURCE_PROFILE_KEY)
		if err != nil {
			return nil, err
		}
	}
	RevealPassword := helper.GetBoolFlagOrEnv(cmd, "reveal-password", "")
	interactive := !(runtimecfg.IsQuiet() || runtimecfg.IsDaemon()) &&
		isatty.IsTerminal(os.Stdin.Fd()) && isatty.IsTerminal(os.Stdout.Fd())

	if !interactive {
		missing := make([]string, 0, 2)
		if strings.TrimSpace(filePath) == "" {
			missing = append(missing, "--profile")
		}
		if strings.TrimSpace(key) == "" {
			missing = append(missing, "--profile-key / ENV "+consts.ENV_TARGET_PROFILE_KEY+" atau "+consts.ENV_SOURCE_PROFILE_KEY)
		}
		if len(missing) > 0 {
			return nil, fmt.Errorf(
				"mode non-interaktif (--quiet): parameter wajib belum diisi (%s). Contoh: sfdbtools profile show --quiet --profile <nama-file> --profile-key <key> [--reveal-password]: %w",
				strings.Join(missing, ", "),
				validation.ErrNonInteractive,
			)
		}
	}

	profileOptions := &profilemodel.ProfileShowOptions{
		RevealPassword: RevealPassword,
		Interactive:    interactive,
		ProfileInfo: domain.ProfileInfo{
			Path:          filePath,
			EncryptionKey: key,
		},
	}

	return profileOptions, nil
}

// ParsingDeleteProfile parses flags for the profile delete command and returns ProfileDeleteOptions
func ParsingDeleteProfile(cmd *cobra.Command) (*profilemodel.ProfileDeleteOptions, error) {
	profilePaths := helper.GetStringSliceFlagOrEnv(cmd, "profile", consts.ENV_SOURCE_PROFILE)
	force := helper.GetBoolFlagOrEnv(cmd, "force", "")
	interactive := !(runtimecfg.IsQuiet() || runtimecfg.IsDaemon()) &&
		isatty.IsTerminal(os.Stdin.Fd()) && isatty.IsTerminal(os.Stdout.Fd())

	if !interactive {
		missing := make([]string, 0, 1)
		if len(profilePaths) == 0 {
			missing = append(missing, "--profile")
		}
		if !force {
			missing = append(missing, "--force")
		}
		if len(missing) > 0 {
			return nil, fmt.Errorf(
				"mode non-interaktif (--quiet): parameter wajib belum diisi (%s). Contoh: sfdbtools profile delete --quiet --force --profile <nama-file> [-f <nama-file-lain>]: %w",
				strings.Join(missing, ", "),
				validation.ErrNonInteractive,
			)
		}
	}

	profileOptions := &profilemodel.ProfileDeleteOptions{
		ProfileInfo: domain.ProfileInfo{
			EncryptionKey: "",
		},
		Profiles:    profilePaths,
		Force:       force,
		Interactive: interactive,
	}

	return profileOptions, nil
}
