package gtid

import (
	"context"

	applog "sfdbtools/internal/services/log"
	"sfdbtools/internal/shared/database"
)

// Capture fetches GTID info when enabled.
//
// TIMING CONSIDERATION:
// Function ini dipanggil SETELAH mysqldump selesai (bukan sebelum) untuk menghindari
// TOCTOU (Time-of-check to time-of-use) race condition.
//
// Rationale:
// - mysqldump --single-transaction membuat consistent snapshot di START
// - Jika GTID di-capture SEBELUM dump, ada time window dimana write transaction
//   bisa terjadi antara GTID capture dan dump start
// - GTID capture SETELAH dump lebih aman: position akan sedikit ahead dari snapshot,
//   tapi tidak akan miss transactions yang sebenarnya ada di backup
//
// Trade-off:
// - GTID position akan include beberapa transaction yang terjadi SETELAH snapshot
// - Untuk PITR: lebih aman over-estimate daripada under-estimate
// - Untuk replication: slave akan request binlog dari position yang sedikit ahead,
//   tapi mysqld akan handle ini dengan graceful skip untuk transaction yang sudah ada
func Capture(ctx context.Context, client *database.Client, log applog.Logger, enabled bool) (*GTIDInfo, error) {
	if !enabled {
		return nil, nil
	}

	log.Info("Mengambil informasi GTID setelah backup selesai...")
	gtidInfo, err := GetFullGTIDInfo(ctx, client)
	if err != nil {
		log.Warnf("Gagal mendapatkan GTID: %v", err)
		return nil, nil
	}

	log.Infof("GTID berhasil diambil: File=%s, Pos=%d", gtidInfo.MasterLogFile, gtidInfo.MasterLogPos)
	return gtidInfo, nil
}
