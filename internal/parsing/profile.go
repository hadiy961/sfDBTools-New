package parsing

import (
	"fmt"
	"os"
	"sfDBTools/internal/applog"
	"sfDBTools/internal/types"
	"sfDBTools/pkg/consts"
	"sfDBTools/pkg/helper"
	"sfDBTools/pkg/runtimecfg"
	"sfDBTools/pkg/validation"
	"strings"

	"github.com/mattn/go-isatty"
	"github.com/spf13/cobra"
)

// ParsingProfile
func ParsingCreateProfile(cmd *cobra.Command, applog applog.Logger) (*types.ProfileCreateOptions, error) {
	host := helper.GetStringFlagOrEnv(cmd, "host", consts.ENV_TARGET_DB_HOST)
	port := helper.GetIntFlagOrEnv(cmd, "port", consts.ENV_TARGET_DB_PORT)
	user := helper.GetStringFlagOrEnv(cmd, "user", consts.ENV_TARGET_DB_USER)
	password := helper.GetStringFlagOrEnv(cmd, "password", consts.ENV_TARGET_DB_PASSWORD)
	name := helper.GetStringFlagOrEnv(cmd, "profile", "")
	outputDir := helper.GetStringFlagOrEnv(cmd, "output-dir", "")
	key := helper.GetStringFlagOrEnv(cmd, "profile-key", consts.ENV_TARGET_PROFILE_KEY)
	if strings.TrimSpace(key) == "" {
		key = helper.GetStringFlagOrEnv(cmd, "profile-key", consts.ENV_SOURCE_PROFILE_KEY)
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

	profileOptions := &types.ProfileCreateOptions{
		ProfileInfo: types.ProfileInfo{
			Name:          name,
			EncryptionKey: key,
			DBInfo: types.DBInfo{
				Host:     host,
				Port:     port,
				User:     user,
				Password: password,
			},
			SSHTunnel: types.SSHTunnelConfig{
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
func ParsingEditProfile(cmd *cobra.Command) (*types.ProfileEditOptions, error) {
	filePath := helper.GetStringFlagOrEnv(cmd, "profile", consts.ENV_SOURCE_PROFILE)
	newName := helper.GetStringFlagOrEnv(cmd, "new-name", "")
	host := helper.GetStringFlagOrEnv(cmd, "host", consts.ENV_TARGET_DB_HOST)
	port := helper.GetIntFlagOrEnv(cmd, "port", consts.ENV_TARGET_DB_PORT)
	user := helper.GetStringFlagOrEnv(cmd, "user", consts.ENV_TARGET_DB_USER)
	password := helper.GetStringFlagOrEnv(cmd, "password", consts.ENV_TARGET_DB_PASSWORD)
	name := helper.GetStringFlagOrEnv(cmd, "profile", "")
	key := helper.GetStringFlagOrEnv(cmd, "profile-key", consts.ENV_TARGET_PROFILE_KEY)
	if strings.TrimSpace(key) == "" {
		key = helper.GetStringFlagOrEnv(cmd, "profile-key", consts.ENV_SOURCE_PROFILE_KEY)
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

	profileOptions := &types.ProfileEditOptions{
		ProfileInfo: types.ProfileInfo{
			Path:          filePath,
			Name:          name,
			EncryptionKey: key,
			DBInfo: types.DBInfo{
				Host:     host,
				Port:     port,
				User:     user,
				Password: password,
			},
			SSHTunnel: types.SSHTunnelConfig{
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
func ParsingShowProfile(cmd *cobra.Command) (*types.ProfileShowOptions, error) {
	filePath := helper.GetStringFlagOrEnv(cmd, "profile", "")
	key := helper.GetStringFlagOrEnv(cmd, "profile-key", consts.ENV_TARGET_PROFILE_KEY)
	if strings.TrimSpace(key) == "" {
		key = helper.GetStringFlagOrEnv(cmd, "profile-key", consts.ENV_SOURCE_PROFILE_KEY)
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

	profileOptions := &types.ProfileShowOptions{
		RevealPassword: RevealPassword,
		Interactive:    interactive,
		ProfileInfo: types.ProfileInfo{
			Path:          filePath,
			EncryptionKey: key,
		},
	}

	return profileOptions, nil
}

// ParsingDeleteProfile parses flags for the profile delete command and returns ProfileDeleteOptions
func ParsingDeleteProfile(cmd *cobra.Command) (*types.ProfileDeleteOptions, error) {
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

	profileOptions := &types.ProfileDeleteOptions{
		ProfileInfo: types.ProfileInfo{
			EncryptionKey: "",
		},
		Profiles:    profilePaths,
		Force:       force,
		Interactive: interactive,
	}

	return profileOptions, nil
}
