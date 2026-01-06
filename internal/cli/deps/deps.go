package deps

import (
	appconfig "sfdbtools/internal/services/config"
	applog "sfdbtools/internal/services/log"
)

// Dependencies adalah struct yang menyimpan semua dependensi global aplikasi.
type Dependencies struct {
	Config *appconfig.Config
	Logger applog.Logger
}

// Global variable untuk menyimpan dependensi yang di-inject
var Deps *Dependencies
