package verify

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sfdbtools/internal/app/backup/model/types_backup"
	applog "sfdbtools/internal/services/log"
	"sfdbtools/internal/ui/progress"
	"sync"
	"time"
)

// CheckOptions mengontrol check mana yang dijalankan
type CheckOptions struct {
	Checksum      bool   // generate/compare checksum
	ChecksumAlgo  string // "sha256", "md5", "xxhash"
	HeaderFooter  bool   // validasi SQL header/footer (butuh streaming)
	SizeCheck     bool   // validasi minimum file size
	MinFileSize   int64  // minimum size dalam bytes
	ExpectedHash  string // jika non-empty, compare dengan hash ini
	EncryptionKey string // key untuk decrypt (jika header/footer check pada .enc file)
}

// VerifyContext menyimpan state selama pipeline verifikasi berjalan
type VerifyContext struct {
	FilePath string
	Opts     CheckOptions
	Result   *types_backup.VerificationResult
	Logger   applog.Logger
}

// VerificationStep adalah interface untuk langkah verifikasi
type VerificationStep interface {
	Execute(ctx *VerifyContext) error
}

// Engine mengorkestrasi langkah-langkah verifikasi
type Engine struct {
	steps []VerificationStep
}

// NewEngine membuat instance engine dengan langkah-langkah standar
func NewEngine() *Engine {
	return &Engine{
		steps: []VerificationStep{
			&SizeStep{},
			&ContentStep{}, // Handles both Checksum and Header/Footer efficiently
		},
	}
}

// Check menjalankan verifikasi pada satu file
func (e *Engine) Check(filePath string, opts CheckOptions, logger applog.Logger) (*types_backup.VerificationResult, error) {
	now := time.Now()
	result := &types_backup.VerificationResult{
		VerifyStatus: "passed",
		VerifiedAt:   &now,
	}

	ctx := &VerifyContext{
		FilePath: filePath,
		Opts:     opts,
		Result:   result,
		Logger:   logger,
	}

	spinner := progress.NewSpinner(fmt.Sprintf("Verifying %s", filepath.Base(filePath)))
	spinner.Start()
	defer spinner.Stop()

	for _, step := range e.steps {
		if err := step.Execute(ctx); err != nil {
			result.VerifyStatus = "failed"
			if result.FailureReason == "" {
				result.FailureReason = err.Error()
			}
			return result, err // Stop on first hard error
		}
		// Jika step menetapkan status failed (soft error, e.g. checksum mismatch), hentikan pipeline
		if result.VerifyStatus == "failed" {
			return result, nil
		}
	}

	return result, nil
}

// CheckBatch melakukan verifikasi pada batch menggunakan worker pool concurrency
func (e *Engine) CheckBatch(dirPath string, opts CheckOptions, logger applog.Logger) (map[string]*types_backup.VerificationResult, error) {
	files, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, fmt.Errorf("error reading directory: %w", err)
	}

	var backupFiles []string
	for _, f := range files {
		if !f.IsDir() && IsBackupFile(f.Name()) {
			backupFiles = append(backupFiles, filepath.Join(dirPath, f.Name()))
		}
	}

	results := make(map[string]*types_backup.VerificationResult)
	var mu sync.Mutex

	// Worker pool configuration
	maxWorkers := 4
	if len(backupFiles) < maxWorkers {
		maxWorkers = len(backupFiles)
	}

	jobs := make(chan string, len(backupFiles))
	var wg sync.WaitGroup

	// Start workers
	for i := 0; i < maxWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for path := range jobs {
				res, err := e.Check(path, opts, logger)
				mu.Lock()
				if err != nil && res == nil {
					results[path] = &types_backup.VerificationResult{
						VerifyStatus:  "failed",
						FailureReason: err.Error(),
					}
				} else {
					results[path] = res
				}
				mu.Unlock()
			}
		}()
	}

	// Queue jobs
	for _, file := range backupFiles {
		jobs <- file
	}
	close(jobs)

	// Wait for all workers
	wg.Wait()

	return results, nil
}

// CheckFromMetadata menjalankan verifikasi menggunakan info dari .meta.json
func (e *Engine) CheckFromMetadata(metaPath string, opts CheckOptions, logger applog.Logger) (*types_backup.VerificationResult, error) {
	data, err := os.ReadFile(metaPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read metadata file: %w", err)
	}

	var meta types_backup.BackupMetadata
	if err := json.Unmarshal(data, &meta); err != nil {
		return nil, fmt.Errorf("failed to parse metadata: %w", err)
	}

	backupFilePath := meta.BackupFile
	if !filepath.IsAbs(backupFilePath) {
		backupFilePath = filepath.Join(filepath.Dir(metaPath), filepath.Base(meta.BackupFile))
	}

	if opts.ExpectedHash == "" && meta.Verification != nil && meta.Verification.ChecksumHash != "" {
		opts.ExpectedHash = meta.Verification.ChecksumHash
		if opts.ChecksumAlgo == "" {
			opts.ChecksumAlgo = meta.Verification.ChecksumAlgo
		}
	}

	return e.Check(backupFilePath, opts, logger)
}

// IsBackupFile mengecek apakah file adalah backup file berdasarkan extension
func IsBackupFile(filename string) bool {
	ext := filepath.Ext(filename)
	return ext == ".sql" || ext == ".gz" || ext == ".zst" || ext == ".enc" || ext == ".zip"
}

// HasFailures mengecek apakah ada failure di batch results
func HasFailures(results map[string]*types_backup.VerificationResult) bool {
	for _, res := range results {
		if res.VerifyStatus == "failed" {
			return true
		}
	}
	return false
}
