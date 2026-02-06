// File : internal/backup/execution/engine.go
// Deskripsi : Main backup execution engine dengan orchestration logic
// Author : Hadiyatna Muflihun
// Tanggal : 2025-12-05
// Last Modified : 27 Januari 2026
package execution

import (
	"context"
	"fmt"
	"strings"

	"sfdbtools/internal/app/backup/gtid"
	"sfdbtools/internal/app/backup/model"
	"sfdbtools/internal/app/backup/model/types_backup"
	"sfdbtools/internal/app/backup/writer"
	profileconn "sfdbtools/internal/app/profile/connection"
	appconfig "sfdbtools/internal/services/config"
	applog "sfdbtools/internal/services/log"
	"sfdbtools/internal/shared/consts"
	"sfdbtools/internal/shared/database"
	"sfdbtools/internal/shared/errorlog"
	"sfdbtools/internal/shared/timex"
)

// StateTracker interface untuk tracking current backup file state.
type StateTracker interface {
	SetCurrentBackupFile(filePath string)
	ClearCurrentBackupFile()
	Cleanup() // Cleanup partial backup files on context cancellation
}

// UserGrantsHooks interface untuk user grants export operations.
type UserGrantsHooks interface {
	ExportUserGrantsIfNeeded(ctx context.Context, outputPath string, dbNames []string) string
	UpdateMetadataUserGrantsPath(outputPath string, userGrantsPath string, permissions string)
}

// Engine adalah core backup execution engine.
// Menghandle orchestration dari dump execution (mariadb-dump/mysqldump), retry logic, dan metadata generation.
type Engine struct {
	Log      applog.Logger
	Config   *appconfig.Config
	Options  *types_backup.BackupDBOptions
	ErrorLog *errorlog.ErrorLogger

	// Dependencies yang di-inject setelah Engine dibuat
	Client            *database.Client
	GTIDInfo          *gtid.GTIDInfo
	ExcludedDatabases []string
	State             StateTracker
	UserGrants        UserGrantsHooks
}

// New membuat Engine instance baru dengan dependencies dasar.
func New(log applog.Logger, cfg *appconfig.Config, opts *types_backup.BackupDBOptions, errLog *errorlog.ErrorLogger) *Engine {
	return &Engine{
		Log:      log,
		Config:   cfg,
		Options:  opts,
		ErrorLog: errLog,
	}
}

// WithDependencies adalah builder method untuk inject runtime dependencies.
// Method ini menghindari duplikasi setup code di caller (DRY principle).
func (e *Engine) WithDependencies(
	client *database.Client,
	gtidInfo *gtid.GTIDInfo,
	excludedDBs []string,
	state StateTracker,
	userGrants UserGrantsHooks,
) *Engine {
	e.Client = client
	e.GTIDInfo = gtidInfo
	e.ExcludedDatabases = excludedDBs
	e.State = state
	e.UserGrants = userGrants
	return e
}

// ExecuteAndBuildBackup menjalankan backup dan menghasilkan DatabaseBackupInfo.
// Includes: dump execution (mariadb-dump/mysqldump), retry logic, metadata/manifest generation.
func (e *Engine) ExecuteAndBuildBackup(
	ctx context.Context,
	cfg types_backup.BackupExecutionConfig,
) (types_backup.DatabaseBackupInfo, error) {
	timer := timex.NewTimer()
	startTime := timer.StartTime()

	// Set current backup file untuk state tracking (cleanup on cancel)
	if e.State != nil {
		e.State.SetCurrentBackupFile(cfg.OutputPath)
		defer e.State.ClearCurrentBackupFile()
	}

	// Get database version sebelum backup (fresh connection)
	dbVersion := ""
	if e.Client != nil {
		if v, err := e.Client.GetVersion(ctx); err == nil {
			dbVersion = v
			e.Log.Infof("Database version: %s", dbVersion)
		} else {
			e.Log.Warnf("Gagal mendapatkan database version: %v", err)
		}
	}

	// Build dump arguments (kompatibel mariadb-dump/mysqldump)
	if e.Options == nil {
		return types_backup.DatabaseBackupInfo{}, fmt.Errorf("ExecuteAndBuildBackup: %w", model.ErrBackupOptionsNotAvailable)
	}

	// Log mulai backup setelah database version
	if !cfg.IsMultiDB && cfg.DBName != "" {
		modeDesc := "database"
		if e.Options.Mode == consts.ModePrimary {
			modeDesc = "database primary"
		} else if e.Options.Mode == consts.ModeSecondary {
			modeDesc = "database secondary"
		}
		// ModeSingle tetap "database" (untuk companion atau single DB tanpa konteks primary/secondary)
		e.Log.Infof("Memulai backup %s: %s", modeDesc, cfg.DBName)
	}

	var dbList []string
	if cfg.IsMultiDB {
		dbList = cfg.DBList
	}

	sshTunnelEnabled := false
	if e.Options != nil && e.Options.Profile.SSHTunnel.Enabled {
		sshTunnelEnabled = true
	}

	mysqldumpArgs := BuildMysqldumpArgs(
		e.Config.Backup.MysqlDumpArgs,
		profileconn.EffectiveDBInfo(&e.Options.Profile),
		e.Options.Filter,
		dbList,
		cfg.DBName,
		cfg.TotalDBFound,
		e.Options.SkipTablesData,
		sshTunnelEnabled,
	)

	if e.Options.DryRun {
		return e.buildDryRunInfo(cfg, mysqldumpArgs, timer, startTime), nil
	}

	writeResult, _, err := e.executeWithRetry(ctx, cfg.OutputPath, mysqldumpArgs)
	if err != nil {
		e.handleBackupError(err, cfg, writeResult)
		return types_backup.DatabaseBackupInfo{}, err
	}

	return e.buildRealBackupInfo(cfg, writeResult, timer, startTime, dbVersion), nil
}

