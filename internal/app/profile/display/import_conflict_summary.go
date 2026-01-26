// File : internal/app/profile/display/import_conflict_summary.go
// Deskripsi : Display ringkasan konflik nama profile saat import (paged)
// Author : Hadiyatna Muflihun
// Tanggal : 25 Januari 2026
// Last Modified : 25 Januari 2026

package display

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"sfdbtools/internal/app/profile/merger"
	profilemodel "sfdbtools/internal/app/profile/model"
	"sfdbtools/internal/shared/fsops"
	"sfdbtools/internal/shared/runtimecfg"
	"sfdbtools/internal/ui/print"
	"sfdbtools/internal/ui/prompt"
	"sfdbtools/internal/ui/table"
)

// PrintImportConflictSummary menampilkan ringkasan konflik nama profile sebelum pre-save checks.
// Tidak menampilkan data sensitif.
func PrintImportConflictSummary(conflicts []profilemodel.ImportRow, baseDir string, pageSize int) {
	if runtimecfg.IsQuiet() {
		return
	}
	if len(conflicts) == 0 {
		return
	}
	if pageSize <= 0 {
		pageSize = 20
	}

	sort.Slice(conflicts, func(i, j int) bool { return conflicts[i].RowNum < conflicts[j].RowNum })

	print.PrintSubHeader("Konflik Nama Profile")
	metrics := [][]string{{"Total konflik", fmt.Sprintf("%d", len(conflicts))}}
	table.Render([]string{"Metrik", "Nilai"}, metrics)

	for start := 0; start < len(conflicts); start += pageSize {
		end := start + pageSize
		if end > len(conflicts) {
			end = len(conflicts)
		}
		page := conflicts[start:end]

		rows := make([][]string, 0, len(page))
		for _, r := range page {
			name := strings.TrimSpace(r.PlannedName)
			if name == "" {
				name = strings.TrimSpace(r.Name)
			}
			status := "duplicate"
			if baseDir != "" && name != "" {
				abs := baseDir + "/" + merger.BuildProfileFileName(name)
				if fsops.PathExists(abs) {
					status = "exists"
				}
			}
			rows = append(rows, []string{strconv.Itoa(r.RowNum), safeName(name), status})
		}

		print.PrintSubHeader(fmt.Sprintf("Daftar Konflik (%d-%d dari %d)", start+1, end, len(conflicts)))
		table.Render([]string{"Row", "Name", "Jenis"}, rows)
		if end < len(conflicts) {
			prompt.WaitForEnter("Tekan Enter untuk lihat halaman konflik berikutnya...")
		}
	}
}
