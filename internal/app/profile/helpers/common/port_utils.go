// File : internal/app/profile/helpers/common/port_utils.go
// Deskripsi : Port validation utilities (reuse existing logic)
// Author : Hadiyatna Muflihun
// Tanggal : 25 Januari 2026
// Last Modified : 25 Januari 2026

package common

import (
	"fmt"
	"strconv"
	"strings"
)

// ParsePort parses string port value dan validates range 1-65535
// Returns parsed port value atau default jika empty, error jika invalid
func ParsePort(value string, defaultPort int, fieldName string) (int, error) {
	v := strings.TrimSpace(value)
	if v == "" {
		return defaultPort, nil
	}
	port, err := strconv.Atoi(v)
	if err != nil {
		return 0, fmt.Errorf("%s tidak valid: %s", fieldName, v)
	}
	if port <= 0 || port > 65535 {
		return 0, fmt.Errorf("%s harus 1-65535, ditemukan: %d", fieldName, port)
	}
	return port, nil
}

// ParsePortAllowZero parses string port value dan validates range 0 atau 1-65535.
// Khusus untuk field yang mengizinkan 0 (mis. ssh_local_port = auto-assign).
func ParsePortAllowZero(value string, defaultPort int, fieldName string) (int, error) {
	v := strings.TrimSpace(value)
	if v == "" {
		return defaultPort, nil
	}
	port, err := strconv.Atoi(v)
	if err != nil {
		return 0, fmt.Errorf("%s tidak valid: %s", fieldName, v)
	}
	if port == 0 {
		return 0, nil
	}
	if port < 0 || port > 65535 {
		return 0, fmt.Errorf("%s harus 0 atau 1-65535, ditemukan: %d", fieldName, port)
	}
	return port, nil
}

// ValidatePort validates port range (dipakai untuk conn-test)
func ValidatePort(port int) error {
	if port <= 0 || port > 65535 {
		return fmt.Errorf("port harus 1-65535, ditemukan: %d", port)
	}
	return nil
}
