// File : internal/crypto/helpers/quiet.go
// Deskripsi : Helper functions untuk quiet mode setup
// Author : Hadiyatna Muflihun
// Tanggal : 2025-11-11
// Last Modified : 2025-11-11

package helpers

import (
	"os"
	"strings"

	"sfDBTools/internal/applog"
	"sfDBTools/pkg/consts"
)

// SetupQuietMode memeriksa environment variable SFDB_QUIET dan mengkonfigurasi logger
// untuk mengarahkan output ke stderr jika quiet mode aktif.
// Ini memastikan stdout tetap bersih untuk data output (pipeline-friendly).
//
// Parameter:
//   - logger: instance logger yang akan dikonfigurasi
//
// Return:
//   - bool: true jika quiet mode aktif, false jika tidak
func SetupQuietMode(logger applog.Logger) bool {
	v := os.Getenv(consts.ENV_QUIET)
	quiet := v != "" && v != "0" && strings.ToLower(v) != "false"

	if quiet && logger != nil {
		logger.SetOutput(os.Stderr)
	}

	return quiet
}
