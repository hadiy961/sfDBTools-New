// File : internal/app/dbscan/helpers/executor.go
// Deskripsi : Helper functions untuk eksekusi scanning database (general purpose)
// Author : Hadiyatna Muflihun
// Tanggal : 16 Desember 2025
// Last Modified : 5 Januari 2026
package helpers

import (
	"context"
	"fmt"
	"time"

	dbscanmodel "sfDBTools/internal/app/dbscan/model"
	applog "sfDBTools/internal/services/log"
	"sfDBTools/pkg/database"
	"sfDBTools/pkg/helper"

	"github.com/dustin/go-humanize"
)

// ScanExecutorOptions adalah opsi untuk executor scanning
type ScanExecutorOptions struct {
	LocalScan     bool
	DisplayResult bool
	IsBackground  bool
	Logger        applog.Logger
	LocalSizes    map[string]int64
}

// ExecuteScanWithSave melakukan scanning database dan menyimpan hasilnya
func ExecuteScanWithSave(
	ctx context.Context,
	sourceClient *database.Client,
	dbNames []string,
	serverHost string,
	serverPort int,
	opts ScanExecutorOptions,
) (*dbscanmodel.ScanResult, map[string]dbscanmodel.DatabaseDetailInfo, error) {
	timer := helper.NewTimer()

	if opts.IsBackground {
		opts.Logger.Info("Memulai proses scanning database...")
		opts.Logger.Infof("Total database yang akan di-scan: %d", len(dbNames))
	}

	if len(dbNames) == 0 {
		return nil, nil, fmt.Errorf("tidak ada database untuk di-scan")
	}

	// Set and restore session max_statement_time
	opts.Logger.Info("Mengatur max_statement_time (session) untuk mencegah query lama...")
	restore, originalMaxStatementTime, err := database.WithSessionMaxStatementTime(ctx, sourceClient, 0)
	if err != nil {
		opts.Logger.Warn("Setup max_statement_time (session) gagal: " + err.Error())
	} else {
		opts.Logger.Infof("Original (session) max_statement_time: %f detik", originalMaxStatementTime)
		defer func(orig float64) {
			bctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if rerr := restore(bctx); rerr != nil {
				opts.Logger.Warn("Gagal mengembalikan max_statement_time (session): " + rerr.Error())
			} else {
				opts.Logger.Info("max_statement_time (session) berhasil dikembalikan.")
			}
		}(originalMaxStatementTime)
	}

	opts.Logger.Info("Memulai pengumpulan detail database...")

	detailsMap := make(map[string]dbscanmodel.DatabaseDetailInfo)
	successCount := 0
	failedCount := 0
	var errors []string

	// Siapkan opsi untuk override size menggunakan hasil local scan (jika ada)
	var collectOpts *DetailCollectOptions
	if opts.LocalScan && opts.LocalSizes != nil {
		sizeProvider := func(ctx context.Context, dbName string) (int64, error) {
			if sz, ok := opts.LocalSizes[dbName]; ok {
				return sz, nil
			}
			return 0, nil
		}
		collectOpts = &DetailCollectOptions{SizeProvider: sizeProvider}
	}

	detailsMap, collectErr := CollectDatabaseDetailsWithOptions(ctx, sourceClient, dbNames, opts.Logger, collectOpts, func(detail dbscanmodel.DatabaseDetailInfo) error {
		detailsMap[detail.DatabaseName] = detail

		// Jika LocalScan aktif dan ada ukuran lokal, pastikan nilai size sudah sesuai
		if opts.LocalScan && opts.LocalSizes != nil {
			if sz, ok := opts.LocalSizes[detail.DatabaseName]; ok {
				detail.SizeBytes = sz
				detail.SizeHuman = humanize.Bytes(uint64(sz))
			}
		}

		successCount++
		_ = serverHost
		_ = serverPort
		return nil
	})

	if collectErr != nil {
		opts.Logger.Errorf("Proses scanning dihentikan: %v", collectErr)
		return &dbscanmodel.ScanResult{
			TotalDatabases: len(dbNames),
			SuccessCount:   successCount,
			FailedCount:    failedCount,
			Duration:       timer.Elapsed().String(),
			Errors:         errors,
		}, detailsMap, collectErr
	}

	result := &dbscanmodel.ScanResult{
		TotalDatabases: len(dbNames),
		SuccessCount:   successCount,
		FailedCount:    failedCount,
		Duration:       timer.Elapsed().String(),
		Errors:         errors,
	}

	return result, detailsMap, nil
}
