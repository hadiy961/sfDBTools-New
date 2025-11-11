// File : internal/restore/restore_result.go
// Deskripsi : Helper functions untuk building dan managing RestoreResult
// Author : Hadiyatna Muflihun
// Tanggal : 2025-11-11
// Last Modified : 2025-11-11

package restore

import (
	"context"
	"fmt"
	"sfDBTools/internal/types"
	"sfDBTools/pkg/database"
	"sfDBTools/pkg/global"
	"time"
)

// RestoreResultBuilder memudahkan building RestoreResult dengan consistent pattern
type RestoreResultBuilder struct {
	result    types.RestoreResult
	startTime time.Time
}

// NewRestoreResultBuilder membuat builder baru untuk RestoreResult
func NewRestoreResultBuilder() *RestoreResultBuilder {
	return &RestoreResultBuilder{
		result: types.RestoreResult{
			FailedDatabases: make(map[string]string),
		},
		startTime: time.Now(),
	}
}

// SetTotalDatabases sets total number of databases
func (b *RestoreResultBuilder) SetTotalDatabases(count int) *RestoreResultBuilder {
	b.result.TotalDatabases = count
	return b
}

// SetPreBackupFile sets pre-backup file path atau directory
func (b *RestoreResultBuilder) SetPreBackupFile(path string) *RestoreResultBuilder {
	b.result.PreBackupFile = path
	return b
}

// AddRestoreInfo menambahkan database restore info ke result
func (b *RestoreResultBuilder) AddRestoreInfo(info types.DatabaseRestoreInfo) *RestoreResultBuilder {
	b.result.RestoreInfo = append(b.result.RestoreInfo, info)
	return b
}

// AddSuccess menambahkan successful restore dan build info
func (b *RestoreResultBuilder) AddSuccess(info types.DatabaseRestoreInfo, duration time.Duration) *RestoreResultBuilder {
	info.Status = "success"
	info.Duration = global.FormatDuration(duration)
	info.Verified = true

	b.result.SuccessfulRestore++
	b.result.RestoreInfo = append(b.result.RestoreInfo, info)
	return b
}

// AddSuccessWithWarning menambahkan successful restore dengan warning
func (b *RestoreResultBuilder) AddSuccessWithWarning(info types.DatabaseRestoreInfo, duration time.Duration, warning string) *RestoreResultBuilder {
	info.Status = "success"
	info.Duration = global.FormatDuration(duration)
	info.Verified = true
	info.Warnings = warning

	b.result.SuccessfulRestore++
	b.result.RestoreInfo = append(b.result.RestoreInfo, info)
	b.result.Errors = append(b.result.Errors, "WARNING: "+warning)
	return b
}

// AddFailure menambahkan failed restore
func (b *RestoreResultBuilder) AddFailure(dbName string, info types.DatabaseRestoreInfo, duration time.Duration, err error) *RestoreResultBuilder {
	info.Status = "failed"
	info.Duration = global.FormatDuration(duration)
	info.ErrorMessage = err.Error()

	b.result.FailedRestore++
	b.result.FailedDatabases[dbName] = err.Error()
	b.result.Errors = append(b.result.Errors, err.Error())
	b.result.RestoreInfo = append(b.result.RestoreInfo, info)
	return b
}

// AddFailureWithPrefix menambahkan failed restore dengan error prefix
func (b *RestoreResultBuilder) AddFailureWithPrefix(dbName string, info types.DatabaseRestoreInfo, duration time.Duration, err error, prefix string) *RestoreResultBuilder {
	errorMsg := fmt.Sprintf("%s: %v", prefix, err)
	info.Status = "failed"
	info.Duration = global.FormatDuration(duration)
	info.ErrorMessage = errorMsg

	b.result.FailedRestore++
	b.result.FailedDatabases[dbName] = errorMsg
	b.result.Errors = append(b.result.Errors, errorMsg)
	b.result.RestoreInfo = append(b.result.RestoreInfo, info)
	return b
}

