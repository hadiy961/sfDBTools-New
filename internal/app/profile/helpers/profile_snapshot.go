// File : internal/app/profile/helpers/profile_snapshot.go
// Deskripsi : Helper untuk load snapshot profile + metadata file (untuk show/edit)
// Author : Hadiyatna Muflihun
// Tanggal : 14 Januari 2026
// Last Modified : 14 Januari 2026

package helpers

import (
	"sfdbtools/internal/app/profile/helpers/snapshot"
	"sfdbtools/internal/domain"
)

type SnapshotLoadOptions = snapshot.SnapshotLoadOptions

// LoadProfileSnapshotFromPath membaca file profile, mencoba dekripsi+parse, lalu mengembalikan snapshot
// yang selalu berisi metadata file (size/mtime) dan baseline value (jika load berhasil).
// Jika load gagal, snapshot tetap dikembalikan (dengan DBInfo kosong) bersama error.
func LoadProfileSnapshotFromPath(opts SnapshotLoadOptions) (*domain.ProfileInfo, error) {
	return snapshot.LoadProfileSnapshotFromPath(opts)
}
