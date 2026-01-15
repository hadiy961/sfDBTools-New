// File : internal/services/log/null_logger.go
// Deskripsi : Null logger (no-op output) untuk menghindari nil check berulang
// Author : Hadiyatna Muflihun
// Tanggal : 14 Januari 2026
// Last Modified : 14 Januari 2026

package applog

import (
	"io"

	"github.com/sirupsen/logrus"
)

// NullLogger mengembalikan logger yang aman (tidak nil) dan tidak menulis output.
// Dipakai untuk mengganti pola `if log != nil` di berbagai tempat.
func NullLogger() Logger {
	l := logrus.New()
	l.SetOutput(io.Discard)
	// Set level paling tinggi agar pemanggilan Info/Debug tidak melakukan pekerjaan tambahan.
	l.SetLevel(logrus.PanicLevel)
	return l
}