// AddSkipped menambahkan skipped restore (untuk dry run)
func (b *RestoreResultBuilder) AddSkipped(info types.DatabaseRestoreInfo) *RestoreResultBuilder {
	info.Status = "skipped"
	info.Duration = "0s"
	b.result.RestoreInfo = append(b.result.RestoreInfo, info)
	return b
}

// Build menyelesaikan building dan return RestoreResult dengan total time
func (b *RestoreResultBuilder) Build() types.RestoreResult {
	b.result.TotalTimeTaken = time.Since(b.startTime)
	return b.result
}

// BuildForSingleDBSuccess builds result untuk single database restore yang sukses
func BuildSingleDBSuccessResult(info types.DatabaseRestoreInfo, duration time.Duration, warning string) types.RestoreResult {
	builder := NewRestoreResultBuilder()
	builder.SetTotalDatabases(1)

	if warning != "" {
		builder.AddSuccessWithWarning(info, duration, warning)
	} else {
		builder.AddSuccess(info, duration)
	}

	return builder.Build()
}

// BuildForSingleDBFailure builds result untuk single database restore yang gagal
func BuildSingleDBFailureResult(dbName string, info types.DatabaseRestoreInfo, duration time.Duration, err error) types.RestoreResult {
	builder := NewRestoreResultBuilder()
	builder.SetTotalDatabases(1)
	builder.AddFailure(dbName, info, duration, err)
	return builder.Build()
}

// BuildForDryRun builds result untuk dry run mode
func BuildDryRunResult(databases []string, sourceFile string) types.RestoreResult {
	builder := NewRestoreResultBuilder()
	builder.SetTotalDatabases(len(databases))

	for _, dbName := range databases {
		info := buildSkippedRestoreInfo(dbName, sourceFile, dbName, 0, "")
		builder.AddSkipped(info)
	}

	return builder.Build()
}

// setupMaxStatementTimeForRestore sets up GLOBAL max_statement_time untuk restore
// Returns restore function yang harus di-defer
func (s *Service) setupMaxStatementTimeForRestore(ctx context.Context) (restoreFunc func(context.Context) error, err error) {
	restore, originalMaxStatementTime, err := database.WithGlobalMaxStatementTime(ctx, s.Client, 0)
	if err != nil {
		s.Log.Warnf("Setup GLOBAL max_statement_time gagal: %v", err)
		return nil, err
	}

	s.Log.Infof("Original GLOBAL max_statement_time: %f detik", originalMaxStatementTime)

	// Return restore function yang akan dipanggil dalam defer
	return func(ctx context.Context) error {
		if rerr := restore(ctx); rerr != nil {
			s.Log.Warnf("Gagal mengembalikan GLOBAL max_statement_time: %v", rerr)
			return rerr
		}
		s.Log.Info("GLOBAL max_statement_time berhasil dikembalikan.")
		return nil
	}, nil
}

// checkMaxAllowedPacket checks dan log max_allowed_packet value dengan warning jika terlalu kecil
func (s *Service) checkMaxAllowedPacket(ctx context.Context) {
	maxPacket, err := s.Client.GetMaxAllowedPacket(ctx)
	if err != nil {
		s.Log.Warnf("Gagal mendapatkan max_allowed_packet: %v", err)
		return
	}

	s.Log.Infof("max_allowed_packet: %d bytes (%.2f MB)", maxPacket, float64(maxPacket)/1024/1024)
	if maxPacket < 16*1024*1024 { // Less than 16MB
		s.Log.Warnf("⚠ max_allowed_packet kecil (< 16MB), kemungkinan ada packet size issue saat restore")
	}
}

// logErrorWithDetail logs error detail ke file dan return log file path
func (s *Service) logErrorWithDetail(metadata map[string]interface{}, errorMsg string, err error) string {
	logFile := s.ErrorLog.LogWithOutput(metadata, errorMsg, err)
	if logFile != "" {
		s.Log.Infof("Error details tersimpan di: %s", logFile)
	}
	return logFile
}
