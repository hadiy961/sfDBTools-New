// File : internal/shared/execx/dump.go
// Deskripsi : Helper untuk resolve binary database CLI (mariadb-dump/mysqldump)
// Author : Hadiyatna Muflihun
// Tanggal : 15 Januari 2026
// Last Modified : 15 Januari 2026

package execx

import (
	"fmt"
	"os/exec"
)

// ResolvedBinary merepresentasikan binary yang berhasil di-resolve dari PATH.
// Name adalah nama command yang user kenal, Path adalah path absolut yang ditemukan.
type ResolvedBinary struct {
	Name string
	Path string
}

// ResolveMariaDBDumpOrMysqldump memilih mariadb-dump jika tersedia, dan fallback ke mysqldump.
// Digunakan untuk kompatibilitas MySQL/MariaDB di berbagai distro.
func ResolveMariaDBDumpOrMysqldump() (ResolvedBinary, error) {
	if p, err := exec.LookPath("mariadb-dump"); err == nil {
		return ResolvedBinary{Name: "mariadb-dump", Path: p}, nil
	}
	if p, err := exec.LookPath("mysqldump"); err == nil {
		return ResolvedBinary{Name: "mysqldump", Path: p}, nil
	}
	return ResolvedBinary{}, fmt.Errorf("binary dump tidak ditemukan: butuh 'mariadb-dump' atau 'mysqldump' di PATH")
}
