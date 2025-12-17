package defaultVal

import (
	"fmt"
	"sfDBTools/internal/appconfig"
	"sfDBTools/internal/types"
)

func DefaultProfileCreateOptions() types.ProfileCreateOptions {
	cfg, err := appconfig.LoadConfigFromEnv()
	if err != nil {
		fmt.Println("Error loading config:", err)
		return types.ProfileCreateOptions{
			OutputDir:   "",
			Interactive: false,
		}
	}
	return types.ProfileCreateOptions{
		OutputDir:   cfg.ConfigDir.DatabaseProfile,
		Interactive: true,
	}
}

func DefaultProfileShowOptions() types.ProfileShowOptions {
	return types.ProfileShowOptions{
		RevealPassword: false,
	}
}
