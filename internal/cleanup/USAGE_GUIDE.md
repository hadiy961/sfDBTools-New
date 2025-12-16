# Cleanup Module - General Purpose Usage Guide

## Overview
Modul cleanup telah di-refactor menjadi general-purpose service yang dapat digunakan oleh modul lain dalam aplikasi sfDBTools. Module ini menyediakan functionality untuk scanning dan menghapus backup files berdasarkan retention policy dan pattern matching.

## Architecture

### Service-Based Design
```go
type Service struct {
    Config         *appconfig.Config      // Application config
    Log            applog.Logger          // Logger instance
    CleanupOptions types.CleanupOptions   // Cleanup options
}
```

### Key Features
- ✅ **Dependency Injection**: No global state, fully testable
- ✅ **Flexible Pattern Matching**: Glob pattern support via doublestar
- ✅ **Multiple Modes**: run, dry-run, pattern-based cleanup
- ✅ **Retention Policy**: Time-based file retention
- ✅ **Comprehensive Logging**: Detailed operation logs
- ✅ **Reusable**: Can be used by any module

## Usage from Other Modules

### 1. Using Cleanup from Backup Module
**Current Implementation** (`internal/backup/executor.go`):
```go
// After backup completion, cleanup old backups if enabled
if s.Config.Backup.Cleanup.Enabled {
    s.Log.Info("Menjalankan cleanup old backups setelah backup...")
    if err := cleanup.CleanupOldBackupsFromBackup(s.Config, s.Log); err != nil {
        s.Log.Warnf("Cleanup old backups gagal: %v", err)
    }
}
```

**Implementation** (`internal/cleanup/cleanup_backup.go`):
```go
func CleanupOldBackupsFromBackup(config *appconfig.Config, logger applog.Logger) error {
    retentionDays := config.Backup.Cleanup.Days
    if retentionDays <= 0 {
        logger.Info("Retention days tidak valid, melewati cleanup")
        return nil
    }

    // Create service instance
    opts := types.CleanupOptions{
        Enabled: true,
        Days:    retentionDays,
        Pattern: "",
    }
    svc := NewCleanupService(config, logger, opts)

    // Scan and delete old files
    baseDir := config.Backup.Output.BaseDirectory
    cutoffTime := time.Now().AddDate(0, 0, -retentionDays)
    
    filesToDelete, err := svc.scanFiles(baseDir, cutoffTime, "")
    if err != nil {
        return fmt.Errorf("gagal scan old backup files: %w", err)
    }

    if len(filesToDelete) == 0 {
        logger.Info("Tidak ada old backup files yang perlu dihapus")
        return nil
    }

    svc.performDeletion(filesToDelete)
    return nil
}
```

### 2. Cleanup Failed Backup Files
**Usage** (`internal/backup/service_helpers.go`):
```go
// Clean up failed backup file
cleanup.CleanupFailedBackup(cfg.OutputPath, s.Log)
```

**Implementation**:
```go
func CleanupFailedBackup(filePath string, logger applog.Logger) {
    if fsops.FileExists(filePath) {
        logger.Infof("Menghapus file backup yang gagal: %s", filePath)
        if err := fsops.RemoveFile(filePath); err != nil {
            logger.Warnf("Gagal menghapus file backup yang gagal: %v", err)
        }
    }
}
```

### 3. Cleanup Partial Backup (Graceful Shutdown)
**Usage** (`internal/backup/service.go`):
```go
func (s *Service) HandleShutdown() {
    if s.backupInProgress && s.currentBackupFile != "" {
        s.Log.Warn("Proses backup dihentikan, melakukan rollback...")
        if err := cleanup.CleanupPartialBackup(fileToRemove, s.Log); err != nil {
            s.Log.Errorf("Gagal menghapus file backup: %v", err)
        }
    }
}
```

**Implementation**:
```go
func CleanupPartialBackup(filePath string, logger applog.Logger) error {
    if err := fsops.RemoveFile(filePath); err != nil {
        return fmt.Errorf("gagal menghapus file backup: %w", err)
    }
    logger.Infof("File backup yang belum selesai berhasil dihapus: %s", filePath)
    return nil
}
```

## Creating Custom Cleanup Service