// executeWithRetry menjalankan dump dengan automatic retry pada failure yang umum.
func (e *Engine) executeWithRetry(ctx context.Context, outputPath string, args []string) (*types_backup.BackupWriteResult, []string, error) {
	// Check context sebelum execute
	if ctx.Err() != nil {
		return nil, args, fmt.Errorf("backup cancelled: %w", ctx.Err())
	}

	writeEngine := writer.New(e.Log, e.Options)

	permissions := e.Config.Backup.Output.FilePermissions
	exec := func(a []string) (*types_backup.BackupWriteResult, error) {
		return writeEngine.ExecuteMysqldumpWithPipe(ctx, a, outputPath, e.Options.Compression.Enabled, e.Options.Compression.Type, permissions)
	}

	attempts := 0
	result, err := exec(args)

	for err != nil && attempts < maxRetries {
		// Jika eksekusi gagal sebelum menghasilkan result (mis. binary tidak ada, gagal buat file),
		// jangan menimpa error asli dengan error retry yang generik.
		if result == nil {
			if ctx.Err() != nil {
				cleanupFailedBackup(outputPath, e.Log)
			}
			return nil, args, err
		}
		if ctx.Err() != nil {
			cleanupFailedBackup(outputPath, e.Log)
			return result, args, err
		}

		// Log error pertama kali sebelum retry
		if attempts == 0 {
			// Log error utama
			e.Log.Errorf("Backup gagal: %v", err)
			
			// Error log file sudah dibuat di writer layer untuk semua errors (fatal dan retry-able)
			// Path error log file akan di-log oleh writer layer untuk fatal errors saja
			// Untuk retry-able errors, kita tidak perlu log path di sini karena akan di-retry
			// Detail error sudah ada di error message utama, tidak perlu log lagi
			
			// Cleanup failed backup file sebelum retry
			cleanupFailedBackup(outputPath, e.Log)
		}

		attempts++

		// Attempt retries dengan berbagai strategy
		newResult, newArgs, newErr := e.attemptRetries(outputPath, args, result, err, exec)
		if newErr == nil {
			return newResult, newArgs, nil
		}

		// Jika tidak ada perubahan args (tidak ada strategi retry yang match), stop lebih awal.
		if equalStringSlice(args, newArgs) {
			return newResult, newArgs, newErr
		}

		result, args, err = newResult, newArgs, newErr
	}

	return result, args, err
}

