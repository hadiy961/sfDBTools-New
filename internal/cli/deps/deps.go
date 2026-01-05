package deps

import (
	"sfDBTools/internal/services/log"
	"sfDBTools/internal/types"
)

// Dependencies adalah struct yang menyimpan semua dependensi global aplikasi.
type Dependencies struct {
	Config *types.Config
	Logger applog.Logger
}

// Global variable untuk menyimpan dependensi yang di-inject
var Deps *Dependencies
