// File : internal/services/crypto/helpers/quiet.go
// Deskripsi : Helper functions untuk quiet mode setup
// Author : Hadiyatna Muflihun
// Tanggal : 2025-11-11
// Last Modified :  2026-01-05
package helpers

import (
	"os"

	"sfDBTools/internal/services/log"
	"sfDBTools/pkg/runtimecfg"
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