### Example: Custom Retention Policy
```go
package mymodule

import (
    "sfDBTools/internal/cleanup"
    "sfDBTools/internal/types"
    "time"
)

func CleanupWithCustomRetention(config *appconfig.Config, logger applog.Logger, days int) error {
    // Create custom cleanup options
    opts := types.CleanupOptions{
        Enabled: true,
        Days:    days,
        Pattern: "",
    }
    
    // Create cleanup service
    svc := cleanup.NewCleanupService(config, logger, opts)
    
    // Create entry config
    entryConfig := types.CleanupEntryConfig{
        HeaderTitle: "Custom Cleanup",
        Mode:        "run",
        ShowOptions: false,
        SuccessMsg:  "Cleanup completed",
        LogPrefix:   "custom-cleanup",
        DryRun:      false,
    }
    
    // Execute cleanup
    return svc.ExecuteCleanupCommand(entryConfig)
}
```

### Example: Pattern-Based Cleanup
```go
func CleanupSpecificPattern(config *appconfig.Config, logger applog.Logger, pattern string) error {
    opts := types.CleanupOptions{
        Enabled: true,
        Days:    config.Backup.Cleanup.Days,
        Pattern: pattern, // e.g., "**/*.sql.gz"
    }
    
    svc := cleanup.NewCleanupService(config, logger, opts)
    
    entryConfig := types.CleanupEntryConfig{
        Mode:    "pattern",
        DryRun:  false,
    }
    
    return svc.ExecuteCleanupCommand(entryConfig)
}
```

### Example: Dry-Run Preview
```go
func PreviewCleanup(config *appconfig.Config, logger applog.Logger) error {
    opts := types.CleanupOptions{
        Enabled: true,
        Days:    config.Backup.Cleanup.Days,
    }
    
    svc := cleanup.NewCleanupService(config, logger, opts)
    
    entryConfig := types.CleanupEntryConfig{
        Mode:    "dry-run",
        DryRun:  true,
    }
    
    return svc.ExecuteCleanupCommand(entryConfig)
}
```

## Service Methods

### Public Methods

#### ExecuteCleanupCommand
Entry point untuk cleanup execution dengan configuration.
```go
func (s *Service) ExecuteCleanupCommand(config types.CleanupEntryConfig) error
```

**Parameters:**
- `config`: CleanupEntryConfig dengan mode, flags, dan messages

**Modes:**
- `"run"`: Execute cleanup (delete files)
- `"dry-run"`: Preview cleanup (no deletion)
- `"pattern"`: Cleanup with glob pattern

### Internal Methods (Accessible within package)

#### scanFiles
Scan directory untuk files yang perlu dihapus.
```go
func (s *Service) scanFiles(baseDir string, cutoff time.Time, pattern string) ([]types_backup.BackupFileInfo, error)
```

**Features:**
- Glob pattern matching via doublestar
- Recursive directory traversal
- Filter by modification time
- Sorted by modification time (oldest first)

#### performDeletion
Execute file deletion dengan logging.
```go
func (s *Service) performDeletion(files []types_backup.BackupFileInfo)
```

**Features:**
- Detailed progress logging
- Error handling per file
- Total size calculation
- Summary statistics

#### logDryRunSummary
Display preview of files to be deleted.
```go
func (s *Service) logDryRunSummary(files []types_backup.BackupFileInfo)
```

## Configuration

### CleanupOptions Structure
```go
type CleanupOptions struct {
    Enabled         bool   // Enable/disable cleanup
    Days            int    // Retention days
    CleanupSchedule string // Schedule (future use)
    Pattern         string // Glob pattern
    Background      bool   // Background mode (future use)
}
```

### CleanupEntryConfig Structure
```go
type CleanupEntryConfig struct {
    HeaderTitle string // UI header title
    Mode        string // "run", "dry-run", "pattern"
    ShowOptions bool   // Display options before execution
    SuccessMsg  string // Success message
    LogPrefix   string // Log prefix for tracking
    DryRun      bool   // Dry-run flag
}
```

## Pattern Matching

### Glob Pattern Syntax
Module menggunakan `doublestar/v4` untuk pattern matching:

**Examples:**
```go
"**/*"              // All files recursively
"**/*.sql.gz"       // All .sql.gz files
"**/*_backup_*"     // Files with _backup_ in name
"**/2024/**"        // Files in 2024 directories
"backup_*.sql"      // backup_*.sql in root only
```

### Pattern Usage
```go
opts := types.CleanupOptions{
    Pattern: "**/*_secondary_*.sql.gz", // Only secondary backups
    Days:    30,
}
```

## Best Practices

### 1. Always Check Retention Days
```go
if config.Backup.Cleanup.Days <= 0 {
    logger.Info("Retention days tidak valid, melewati cleanup")
    return nil
}
```

