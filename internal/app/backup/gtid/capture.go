package gtid

import (
	"context"

	applog "sfdbtools/internal/services/log"
	"sfdbtools/pkg/database"
)

// Capture fetches GTID info when enabled.
func Capture(ctx context.Context, client *database.Client, log applog.Logger, enabled bool) (*GTIDInfo, error) {
	if !enabled {
		return nil, nil
	}

	log.Info("Mengambil informasi GTID sebelum backup...")
	gtidInfo, err := GetFullGTIDInfo(ctx, client)
	if err != nil {
		log.Warnf("Gagal mendapatkan GTID: %v", err)
		return nil, nil
	}

	log.Infof("GTID berhasil diambil: File=%s, Pos=%d", gtidInfo.MasterLogFile, gtidInfo.MasterLogPos)
	return gtidInfo, nil
}
