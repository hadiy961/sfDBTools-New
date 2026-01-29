// File : internal/app/dbcopy/helpers/errors.go
// Deskripsi : Helper functions untuk error handling
// Author : Hadiyatna Muflihun
// Tanggal : 28 Januari 2026
// Last Modified : 28 Januari 2026
package helpers

import (
	"strings"

	"sfdbtools/internal/ui/print"
)

// MaybePrintBackupFailureHint mencetak hint yang membantu jika backup gagal karena masalah umum
func MaybePrintBackupFailureHint(err error) {
	if err == nil {
		return
	}
	msg := strings.ToLower(err.Error())
	if strings.Contains(msg, "view '") || strings.Contains(msg, "(1356)") || strings.Contains(msg, "definer/invoker") {
		print.PrintWarning("⚠️  Backup gagal karena ada VIEW bermasalah / privilege definernya tidak valid (error 1356).")
		print.PrintWarning("    Solusi: perbaiki/recreate VIEW di source DB, atau tambahkan mysqldump args seperti --ignore-table=db.schema_view (opsional), atau gunakan --force (berisiko dump tidak lengkap).")
		print.PrintWarning("    Catatan: exclude-data tidak menyelesaikan error VIEW karena mariadb-dump tetap perlu membaca metadata VIEW.")
	}
}
