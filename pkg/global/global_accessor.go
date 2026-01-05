package global

import (
	"sfDBTools/internal/services/config"
	"sfDBTools/internal/services/log"
	appdeps "sfDBTools/internal/cli/deps"
)

// GetLogger adalah helper untuk mengakses logger dari package lain
func GetLogger() applog.Logger {
	if appdeps.Deps == nil {
		return nil
	}
	return appdeps.Deps.Logger
}

// GetConfig adalah helper untuk mengakses config dari package lain
func GetConfig() *appconfig.Config {
	if appdeps.Deps == nil {
		return nil
	}
	return appdeps.Deps.Config
}
