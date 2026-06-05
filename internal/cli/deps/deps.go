package deps

import (
	appconfig "sfdbtools/internal/services/config"
	applog "sfdbtools/internal/services/log"
	"sfdbtools/internal/services/notify"
)

// Dependencies adalah struct yang menyimpan semua dependensi global aplikasi.
type Dependencies struct {
	Config        *appconfig.Config
	Logger        applog.Logger
	NotifyService *notify.Service
}

// Global variable untuk menyimpan dependensi yang di-inject
var Deps *Dependencies
