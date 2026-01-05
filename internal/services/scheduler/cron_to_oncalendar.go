// File : internal/services/scheduler/cron_to_oncalendar.go
// Deskripsi : Konversi subset cron (5 kolom) ke systemd OnCalendar
// Author : Hadiyatna Muflihun
// Tanggal : 2026-01-02
// Last Modified :  2026-01-05
package schedulerutil

import (
	"fmt"
	"strconv"
	"strings"
)

// CronToOnCalendar mengkonversi format cron 5 kolom menjadi ekspresi systemd OnCalendar.
// Scope (subset) yang didukung untuk menjaga implementasi sederhana dan aman:
//   - "M H * * *" (harian)
//   - "*/N * * * *" (tiap N menit)
//   - "M */N * * *" (tiap N jam)
//   - "M H */N * *" (tiap N hari)
//   - "M H * * D" (mingguan; D=0-6; 0=Sun)
//   - "M H DOM * *" (bulanan; DOM=1-31)
//   - "M * * * *" (tiap jam pada menit M)
func CronToOnCalendar(cron string) (string, error) {
	fields := strings.Fields(strings.TrimSpace(cron))
	if len(fields) != 5 {
		return "", fmt.Errorf("cron harus 5 kolom, contoh: '0 2 * * *'")
	}
	minField, hourField, domField, monthField, dowField := fields[0], fields[1], fields[2], fields[3], fields[4]
	if monthField != "*" {
		return "", fmt.Errorf("cron dengan kolom month selain '*' belum didukung")
	}

	// 1) */N * * * *
	if strings.HasPrefix(minField, "*/") && hourField == "*" && domField == "*" && dowField == "*" {
		n, err := parseStep(minField)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("*:0/%d", n), nil
	}

	// 2) M * * * * (tiap jam)
	if hourField == "*" && domField == "*" && dowField == "*" {
		m, err := parseNumberInRange(minField, 0, 59, "minute")
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("*-*-* *:%02d:00", m), nil
	}

	// 3) M */N * * * (tiap N jam)
	if strings.HasPrefix(hourField, "*/") && domField == "*" && dowField == "*" {
		m, err := parseNumberInRange(minField, 0, 59, "minute")
		if err != nil {
			return "", err
		}
		n, err := parseStep(hourField)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("*-*-* 00/%d:%02d:00", n, m), nil
	}

	// 4) M H */N * * (tiap N hari)
	if strings.HasPrefix(domField, "*/") && dowField == "*" {
		m, err := parseNumberInRange(minField, 0, 59, "minute")
		if err != nil {
			return "", err
		}
		h, err := parseNumberInRange(hourField, 0, 23, "hour")
		if err != nil {
			return "", err
		}
		n, err := parseStep(domField)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("*-*-01/%d %02d:%02d:00", n, h, m), nil
	}

	// 5) M H * * D (mingguan)
	if domField == "*" && dowField != "*" {
		m, err := parseNumberInRange(minField, 0, 59, "minute")
		if err != nil {
			return "", err
		}
		h, err := parseNumberInRange(hourField, 0, 23, "hour")
		if err != nil {
			return "", err
		}
		dow, err := parseNumberInRange(dowField, 0, 6, "day-of-week")
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%s *-*-* %02d:%02d:00", cronDowToSystemd(dow), h, m), nil
	}

	// 6) M H DOM * * (bulanan)
	if domField != "*" && dowField == "*" {
		m, err := parseNumberInRange(minField, 0, 59, "minute")
		if err != nil {
			return "", err
		}
		h, err := parseNumberInRange(hourField, 0, 23, "hour")
		if err != nil {
			return "", err
		}
		dom, err := parseNumberInRange(domField, 1, 31, "day-of-month")
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("*-*-%02d %02d:%02d:00", dom, h, m), nil
	}

	// 7) M H * * * (harian)
	if domField == "*" && dowField == "*" {
		m, err := parseNumberInRange(minField, 0, 59, "minute")
		if err != nil {
			return "", err
		}
		h, err := parseNumberInRange(hourField, 0, 23, "hour")
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("*-*-* %02d:%02d:00", h, m), nil
	}

	return "", fmt.Errorf("cron pattern belum didukung: '%s'", cron)
}

func parseStep(field string) (int, error) {
	// field: */N
	parts := strings.Split(field, "/")
	if len(parts) != 2 || parts[0] != "*" {
		return 0, fmt.Errorf("format step tidak valid: '%s'", field)
	}
	n, err := strconv.Atoi(parts[1])
	if err != nil || n <= 0 {
		return 0, fmt.Errorf("nilai step tidak valid: '%s'", field)
	}
	return n, nil
}

func parseNumberInRange(field string, min int, max int, label string) (int, error) {
	v, err := strconv.Atoi(field)
	if err != nil {
		return 0, fmt.Errorf("%s harus angka: '%s'", label, field)
	}
	if v < min || v > max {
		return 0, fmt.Errorf("%s di luar range (%d-%d): %d", label, min, max, v)
	}
	return v, nil
}

func cronDowToSystemd(dow int) string {
	switch dow {
	case 0:
		return "Sun"
	case 1:
		return "Mon"
	case 2:
		return "Tue"
	case 3:
		return "Wed"
	case 4:
		return "Thu"
	case 5:
		return "Fri"
	case 6:
		return "Sat"
	default:
		return "Sun"
	}
}
