package helpers

import (
	"context"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	dbscanmodel "sfdbtools/internal/app/dbscan/model"
	applog "sfdbtools/internal/services/log"
	"sfdbtools/internal/shared/database"
)

// CollectDatabaseDetails mengumpulkan detail informasi untuk semua database secara concurrent
// dan memanggil callback onDetail setiap kali hasil untuk sebuah database tersedia.
// Jika onDetail mengembalikan error, proses akan dihentikan (early-cancel) dan error dikembalikan.
func CollectDatabaseDetails(ctx context.Context, client *database.Client, dbNames []string, logger applog.Logger, onDetail func(dbscanmodel.DatabaseDetailInfo) error) (map[string]dbscanmodel.DatabaseDetailInfo, error) {
	return CollectDatabaseDetailsWithOptions(ctx, client, dbNames, logger, nil, onDetail)
}

// CollectDatabaseDetailsWithOptions sama seperti CollectDatabaseDetails namun menerima opsi tambahan.
func CollectDatabaseDetailsWithOptions(ctx context.Context, client *database.Client, dbNames []string, logger applog.Logger, opts *DetailCollectOptions, onDetail func(dbscanmodel.DatabaseDetailInfo) error) (map[string]dbscanmodel.DatabaseDetailInfo, error) {
	const jobTimeout = 300 * time.Second // Increase overall timeout

	if len(dbNames) == 0 {
		logger.Infof("No databases to collect details for")
		return map[string]dbscanmodel.DatabaseDetailInfo{}, nil
	}

	maxWorkers := runtime.NumCPU()
	if maxWorkers < 1 {
		maxWorkers = 1
	}
	if maxWorkers > len(dbNames) {
		maxWorkers = len(dbNames)
	}

	total := len(dbNames)
	var started int32
	var completed int32
	var failed int32

	logger.Infof("Mengumpulkan detail informasi untuk %d database... workers=%d", total, maxWorkers)
	startTime := time.Now()

	runCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	jobs := make(chan DatabaseDetailJob)
	results := make(chan dbscanmodel.DatabaseDetailInfo, maxWorkers*2)

	var wg sync.WaitGroup
	for w := 0; w < maxWorkers; w++ {
		wg.Add(1)
		var sizeProvider func(context.Context, string) (int64, error)
		if opts != nil {
			sizeProvider = opts.SizeProvider
		}
		go databaseDetailWorker(runCtx, client, logger, jobs, results, &wg, jobTimeout, &started, &completed, &failed, total, sizeProvider)
	}

	go func() {
		defer close(jobs)
		for _, dbName := range dbNames {
			select {
			case <-runCtx.Done():
				return
			case jobs <- DatabaseDetailJob{DatabaseName: dbName}:
			}
		}
	}()

	go func() {
		wg.Wait()
		close(results)
	}()

	detailMap := make(map[string]dbscanmodel.DatabaseDetailInfo)
	var firstErr error
	for result := range results {
		detailMap[result.DatabaseName] = result

		if firstErr != nil {
			continue
		}
		if onDetail != nil {
			if err := onDetail(result); err != nil {
				firstErr = err
				cancel()
			}
		}
	}

	duration := time.Since(startTime)
	logger.Infof("Pengumpulan detail database selesai dalam %v", duration)

	return detailMap, firstErr
}

func databaseDetailWorker(
	ctx context.Context,
	client *database.Client,
	logger applog.Logger,
	jobs <-chan DatabaseDetailJob,
	results chan<- dbscanmodel.DatabaseDetailInfo,
	wg *sync.WaitGroup,
	timeout time.Duration,
	started *int32,
	completed *int32,
	failed *int32,
	total int,
	sizeProvider func(context.Context, string) (int64, error),
) {
	defer wg.Done()

	for job := range jobs {
		atomic.AddInt32(started, 1)

		jobStart := time.Now()
		jobCtx, cancel := context.WithTimeout(ctx, timeout)

		result := collectSingleDatabaseDetail(jobCtx, client, logger, job.DatabaseName, sizeProvider)
		results <- result

		if result.Error != "" {
			atomic.AddInt32(failed, 1)
		}
		done := atomic.AddInt32(completed, 1)

		elapsed := time.Since(jobStart)
		percent := (float64(done) / float64(total)) * 100.0
		logger.Infof("Progress: %d/%d completed (%.1f%%), failed=%d (%s), in %v", done, total, percent, atomic.LoadInt32(failed), job.DatabaseName, elapsed)

		cancel()
	}
}
