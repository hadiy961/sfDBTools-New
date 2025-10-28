package global

import (
	"sfDBTools/internal/appconfig"
	"sfDBTools/internal/applog"
	"sfDBTools/internal/types"
)

// GetLogger adalah helper untuk mengakses logger dari package lain
func GetLogger() applog.Logger {
	if types.Deps == nil {
		return nil
	}
	return types.Deps.Logger
}

// GetConfig adalah helper untuk mengakses config dari package lain
func GetConfig() *appconfig.Config {
	if types.Deps == nil {
		return nil
	}
	return types.Deps.Config
}
