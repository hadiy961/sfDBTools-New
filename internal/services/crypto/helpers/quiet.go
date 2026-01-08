// File : internal/services/crypto/helpers/quiet.go
// Deskripsi : Helper functions untuk quiet mode setup
// Author : Hadiyatna Muflihun
// Tanggal : 11 November 2025
// Last Modified : 5 Januari 2026
package helpers

import (
	"os"

	applog "sfdbtools/internal/services/log"
	"sfdbtools/internal/shared/runtimecfg"
)

// SetupQuietMode memeriksa mode quiet/daemon berbasis parameter dan mengkonfigurasi logger
// untuk mengarahkan output ke stderr jika quiet mode aktif.
// Ini memastikan stdout tetap bersih untuk data output (pipeline-friendly).
//
// Parameter:
//   - logger: instance logger yang akan dikonfigurasi
//
// Return:
//   - bool: true jika quiet mode aktif, false jika tidak
func SetupQuietMode(logger applog.Logger) bool {
	quiet := runtimecfg.IsQuiet() || runtimecfg.IsDaemon()

	if quiet && logger != nil {
		logger.SetOutput(os.Stderr)
	}

	return quiet
}
