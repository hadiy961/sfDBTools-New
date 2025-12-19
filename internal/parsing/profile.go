package parsing

import (
	"sfDBTools/internal/applog"
	"sfDBTools/internal/types"
	"sfDBTools/pkg/consts"
	"sfDBTools/pkg/helper"

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
	key := helper.GetStringFlagOrEnv(cmd, "profile-key", consts.ENV_SOURCE_PROFILE_KEY)
	interactive := helper.GetBoolFlagOrEnv(cmd, "interactive", "")

	if port == 0 {
		port = 3306
	}
	if host == "" {
		host = "localhost"
	}
	if user == "" {
		user = "root"
	}
	if name == "" {
		name = "localhost_3306"
	}
	if outputDir == "" {
		outputDir = "."
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
	key := helper.GetStringFlagOrEnv(cmd, "profile-key", consts.ENV_SOURCE_PROFILE_KEY)
	interactive := helper.GetBoolFlagOrEnv(cmd, "interactive", "")

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
		},
		NewName:     newName,
		Interactive: interactive,
	}

	return profileOptions, nil
}

// ParsingShowProfile
func ParsingShowProfile(cmd *cobra.Command) (*types.ProfileShowOptions, error) {
	filePath := helper.GetStringFlagOrEnv(cmd, "profile", "")
	key := helper.GetStringFlagOrEnv(cmd, "profile-key", consts.ENV_SOURCE_PROFILE_KEY)
	RevealPassword := helper.GetBoolFlagOrEnv(cmd, "reveal-password", "")

	profileOptions := &types.ProfileShowOptions{
		RevealPassword: RevealPassword,
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
	key := helper.GetStringFlagOrEnv(cmd, "profile-key", consts.ENV_SOURCE_PROFILE_KEY)
	force := helper.GetBoolFlagOrEnv(cmd, "force", "")

	profileOptions := &types.ProfileDeleteOptions{
		ProfileInfo: types.ProfileInfo{
			EncryptionKey: key,
		},
		Profiles: profilePaths,
		Force:    force,
	}

	return profileOptions, nil
}
