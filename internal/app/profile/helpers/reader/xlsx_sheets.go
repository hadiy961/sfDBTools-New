// File : internal/app/profile/helpers/reader/xlsx_sheets.go
// Deskripsi : Helper untuk membaca daftar sheet dari XLSX
// Author : Hadiyatna Muflihun
// Tanggal : 25 Januari 2026
// Last Modified : 25 Januari 2026

package reader

import (
	"fmt"
	"os"
	"strings"

	"github.com/xuri/excelize/v2"
)

// ListXLSXSheets mengembalikan daftar nama sheet dari file XLSX.
func ListXLSXSheets(path string) ([]string, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return nil, fmt.Errorf("path XLSX kosong")
	}
	fi, statErr := os.Stat(path)
	if statErr != nil {
		return nil, fmt.Errorf("file XLSX tidak ditemukan: %w", statErr)
	}
	if fi.IsDir() {
		return nil, fmt.Errorf("path XLSX adalah direktori: %s", path)
	}

	f, err := excelize.OpenFile(path)
	if err != nil {
		return nil, fmt.Errorf("gagal membuka XLSX: %w", err)
	}
	defer func() { _ = f.Close() }()

	list := f.GetSheetList()
	if len(list) == 0 {
		return nil, fmt.Errorf("XLSX tidak punya sheet")
	}

	out := make([]string, 0, len(list))
	seen := map[string]bool{}
	for _, s := range list {
		v := strings.TrimSpace(s)
		if v == "" {
			continue
		}
		key := strings.ToLower(v)
		if seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, v)
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("XLSX tidak punya sheet valid")
	}
	return out, nil
}
