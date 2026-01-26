// File : internal/app/profile/helpers/reader/source_reader.go
// Deskripsi : Reader untuk berbagai sumber import (XLSX, Google Spreadsheet)
// Author : Hadiyatna Muflihun
// Tanggal : 25 Januari 2026
// Last Modified : 25 Januari 2026

package reader

import (
	"encoding/csv"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"
)

// ImportTableResult represents hasil pembacaan table dari sumber import
type ImportTableResult struct {
	Headers  []string
	DataRows [][]string
	Warnings []string
	Source   string // Label sumber: "XLSX", "Google Spreadsheet", dll
}

// ReadImportTable membaca table dari berbagai sumber (XLSX lokal atau Google Spreadsheet)
// Ini adalah dispatcher function yang menentukan reader mana yang akan dipakai
func ReadImportTable(input, gsheetURL, sheet string, gid int) (*ImportTableResult, error) {
	if strings.TrimSpace(input) != "" {
		return ReadXLSX(input, sheet)
	}

	if strings.TrimSpace(gsheetURL) != "" {
		return ReadGoogleSheetAsCSV(gsheetURL, gid)
	}

	return nil, fmt.Errorf("sumber import kosong")
}

// ReadXLSX membaca file XLSX lokal dan mengembalikan headers + data rows
func ReadXLSX(path string, sheet string) (*ImportTableResult, error) {
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

	sheetName := strings.TrimSpace(sheet)
	if sheetName == "" {
		list := f.GetSheetList()
		if len(list) == 0 {
			return nil, fmt.Errorf("XLSX tidak punya sheet")
		}
		sheetName = list[0]
	}

	allRows, err := f.GetRows(sheetName)
	if err != nil {
		return nil, fmt.Errorf("gagal membaca sheet '%s': %w", sheetName, err)
	}
	if len(allRows) == 0 {
		return nil, fmt.Errorf("sheet '%s' kosong", sheetName)
	}

	headers := allRows[0]
	if len(headers) == 0 {
		return nil, fmt.Errorf("header kosong di sheet '%s'", sheetName)
	}

	data := [][]string{}
	if len(allRows) > 1 {
		data = allRows[1:]
	}

	// Pad rows agar minimal sepanjang header
	rows := make([][]string, 0, len(data))
	for _, r := range data {
		if len(r) < len(headers) {
			padded := make([]string, len(headers))
			copy(padded, r)
			r = padded
		}
		rows = append(rows, r)
	}

	return &ImportTableResult{
		Headers:  headers,
		DataRows: rows,
		Warnings: nil,
		Source:   "XLSX",
	}, nil
}

// ReadGoogleSheetAsCSV membaca Google Spreadsheet via CSV export URL
func ReadGoogleSheetAsCSV(sheetURL string, gid int) (*ImportTableResult, error) {
	id, err := extractSpreadsheetID(sheetURL)
	if err != nil {
		return nil, err
	}
	exportURL := fmt.Sprintf("https://docs.google.com/spreadsheets/d/%s/export?format=csv&gid=%d", id, gid)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Get(exportURL)
	if err != nil {
		return nil, fmt.Errorf("gagal download google sheet: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("gagal download google sheet: HTTP %d", resp.StatusCode)
	}

	r := csv.NewReader(resp.Body)
	r.FieldsPerRecord = -1

	records := [][]string{}
	for {
		rec, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("gagal parse CSV dari google sheet: %w", err)
		}
		records = append(records, rec)
	}
	if len(records) == 0 {
		return nil, fmt.Errorf("google sheet kosong")
	}

	headers := records[0]
	rows := [][]string{}
	if len(records) > 1 {
		rows = records[1:]
	}

	return &ImportTableResult{
		Headers:  headers,
		DataRows: rows,
		Warnings: nil,
		Source:   "Google Spreadsheet",
	}, nil
}

// extractSpreadsheetID mengekstrak spreadsheet ID dari Google Sheets URL
func extractSpreadsheetID(raw string) (string, error) {
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil {
		return "", fmt.Errorf("url google sheet tidak valid: %w", err)
	}
	path := u.Path
	// Expected: /spreadsheets/d/<id>/edit
	parts := strings.Split(path, "/")
	for i := 0; i < len(parts); i++ {
		if parts[i] == "d" && i+1 < len(parts) {
			id := strings.TrimSpace(parts[i+1])
			if id != "" {
				return id, nil
			}
		}
	}
	// Fallback: try raw string search
	needle := "/spreadsheets/d/"
	if idx := strings.Index(path, needle); idx >= 0 {
		rest := path[idx+len(needle):]
		seg := strings.SplitN(rest, "/", 2)
		if len(seg) > 0 && strings.TrimSpace(seg[0]) != "" {
			return strings.TrimSpace(seg[0]), nil
		}
	}
	return "", fmt.Errorf("tidak bisa menemukan spreadsheet id dari URL")
}
