package global

import (
	appdeps "sfdbtools/internal/cli/deps"
	appconfig "sfdbtools/internal/services/config"
	applog "sfdbtools/internal/services/log"
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
