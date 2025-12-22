package backup

import (
	"context"
	"sfDBTools/internal/backup/helpers"
)

// CaptureAndSaveGTID mengambil dan menyimpan GTID info jika diperlukan
func (s *Service) CaptureAndSaveGTID(ctx context.Context, backupFilePath string) error {
	if !s.BackupDBOptions.CaptureGTID {
		return nil
	}

	s.Log.Info("Mengambil informasi GTID sebelum backup...")
	gtidInfo, err := helpers.GetFullGTIDInfo(ctx, s.Client)
	if err != nil {
		s.Log.Warnf("Gagal mendapatkan GTID: %v", err)
		return nil
	}

	s.Log.Infof("GTID berhasil diambil: File=%s, Pos=%d", gtidInfo.MasterLogFile, gtidInfo.MasterLogPos)

	// Simpan GTID info ke service untuk dimasukkan ke metadata nanti
	s.gtidInfo = gtidInfo

	return nil
}
