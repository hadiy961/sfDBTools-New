package helper

import (
	"os"
	"strconv"
	"strings"

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

func GetIntFlagOrEnv(cmd *cobra.Command, flagName, envName string) int {
	val, _ := cmd.Flags().GetInt(flagName)
	if val != 0 {
		return val
	}
	env := os.Getenv(envName)
	if env != "" {
		// ignore error, fallback ke default jika gagal
		if i, err := strconv.Atoi(env); err == nil {
			return i
		}
	}
	return val
}

func GetBoolFlagOrEnv(cmd *cobra.Command, flagName, envName string) bool {
	val, _ := cmd.Flags().GetBool(flagName)
	// Cobra default: false jika tidak di-set, jadi cek ENV jika flag tidak di-set
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
	// 1. Coba ambil nilai dari CLI Flag
	// Ignore error as we assume the flag is correctly registered.
	val, _ := cmd.Flags().GetStringSlice(flagName)

	// Jika flag diubah secara eksplisit, kembalikan nilai flag.
	if cmd.Flags().Changed(flagName) {
		return val
	}

	// 2. Cek Environment Variable (diasumsikan format comma-separated)
	env := os.Getenv(envName)
	if env != "" {
		// Pisahkan string ENV berdasarkan koma
		slice := strings.Split(env, ",")

		var cleanedSlice []string
		// Bersihkan spasi dan filter elemen kosong
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

// GetStringArrayFlagOrEnv mengambil nilai dari StringArray flag atau environment variable
func GetStringArrayFlagOrEnv(cmd *cobra.Command, flagName, envName string) []string {
	// 1. Coba ambil nilai dari CLI Flag (StringArray)
	val, _ := cmd.Flags().GetStringArray(flagName)

	// Jika flag diubah secara eksplisit, kembalikan nilai flag.
	if cmd.Flags().Changed(flagName) {
		return val
	}

	// 2. Cek Environment Variable (diasumsikan format comma-separated)
	if envName != "" {
		env := os.Getenv(envName)
		if env != "" {
			// Pisahkan string ENV berdasarkan koma
			slice := strings.Split(env, ",")

			var cleanedSlice []string
			// Bersihkan spasi dan filter elemen kosong
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

	// 3. Kembalikan nilai default dari flag (bisa empty atau dari opts)
	return val
}
