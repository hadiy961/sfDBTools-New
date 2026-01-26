// File : internal/app/profile/display/import_invalid_rows.go
// Deskripsi : Display ringkasan invalid rows saat import profile (paged)
// Author : Hadiyatna Muflihun
// Tanggal : 25 Januari 2026
// Last Modified : 25 Januari 2026

package display

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	profilemodel "sfdbtools/internal/app/profile/model"
	"sfdbtools/internal/shared/runtimecfg"
	"sfdbtools/internal/ui/print"
	"sfdbtools/internal/ui/prompt"
	"sfdbtools/internal/ui/table"
)

// PrintImportInvalidRowsSummary menampilkan ringkasan invalid rows (group by row) dengan paging.
// Tidak menampilkan data sensitif (password/profile_key/ssh_password).
func PrintImportInvalidRowsSummary(errs []profilemodel.ImportCellError, srcLabel string, pageSize int) {
	if runtimecfg.IsQuiet() {
		return
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	if len(errs) == 0 {
		return
	}

	perRow := map[int][]profilemodel.ImportCellError{}
	for _, e := range errs {
		perRow[e.Row] = append(perRow[e.Row], e)
	}

	rows := make([]int, 0, len(perRow))
	for rn := range perRow {
		rows = append(rows, rn)
	}
	sort.Ints(rows)

	totalErrors := len(errs)
	invalidRows := len(rows)

	print.PrintSubHeader("Invalid Rows")
	metrics := [][]string{}
	if strings.TrimSpace(srcLabel) != "" {
		metrics = append(metrics, []string{"Sumber", srcLabel})
	}
	metrics = append(metrics,
		[]string{"Invalid rows", fmt.Sprintf("%d", invalidRows)},
		[]string{"Total errors", fmt.Sprintf("%d", totalErrors)},
	)
	table.Render([]string{"Metrik", "Nilai"}, metrics)

	// Paged table
	for start := 0; start < len(rows); start += pageSize {
		end := start + pageSize
		if end > len(rows) {
			end = len(rows)
		}

		page := rows[start:end]
		tbl := make([][]string, 0, len(page))
		for _, rn := range page {
			list := perRow[rn]
			// Stabil: sort by column+message
			sort.Slice(list, func(i, j int) bool {
				a := strings.ToLower(strings.TrimSpace(list[i].Column)) + "|" + strings.ToLower(strings.TrimSpace(list[i].Message))
				b := strings.ToLower(strings.TrimSpace(list[j].Column)) + "|" + strings.ToLower(strings.TrimSpace(list[j].Message))
				return a < b
			})

			// Ringkas: tampilkan max 2 error message per row
			msgs := make([]string, 0, 2)
			for i := 0; i < len(list) && i < 2; i++ {
				col := strings.TrimSpace(list[i].Column)
				msg := strings.TrimSpace(list[i].Message)
				if col == "" {
					msgs = append(msgs, msg)
				} else {
					msgs = append(msgs, fmt.Sprintf("%s: %s", col, msg))
				}
			}
			extra := ""
			if len(list) > 2 {
				extra = fmt.Sprintf("(+%d)", len(list)-2)
			}

			tbl = append(tbl, []string{strconv.Itoa(rn), fmt.Sprintf("%d", len(list)), strings.Join(msgs, "; "), extra})
		}

		print.PrintSubHeader(fmt.Sprintf("Daftar Invalid Rows (%d-%d dari %d)", start+1, end, len(rows)))
		table.Render([]string{"Row", "Errors", "Ringkasan", "More"}, tbl)

		if end < len(rows) {
			prompt.WaitForEnter("Tekan Enter untuk lihat halaman berikutnya...")
		}
	}
}
