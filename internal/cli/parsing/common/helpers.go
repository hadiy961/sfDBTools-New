// File : internal/cli/parsing/common/helpers.go
// Deskripsi : Common helper functions untuk parsing operations (shared across all parsers)
// Author : Hadiyatna Muflihun
// Tanggal : 21 Januari 2026
// Last Modified : 23 Januari 2026

package common

import (
	"fmt"
	"os"
	resolver "sfdbtools/internal/cli/resolver"
	"sfdbtools/internal/domain"
	"sfdbtools/internal/shared/consts"
	"sfdbtools/internal/shared/runtimecfg"
	"sfdbtools/internal/shared/validation"
	"strings"

	"github.com/mattn/go-isatty"
	"github.com/spf13/cobra"
)

// IsInteractiveMode mendeteksi apakah mode interaktif aktif
func IsInteractiveMode() bool {
	return !(runtimecfg.IsQuiet() || runtimecfg.IsDaemon()) &&
		isatty.IsTerminal(os.Stdin.Fd()) && isatty.IsTerminal(os.Stdout.Fd())
}

// ResolveEncryptionKey me-resolve encryption key dengan fallback env vars dan tracking source
func ResolveEncryptionKey(cmd *cobra.Command, primaryEnv, fallbackEnv string) (key string, source string, err error) {
	key, err = resolver.GetSecretStringFlagOrEnv(cmd, "profile-key", primaryEnv)
	if err != nil {
		return "", "", err
	}

	if strings.TrimSpace(key) != "" {
		if cmd.Flags().Changed("profile-key") {
			source = "flag"
		} else {
			source = "env"
		}
		return key, source, nil
	}

	// Fallback ke secondary env
	if fallbackEnv != "" {
		key, err = resolver.GetSecretStringFlagOrEnv(cmd, "profile-key", fallbackEnv)
		if err != nil {
			return "", "", err
		}
		if strings.TrimSpace(key) != "" {
			if cmd.Flags().Changed("profile-key") {
				source = "flag"
			} else {
				source = "env"
			}
		}
	}

	return key, source, nil
}

// ParseDBConfig parsing DB configuration dari flags/env.
// Catatan: fungsi ini tidak melakukan defaulting port. Port=0 dianggap "unset"
// dan akan ditangani oleh caller (mis. wizard interaktif atau create non-interaktif).
func ParseDBConfig(cmd *cobra.Command) domain.DBInfo {
	port := resolver.GetIntFlagOrEnv(cmd, "port", consts.ENV_TARGET_DB_PORT)
	return domain.DBInfo{
		Host:     resolver.GetStringFlagOrEnv(cmd, "host", consts.ENV_TARGET_DB_HOST),
		Port:     port,
		User:     resolver.GetStringFlagOrEnv(cmd, "user", consts.ENV_TARGET_DB_USER),
		Password: resolver.GetStringFlagOrEnv(cmd, "password", consts.ENV_TARGET_DB_PASSWORD),
	}
}

// ParseSSHConfig parsing SSH tunnel configuration dari flags/env
func ParseSSHConfig(cmd *cobra.Command) domain.SSHTunnelConfig {
	sshPort := resolver.GetIntFlagOrEnv(cmd, "ssh-port", "")
	if sshPort == 0 {
		sshPort = 22
	}

	return domain.SSHTunnelConfig{
		Enabled:      resolver.GetBoolFlagOrEnv(cmd, "ssh", ""),
		Host:         resolver.GetStringFlagOrEnv(cmd, "ssh-host", ""),
		Port:         sshPort,
		User:         resolver.GetStringFlagOrEnv(cmd, "ssh-user", ""),
		Password:     resolver.GetStringFlagOrEnv(cmd, "ssh-password", ""),
		IdentityFile: resolver.GetStringFlagOrEnv(cmd, "ssh-identity-file", ""),
		LocalPort:    resolver.GetIntFlagOrEnv(cmd, "ssh-local-port", ""),
	}
}

// ValidateNonInteractive validasi parameter wajib untuk non-interactive mode
func ValidateNonInteractive(interactive bool, missing []string, exampleUsage string) error {
	if !interactive && len(missing) > 0 {
		return fmt.Errorf(
			"mode non-interaktif (--quiet): parameter wajib belum diisi (%s). %s: %w",
			strings.Join(missing, ", "),
			exampleUsage,
			validation.ErrNonInteractive,
		)
	}
	return nil
}
