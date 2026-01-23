package gtid

import (
	"context"

	applog "sfdbtools/internal/services/log"
	"sfdbtools/internal/shared/database"
)

// Capture fetches GTID info ketika fitur diaktifkan.
//
// Catatan timing:
//   - Kapan GTID di-capture (sebelum/after dump) adalah tanggung jawab caller.
//   - sfdbtools menggunakan GTID terutama untuk dicatat ke metadata backup (.meta.json).
//     Karena metadata ditulis saat dump selesai, GTID yang ingin tersimpan ke metadata
//     harus sudah tersedia sebelum eksekusi dump dimulai.
func Capture(ctx context.Context, client *database.Client, log applog.Logger, enabled bool) (*GTIDInfo, error) {
	if !enabled {
		return nil, nil
	}

	log.Info("Mengambil informasi GTID...")
	gtidInfo, err := GetFullGTIDInfo(ctx, client)
	if err != nil {
		log.Warnf("Gagal mendapatkan GTID: %v", err)
		return nil, nil
	}

	log.Infof("GTID berhasil diambil: File=%s, Pos=%d", gtidInfo.MasterLogFile, gtidInfo.MasterLogPos)
	return gtidInfo, nil
}
