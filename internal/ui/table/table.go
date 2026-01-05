// File : internal/ui/table/table.go
// Deskripsi : Renderer tabel untuk UI facade
// Author : Hadiyatna Muflihun
// Tanggal : 5 Januari 2026
// Last Modified : 5 Januari 2026

package table

import (
	"os"

	"github.com/olekukonko/tablewriter"
)

// Render merender tabel ke stdout.
func Render(headers []string, rows [][]string) {
	if len(headers) == 0 || len(rows) == 0 {
		return
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.Header(headers)
	table.Bulk(rows)
	table.Render()
}
