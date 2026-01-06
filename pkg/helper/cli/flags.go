package cli

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"sfdbtools/pkg/encrypt"

	"github.com/spf13/cobra"
)

func GetStringFlagOrEnv(cmd *cobra.Command, flagName, envName string) string {
	val, _ := cmd.Flags().GetString(flagName)
	if val != "" {
		return val
	}
	env := os.Getenv(envName)
	if env != "" {
		return env
	}
	return val
}

// GetSecretStringFlagOrEnv mengambil nilai secret dari flag atau env.
// - Flag selalu dianggap plaintext (tidak didekripsi meskipun bernilai "SFDBTOOLS:...")
// - Env mendukung plaintext maupun format terenkripsi "SFDBTOOLS:<payload>".
// Jika env memakai prefix namun payload invalid, akan mengembalikan error (fail-fast).
func GetSecretStringFlagOrEnv(cmd *cobra.Command, flagName, envName string) (string, error) {
	val, _ := cmd.Flags().GetString(flagName)
	if val != "" {
		return val, nil
	}
	if strings.TrimSpace(envName) == "" {
		return "", nil
	}

	v, err := encrypt.ResolveEnvSecret(envName)
	if err != nil {
		return "", fmt.Errorf("gagal membaca env %s: %w", envName, err)
	}
	return v, nil
}

func GetIntFlagOrEnv(cmd *cobra.Command, flagName, envName string) int {
	val, _ := cmd.Flags().GetInt(flagName)
	if val != 0 {
		return val
	}
	env := os.Getenv(envName)
	if env != "" {
		if i, err := strconv.Atoi(env); err == nil {
			return i
		}
	}
	return val
}

func GetBoolFlagOrEnv(cmd *cobra.Command, flagName, envName string) bool {
	val, _ := cmd.Flags().GetBool(flagName)
	if cmd.Flags().Changed(flagName) {
		return val
	}
	env := os.Getenv(envName)
	if env != "" {
		env = strings.ToLower(env)
		return env == "1" || env == "true" || env == "yes"
	}
	return val
}

func GetStringSliceFlagOrEnv(cmd *cobra.Command, flagName, envName string) []string {
	val, _ := cmd.Flags().GetStringSlice(flagName)
	if cmd.Flags().Changed(flagName) {
		return val
	}

	env := os.Getenv(envName)
	if env != "" {
		slice := strings.Split(env, ",")

		var cleanedSlice []string
		for _, s := range slice {
			trimmed := strings.TrimSpace(s)
			if trimmed != "" {
				cleanedSlice = append(cleanedSlice, trimmed)
			}
		}

		if len(cleanedSlice) > 0 {
			return cleanedSlice
		}
	}

	return nil
}

// GetStringArrayFlagOrEnv mengambil nilai dari StringArray flag atau environment variable.
func GetStringArrayFlagOrEnv(cmd *cobra.Command, flagName, envName string) []string {
	val, _ := cmd.Flags().GetStringArray(flagName)
	if cmd.Flags().Changed(flagName) {
		return val
	}

	if envName != "" {
		env := os.Getenv(envName)
		if env != "" {
			slice := strings.Split(env, ",")

			var cleanedSlice []string
			for _, s := range slice {
				trimmed := strings.TrimSpace(s)
				if trimmed != "" {
					cleanedSlice = append(cleanedSlice, trimmed)
				}
			}

			if len(cleanedSlice) > 0 {
				return cleanedSlice
			}
		}
	}

	return val
}
