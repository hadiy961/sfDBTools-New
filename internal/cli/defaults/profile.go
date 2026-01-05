package defaultVal

import (
	"fmt"
	profilemodel "sfDBTools/internal/app/profile/model"
	appconfig "sfDBTools/internal/services/config"
)

func DefaultProfileCreateOptions() profilemodel.ProfileCreateOptions {
	cfg, err := appconfig.LoadConfigFromEnv()
	if err != nil {
		fmt.Println("Error loading config:", err)
		return profilemodel.ProfileCreateOptions{
			OutputDir:   "",
			Interactive: false,
		}
	}
	return profilemodel.ProfileCreateOptions{
		OutputDir:   cfg.ConfigDir.DatabaseProfile,
		Interactive: true,
	}
}

func DefaultProfileShowOptions() profilemodel.ProfileShowOptions {
	return profilemodel.ProfileShowOptions{
		RevealPassword: false,
	}
}
