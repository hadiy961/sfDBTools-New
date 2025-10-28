package parsing

import (
	"sfDBTools/internal/applog"
	"sfDBTools/internal/types"
	"sfDBTools/pkg/consts"
	"sfDBTools/pkg/helper"

	"github.com/spf13/cobra"
)

// ParsingProfile
func ParsingScanAllDBOptions(cmd *cobra.Command, applog applog.Logger) (*types.ScanAllDBOptions, error) {
	filePath := helper.GetStringFlagOrEnv(cmd, "file", consts.ENV_SOURCE_PROFILE)
	key := helper.GetStringFlagOrEnv(cmd, "key", consts.ENV_SOURCE_PROFILE_KEY)
	savetodb := helper.GetBoolFlagOrEnv(cmd, "save-to-db", "")
	excludesystem := helper.GetBoolFlagOrEnv(cmd, "exclude-system", "")
	background := helper.GetBoolFlagOrEnv(cmd, "background", "")

	ScanAllDBOptions := &types.ScanAllDBOptions{
		ProfileInfo: types.ProfileInfo{
			Path:          filePath,
			EncryptionKey: key,
		},
		SaveToDB:      savetodb,
		ExcludeSystem: excludesystem,
		Background:    background,
	}

	return ScanAllDBOptions, nil
}
