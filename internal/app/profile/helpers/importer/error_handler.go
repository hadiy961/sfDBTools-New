// File : internal/app/profile/helpers/importer/error_handler.go
// Deskripsi : Error handling helpers untuk import workflow
// Author : Hadiyatna Muflihun
// Tanggal : 25 Januari 2026
// Last Modified : 25 Januari 2026

package importer

import (
	"fmt"
	"sort"
	"strings"

	profilemodel "sfdbtools/internal/app/profile/model"
)

const (
	defaultMaxRowsInErrorOutput   = 50
	defaultMaxErrorsPerRowInError = 4
)

// GroupErrorsByRow mengelompokkan errors berdasarkan row number
func GroupErrorsByRow(errs []profilemodel.ImportCellError) map[int][]profilemodel.ImportCellError {
	out := map[int][]profilemodel.ImportCellError{}
	for _, er := range errs {
		out[er.Row] = append(out[er.Row], er)
	}
	return out
}

// FormatImportErrors formats validation errors untuk user-friendly output
func FormatImportErrors(title string, src string, errs []profilemodel.ImportCellError) error {
	// Group per-row supaya output tidak terlalu noisy untuk sheet besar.
	perRow := GroupErrorsByRow(errs)
	rows := make([]int, 0, len(perRow))
	for rn := range perRow {
		rows = append(rows, rn)
	}
	sort.Ints(rows)

	maxRows := defaultMaxRowsInErrorOutput
	maxPerRow := defaultMaxErrorsPerRowInError
	truncated := false
	if maxRows > 0 && len(rows) > maxRows {
		rows = rows[:maxRows]
		truncated = true
	}

	b := strings.Builder{}
	b.WriteString(title)
	if strings.TrimSpace(src) != "" {
		b.WriteString(" (source: ")
		b.WriteString(src)
		b.WriteString(")")
	}
	b.WriteString(":\n")

	for _, rn := range rows {
		errsInRow := perRow[rn]
		// Stabil: sort by column+message
		sort.Slice(errsInRow, func(i, j int) bool {
			ai := strings.ToLower(strings.TrimSpace(errsInRow[i].Column)) + "|" + strings.ToLower(strings.TrimSpace(errsInRow[i].Message))
			aj := strings.ToLower(strings.TrimSpace(errsInRow[j].Column)) + "|" + strings.ToLower(strings.TrimSpace(errsInRow[j].Message))
			return ai < aj
		})

		shown := errsInRow
		more := 0
		if maxPerRow > 0 && len(errsInRow) > maxPerRow {
			shown = errsInRow[:maxPerRow]
			more = len(errsInRow) - maxPerRow
		}

		parts := make([]string, 0, len(shown))
		for _, e := range shown {
			col := strings.TrimSpace(e.Column)
			msg := strings.TrimSpace(e.Message)
			if col == "" {
				parts = append(parts, msg)
				continue
			}
			parts = append(parts, fmt.Sprintf("%s: %s", col, msg))
		}
		line := fmt.Sprintf("- row %d: %s", rn, strings.Join(parts, "; "))
		if more > 0 {
			line += fmt.Sprintf(" (+%d error lagi)", more)
		}
		b.WriteString(line)
		b.WriteString("\n")
	}

	if truncated {
		b.WriteString(fmt.Sprintf("(output dibatasi: hanya menampilkan %d baris invalid pertama)\n", maxRows))
	}

	return fmt.Errorf("%s", strings.TrimSpace(b.String()))
}

// GetInvalidRowNumbers extracts row numbers dari error list (sorted, limited)
func GetInvalidRowNumbers(errs []profilemodel.ImportCellError, max int) []int {
	perRow := GroupErrorsByRow(errs)
	rows := make([]int, 0, len(perRow))
	for rn := range perRow {
		rows = append(rows, rn)
	}
	sort.Ints(rows)
	if max > 0 && len(rows) > max {
		rows = rows[:max]
	}
	return rows
}
