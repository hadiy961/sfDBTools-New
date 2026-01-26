// File : internal/app/profile/helpers/reader/gsheet_tabs.go
// Deskripsi : Best-effort listing tab (sheet name + gid) untuk Google Spreadsheet
// Author : Hadiyatna Muflihun
// Tanggal : 25 Januari 2026
// Last Modified : 25 Januari 2026

package reader

import (
	"fmt"
	"io"
	"net/http"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

// GoogleSheetTab merepresentasikan tab di Google Spreadsheet.
type GoogleSheetTab struct {
	Name string
	GID  int
}

// ListGoogleSheetTabsBestEffort mencoba membaca daftar tab (nama + gid) dari Google Spreadsheet.
// Catatan:
// - Ini best-effort untuk sheet public (tanpa auth).
// - Jika gagal, kembalikan slice kosong dan error untuk memicu fallback input gid.
func ListGoogleSheetTabsBestEffort(sheetURL string) ([]GoogleSheetTab, error) {
	id, err := extractSpreadsheetID(sheetURL)
	if err != nil {
		return nil, err
	}

	// Pakai /edit agar mendapatkan HTML yang biasanya memuat metadata sheet.
	pageURL := fmt.Sprintf("https://docs.google.com/spreadsheets/d/%s/edit", id)
	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Get(pageURL)
	if err != nil {
		return nil, fmt.Errorf("gagal membaca metadata google sheet: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("gagal membaca metadata google sheet: HTTP %d", resp.StatusCode)
	}

	// Batas aman agar tidak membaca HTML terlalu besar.
	body, err := io.ReadAll(io.LimitReader(resp.Body, 2*1024*1024))
	if err != nil {
		return nil, fmt.Errorf("gagal membaca response google sheet: %w", err)
	}
	html := string(body)
	if strings.TrimSpace(html) == "" {
		return nil, fmt.Errorf("response google sheet kosong")
	}

	reGID := regexp.MustCompile(`\"gid\":([0-9]+)`)
	matches := reGID.FindAllStringSubmatchIndex(html, -1)
	if len(matches) == 0 {
		return nil, fmt.Errorf("gid tidak ditemukan")
	}

	tabsByGID := map[int]GoogleSheetTab{}
	for _, m := range matches {
		// submatch 1 contains gid digits
		gidStr := html[m[2]:m[3]]
		gid, err := strconv.Atoi(gidStr)
		if err != nil {
			continue
		}
		if _, ok := tabsByGID[gid]; ok {
			continue
		}

		// Cari name di sekitar lokasi match (best-effort).
		start := m[0] - 600
		if start < 0 {
			start = 0
		}
		end := m[1] + 600
		if end > len(html) {
			end = len(html)
		}
		chunk := html[start:end]

		name := ""
		reName := regexp.MustCompile(`\"name\":\"([^\"]+)\"`)
		if nm := reName.FindStringSubmatch(chunk); len(nm) == 2 {
			name = nm[1]
			if unq, uerr := strconv.Unquote("\"" + name + "\""); uerr == nil {
				name = unq
			}
			name = strings.ReplaceAll(name, "\\u0026", "&")
			name = strings.TrimSpace(name)
		}

		tabsByGID[gid] = GoogleSheetTab{Name: name, GID: gid}
	}

	if len(tabsByGID) == 0 {
		return nil, fmt.Errorf("tab tidak ditemukan")
	}

	out := make([]GoogleSheetTab, 0, len(tabsByGID))
	for _, t := range tabsByGID {
		out = append(out, t)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].GID == out[j].GID {
			return out[i].Name < out[j].Name
		}
		return out[i].GID < out[j].GID
	})
	return out, nil
}
