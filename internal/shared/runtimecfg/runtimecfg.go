// File : internal/shared/runtimecfg/runtimecfg.go
// Deskripsi : Konfigurasi runtime berbasis flag (tanpa env) untuk quiet
// Author : Hadiyatna Muflihun
// Tanggal : 4 Januari 2026
// Last Modified : 23 Januari 2026

package runtimecfg

import "strings"

var quiet bool

// SetQuiet mengaktifkan quiet mode (tanpa spinner/UI noisy, console diarahkan ke stderr oleh logger).
func SetQuiet(v bool) {
	quiet = v
}

func IsQuiet() bool {
	return quiet
}

// BootstrapFromArgs mem-parsing args untuk flag global --quiet/-q.
// Parsing ini sengaja sederhana supaya bisa dipakai sebelum cobra init.
func BootstrapFromArgs(args []string) {
	if hasBoolFlag(args, "--quiet") || hasBoolFlag(args, "--quite") || hasShortFlag(args, "-q") {
		SetQuiet(true)
	}
}

func hasBoolFlag(args []string, name string) bool {
	for _, a := range args {
		if a == name {
			return true
		}
		if strings.HasPrefix(a, name+"=") {
			v := strings.TrimPrefix(a, name+"=")
			v = strings.ToLower(strings.TrimSpace(v))
			if v == "" || v == "1" || v == "true" || v == "yes" {
				return true
			}
		}
	}
	return false
}

func hasShortFlag(args []string, short string) bool {
	for _, a := range args {
		if a == short {
			return true
		}
		// Support simple bundling like -qv
		if strings.HasPrefix(a, "-") && !strings.HasPrefix(a, "--") {
			if strings.Contains(a, strings.TrimPrefix(short, "-")) {
				return true
			}
		}
	}
	return false
}
