// File : internal/restore/custom_parser.go
// Deskripsi : Parser untuk format paste account detail (SFCola)
// Author : Hadiyatna Muflihun
// Tanggal : 24 Desember 2025

package restore

import (
	"fmt"
	"strings"
)

type sfColaAccountExtract struct {
	Database      string
	DatabaseDmart string
	UserAdmin     string
	PassAdmin     string
	UserFin       string
	PassFin       string
	UserUser      string
	PassUser      string
}

func parseSFColaAccountDetail(text string) (sfColaAccountExtract, error) {
	var out sfColaAccountExtract

	// Normalize line endings
	text = strings.ReplaceAll(text, "\r\n", "\n")
	lines := strings.Split(text, "\n")

	getKeyVal := func(line string) (string, string, bool) {
		line = strings.TrimSpace(line)
		if line == "" {
			return "", "", false
		}
		idx := strings.Index(line, ":")
		if idx < 0 {
			return "", "", false
		}
		key := strings.TrimSpace(line[:idx])
		val := strings.TrimSpace(line[idx+1:])
		return strings.ToLower(key), val, true
	}

	for _, line := range lines {
		k, v, ok := getKeyVal(line)
		if !ok {
			continue
		}
		switch k {
		case "database":
			out.Database = v
		case "database dmart":
			out.DatabaseDmart = v
		case "user admin":
			out.UserAdmin = v
		case "pass admin":
			out.PassAdmin = v
		case "user fin":
			out.UserFin = v
		case "pass fin":
			out.PassFin = v
		case "user user":
			out.UserUser = v
		case "pass user":
			out.PassUser = v
		}
	}

	missing := make([]string, 0, 8)
	if strings.TrimSpace(out.Database) == "" {
		missing = append(missing, "Database")
	}
	if strings.TrimSpace(out.DatabaseDmart) == "" {
		missing = append(missing, "Database DMART")
	}
	if strings.TrimSpace(out.UserAdmin) == "" {
		missing = append(missing, "User Admin")
	}
	if strings.TrimSpace(out.PassAdmin) == "" {
		missing = append(missing, "Pass Admin")
	}
	if strings.TrimSpace(out.UserFin) == "" {
		missing = append(missing, "User Fin")
	}
	if strings.TrimSpace(out.PassFin) == "" {
		missing = append(missing, "Pass Fin")
	}
	if strings.TrimSpace(out.UserUser) == "" {
		missing = append(missing, "User User")
	}
	if strings.TrimSpace(out.PassUser) == "" {
		missing = append(missing, "Pass User")
	}
	if len(missing) > 0 {
		return sfColaAccountExtract{}, fmt.Errorf("account detail tidak lengkap, field hilang: %s", strings.Join(missing, ", "))
	}

	return out, nil
}