func equalStringSlice(a []string, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := 0; i < len(a); i++ {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func firstLineContaining(haystack string, needleLower string) string {
	if haystack == "" || needleLower == "" {
		return ""
	}
	for _, line := range strings.Split(haystack, "\n") {
		l := strings.ToLower(strings.TrimSpace(line))
		if l == "" {
			continue
		}
		if strings.Contains(l, needleLower) {
			return strings.TrimSpace(line)
		}
	}
	return ""
}

// attemptRetries mencoba retry dengan berbagai strategy.
func (e *Engine) attemptRetries(
	outputPath string,
	args []string,
	result *types_backup.BackupWriteResult,
	lastErr error,
	exec func([]string) (*types_backup.BackupWriteResult, error),
) (*types_backup.BackupWriteResult, []string, error) {
	if result == nil {
		return result, args, fmt.Errorf("attemptRetries: result is nil (likely exec func panic): %w", model.ErrInvalidRetryState)
	}
	if lastErr == nil {
		return result, args, fmt.Errorf("attemptRetries: lastErr is nil (invalid state): %w", model.ErrInvalidRetryState)
	}

	// Note: attemptRetries tidak punya akses ke ctx, tapi caller (executeWithRetry)
	// sudah check ctx.Err() sebelum memanggil attemptRetries, jadi context
	// cancellation sudah ditangani di level yang lebih tinggi.

	// Strategy 1: SSL mismatch - add --skip-ssl
	if IsSSLMismatchRequiredButServerNoSupport(result.StderrOutput) {
		reason := firstLineContaining(result.StderrOutput, "tls/ssl error")
		// Error detail sudah ada di error message utama, tidak perlu log lagi untuk menghindari noise
		logMsg := "Retry dengan --skip-ssl..."
		if strings.TrimSpace(reason) != "" {
			logMsg = fmt.Sprintf("Retry dengan --skip-ssl (deteksi: %s)...", reason)
		}
		newResult, newArgs, matched, retryErr := e.tryRetry(outputPath, args, exec, AddDisableSSLArgs, logMsg)
		if matched {
			if retryErr == nil {
				return newResult, newArgs, nil
			}
			return newResult, newArgs, fmt.Errorf("attemptRetries: retry dengan --skip-ssl gagal: %w", retryErr)
		}
	}

	// Strategy 2: Unsupported option - remove the problematic option
	if newArgs, removed, canRetry := RemoveUnsupportedMysqldumpOption(args, result.StderrOutput); canRetry {
		e.Log.Warnf("Retry tanpa opsi: %s", removed)
		cleanupFailedBackup(outputPath, e.Log)
		if newResult, retryErr := exec(newArgs); retryErr == nil {
			return newResult, newArgs, nil
		} else {
			return newResult, newArgs, fmt.Errorf("attemptRetries: retry tanpa opsi %s gagal: %w", removed, retryErr)
		}
	}

	// No retry strategy matched: jangan timpa error yang lebih informatif.
	// Contoh: strategi SSL sudah dicoba tapi gagal, error-nya harus tetap itu.
	return result, args, lastErr
}

// tryRetry adalah helper untuk execute retry dengan args modifier function.
// Returns: (result, args, success)

func (e *Engine) tryRetry(
	outputPath string,
	args []string,
	exec func([]string) (*types_backup.BackupWriteResult, error),
	argsModifier func([]string) ([]string, bool),
	logMessage string,

) (*types_backup.BackupWriteResult, []string, bool, error) {
	newArgs, modified := argsModifier(args)
	if !modified {
		return nil, args, false, nil
	}

	// Log retry message (cleanup sudah dilakukan di executeWithRetry sebelum attemptRetries)
	e.Log.Warn(logMessage)

	result, err := exec(newArgs)
	return result, newArgs, true, err
}

// handleBackupError handles backup failure - error logging dan cleanup.
func (e *Engine) handleBackupError(
	err error,
	cfg types_backup.BackupExecutionConfig,
	writeResult *types_backup.BackupWriteResult,
) {
	stderrDetail := ""
	if writeResult != nil {
		stderrDetail = writeResult.StderrOutput
	}

	// Log ke error log file
	if e.ErrorLog != nil {
		logMeta := map[string]interface{}{
			"type": cfg.BackupType + "_backup",
			"file": cfg.OutputPath,
		}
		if !cfg.IsMultiDB {
			logMeta["database"] = cfg.DBName
		}
		path := e.ErrorLog.LogWithOutput(logMeta, stderrDetail, err)
		// Jangan spam untuk multi-DB; error summary akan ditangani oleh layer command.
		// Untuk single DB, ini sangat membantu troubleshooting.
		if !cfg.IsMultiDB && strings.TrimSpace(path) != "" {
			e.Log.Warnf("Detail error tersimpan di: %s", path)
		}
	}

	// Cleanup failed backup file
	cleanupFailedBackup(cfg.OutputPath, e.Log)

	// Log error message
	if cfg.IsMultiDB {
		// Untuk multi-DB (backup all/combined), error akan dilaporkan oleh layer command.
		// Detail sudah tersimpan di error log file, jadi jangan spam log console.
	} else {
		e.Log.Error(fmt.Sprintf("gagal backup database %s: %v", cfg.DBName, err))
	}
}
