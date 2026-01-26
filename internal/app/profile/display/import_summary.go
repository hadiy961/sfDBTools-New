// File : internal/app/profile/display/import_summary.go
// Deskripsi : Display helper untuk import plan summary
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
	"sfdbtools/internal/ui/table"
)

// PrintImportPlanSummary menampilkan ringkasan plan import sebelum eksekusi
func PrintImportPlanSummary(planned []profilemodel.ImportRow, srcLabel string) {
	if runtimecfg.IsQuiet() {
		return
	}
	print.PrintSubHeader("Rencana Import Profile")

	// Table 1: metrik ringkas
	metrics := [][]string{}
	if strings.TrimSpace(srcLabel) != "" {
		metrics = append(metrics, []string{"Sumber", srcLabel})
	}
	metrics = append(metrics, []string{"Total akan disimpan", fmt.Sprintf("%d", countPlanned(planned))})
	table.Render([]string{"Metrik", "Nilai"}, metrics)

	create := []profilemodel.ImportRow{}
	overwrite := []profilemodel.ImportRow{}
	renames := []profilemodel.ImportRow{}
	skips := []profilemodel.ImportRow{}
	skipByReason := map[string][]profilemodel.ImportRow{}

	for _, r := range planned {
		if r.Skip {
			skips = append(skips, r)
			reason := strings.ToLower(strings.TrimSpace(r.SkipReason))
			if reason == "" {
				reason = profilemodel.ImportSkipReasonUnknown
			}
			skipByReason[reason] = append(skipByReason[reason], r)
			continue
		}
		switch strings.ToLower(strings.TrimSpace(r.PlanAction)) {
		case profilemodel.ImportPlanOverwrite:
			overwrite = append(overwrite, r)
		case profilemodel.ImportPlanRename:
			renames = append(renames, r)
		default:
			create = append(create, r)
		}
	}

	// Table 2: ringkasan aksi + row numbers
	sortByRowNum(create)
	sortByRowNum(overwrite)
	sortByRowNum(renames)
	sortByRowNum(skips)

	actions := [][]string{
		{"Create", fmt.Sprintf("%d", len(create)), joinRowNums(create), ""},
		{"Overwrite", fmt.Sprintf("%d", len(overwrite)), joinRowNums(overwrite), "menimpa file existing"},
		{"Rename", fmt.Sprintf("%d", len(renames)), joinRowNums(renames), "lihat tabel rename mapping"},
		{"Skip", fmt.Sprintf("%d", len(skips)), joinRowNums(skips), "lihat tabel alasan skip"},
	}
	table.Render([]string{"Aksi", "Jumlah", "Rows", "Catatan"}, actions)

	// Table 3 (opsional): rename mapping
	if len(renames) > 0 {
		printRenameMappingTable(renames)
	}

	// Table 4 (opsional): alasan skip
	if len(skips) > 0 {
		printSkipReasonsTable(skipByReason)
	}
}

func printRenameMappingTable(renames []profilemodel.ImportRow) {
	print.PrintSubHeader("Rename Mapping")
	rows := make([][]string, 0, len(renames))
	for _, r := range renames {
		from := strings.TrimSpace(r.RenamedFrom)
		if from == "" {
			from = strings.TrimSpace(r.Name)
		}
		rows = append(rows, []string{strconv.Itoa(r.RowNum), safeName(from), safeName(r.PlannedName)})
	}
	table.Render([]string{"Row", "Dari", "Menjadi"}, rows)
}

func printSkipReasonsTable(skipByReason map[string][]profilemodel.ImportRow) {
	print.PrintSubHeader("Alasan Skip")
	keys := make([]string, 0, len(skipByReason))
	for k := range skipByReason {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	rows := [][]string{}
	for _, k := range keys {
		list := skipByReason[k]
		if len(list) == 0 {
			continue
		}
		sortByRowNum(list)
		rows = append(rows, []string{skipReasonLabel(k), joinRowNums(list)})
	}
	table.Render([]string{"Alasan", "Rows"}, rows)
}

func skipReasonLabel(reason string) string {
	switch strings.ToLower(strings.TrimSpace(reason)) {
	case profilemodel.ImportSkipReasonInvalid:
		return "invalid-row"
	case profilemodel.ImportSkipReasonDuplicate:
		return "duplicate-name"
	case profilemodel.ImportSkipReasonConflict:
		return "conflict-skip"
	case profilemodel.ImportSkipReasonConnTest:
		return "conn-test-failed"
	default:
		return "unknown"
	}
}

func joinRowNums(rows []profilemodel.ImportRow) string {
	if len(rows) == 0 {
		return "-"
	}
	nums := make([]int, 0, len(rows))
	for _, r := range rows {
		if r.RowNum > 0 {
			nums = append(nums, r.RowNum)
		}
	}
	if len(nums) == 0 {
		return "-"
	}
	sort.Ints(nums)
	parts := make([]string, 0, len(nums))
	for _, n := range nums {
		parts = append(parts, strconv.Itoa(n))
	}
	return strings.Join(parts, ", ")
}

func countPlanned(rows []profilemodel.ImportRow) int {
	n := 0
	for _, r := range rows {
		if !r.Skip {
			n++
		}
	}
	return n
}

func safeName(name string) string {
	v := strings.TrimSpace(name)
	if v == "" {
		return "(unknown)"
	}
	return v
}

func sortByRowNum(rows []profilemodel.ImportRow) {
	sort.Slice(rows, func(i, j int) bool { return rows[i].RowNum < rows[j].RowNum })
}