### 2. Use Dry-Run for Testing
```go
// Test cleanup before actual deletion
entryConfig := types.CleanupEntryConfig{
    Mode:   "dry-run",
    DryRun: true,
}
```

### 3. Handle Errors Gracefully
```go
if err := cleanup.CleanupOldBackupsFromBackup(config, logger); err != nil {
    logger.Warnf("Cleanup gagal: %v", err)
    // Don't fail main operation, just log warning
}
```

### 4. Log Operations Thoroughly
```go
logger.Infof("Melakukan cleanup backup files lebih dari %d hari", days)
// Service akan log setiap file deletion
```

### 5. Use Pattern for Selective Cleanup
```go
// Cleanup only specific backup types
opts := types.CleanupOptions{
    Pattern: fmt.Sprintf("**/%s_*.sql.gz", dbName),
    Days:    7,
}
```

## Integration Examples

### Example 1: Scheduled Cleanup
```go
func ScheduledCleanup(config *appconfig.Config, logger applog.Logger) {
    ticker := time.NewTicker(24 * time.Hour)
    defer ticker.Stop()
    
    for range ticker.C {
        logger.Info("Running scheduled cleanup...")
        if err := cleanup.CleanupOldBackupsFromBackup(config, logger); err != nil {
            logger.Errorf("Scheduled cleanup failed: %v", err)
        }
    }
}
```

### Example 2: Conditional Cleanup
```go
func CleanupIfNeeded(config *appconfig.Config, logger applog.Logger) error {
    // Check disk space first
    if diskSpaceBelow(config.Backup.Output.BaseDirectory, 10) {
        logger.Warn("Disk space low, performing aggressive cleanup...")
        
        opts := types.CleanupOptions{
            Days: 7, // More aggressive retention
        }
        svc := cleanup.NewCleanupService(config, logger, opts)
        
        entryConfig := types.CleanupEntryConfig{Mode: "run"}
        return svc.ExecuteCleanupCommand(entryConfig)
    }
    
    return nil
}
```

### Example 3: Module-Specific Cleanup
```go
package restore

func CleanupRestoreFiles(config *appconfig.Config, logger applog.Logger) error {
    opts := types.CleanupOptions{
        Pattern: "**/restore_temp_*",
        Days:    1, // Very short retention for temp files
    }
    
    svc := cleanup.NewCleanupService(config, logger, opts)
    entryConfig := types.CleanupEntryConfig{
        Mode:       "pattern",
        LogPrefix:  "restore-cleanup",
    }
    
    return svc.ExecuteCleanupCommand(entryConfig)
}
```

## Testing

### Unit Test Example
```go
func TestCleanupService(t *testing.T) {
    // Create test config
    config := &appconfig.Config{
        Backup: appconfig.BackupConfig{
            Output: appconfig.OutputConfig{
                BaseDirectory: "/tmp/test-backups",
            },
            Cleanup: appconfig.CleanupConfig{
                Days: 30,
            },
        },
    }
    
    // Create test logger
    logger := applog.NewLogger(...)
    
    // Create cleanup options
    opts := types.CleanupOptions{
        Days: 30,
    }
    
    // Create service
    svc := cleanup.NewCleanupService(config, logger, opts)
    
    // Test dry-run
    entryConfig := types.CleanupEntryConfig{
        Mode:   "dry-run",
        DryRun: true,
    }
    
    err := svc.ExecuteCleanupCommand(entryConfig)
    assert.NoError(t, err)
}
```

## Troubleshooting

### Common Issues

1. **No files deleted despite old files existing**
   - Check retention days configuration
   - Verify file modification times
   - Ensure pattern matches files

2. **Permission denied errors**
   - Check file permissions
   - Ensure process has write access to directory

3. **Pattern not matching files**
   - Use dry-run to test pattern
   - Check doublestar glob syntax
   - Verify pattern escaping

## Future Enhancements

Potential improvements for the cleanup module:

1. **Size-based cleanup**: Delete oldest files until size limit
2. **Count-based cleanup**: Keep only N most recent backups
3. **Cleanup metrics**: Track cleanup statistics
4. **Cleanup scheduling**: Built-in cron-like scheduling
5. **Multi-directory cleanup**: Cleanup across multiple directories
6. **Selective retention**: Different policies for different file types

## Conclusion

Cleanup module sekarang merupakan general-purpose service yang:
- ✅ Dapat digunakan oleh berbagai module
- ✅ Memiliki API yang clean dan testable
- ✅ Support multiple cleanup strategies
- ✅ Fully configurable dan extensible
- ✅ Production-ready dengan comprehensive logging

For questions or improvements, refer to the development team.
