# Cleanup Module Refactoring Notes

## Tanggal: 2025-12-16

### Overview
Modul cleanup telah di-refactor untuk mengikuti pattern yang sama dengan modul backup, meningkatkan consistency, testability, dan maintainability.

## Perubahan Utama

### 1. Service Pattern dengan Proper Dependency Injection
**Sebelum:**
```go
var Logger applog.Logger
var cfg *appconfig.Config

func NewCleanupService(logger applog.Logger, config *appconfig.Config) *Service {
    return &Service{Logger: logger, Config: config}
}
```

**Sesudah:**
```go
type Service struct {
    Config         *appconfig.Config
    Log            applog.Logger
    CleanupOptions types.CleanupOptions
}

func NewCleanupService(config *appconfig.Config, logger applog.Logger, opts types.CleanupOptions) *Service {
    return &Service{Config: config, Log: logger, CleanupOptions: opts}
}
```

**Benefits:**
- Eliminasi package-level variables untuk better testability
- Proper dependency injection pattern
- Consistent dengan backup module

### 2. Unified Command Execution Pattern
**File Baru:** `internal/cleanup/command.go`

**API:**
```go
func ExecuteCleanup(cmd *cobra.Command, deps *types.Dependencies, mode string) error
```

**Pattern:**
- Parse options dari flags
- Create service dengan proper dependencies
- Execute cleanup dengan CleanupEntryConfig
- Display results

**Benefits:**
- Single entry point untuk semua cleanup commands
- Consistent dengan `backup.ExecuteBackup()` pattern
- Reduced duplication di command layer

### 3. Implementasi ExecuteCleanupCommand
**File:** `internal/cleanup/cleanup_entry.go`

**Features:**
- Menggunakan `CleanupEntryConfig` untuk configuration
- Support untuk mode: "run", "dry-run", "pattern"
- Display options functionality
- Proper error handling

### 4. Refactor Core Functions ke Service Methods
**Perubahan di `cleanup_core.go`:**
- `cleanupCore()` → `(s *Service) cleanupCore()`
- `scanFiles()` → `(s *Service) scanFiles()`
- `performDeletion()` → `(s *Service) performDeletion()`
- `logDryRunSummary()` → `(s *Service) logDryRunSummary()`

**Removed:**
- Package-level variables (Logger, cfg)
- `SetConfig()` function
- `scanFilesWithLogger()` wrapper
- `performDeletionWithLogger()` wrapper

**Benefits:**
- No global state
- Better encapsulation
- Easier to test with different configurations

### 5. Simplified Command Files
**Files Updated:**
- `cmd/cmd_cleanup/cmd_cleanup_run.go`
- `cmd/cmd_cleanup/cmd_cleanup_dryrun.go`
- `cmd/cmd_cleanup/cmd_cleanup_pattern.go`

**Sebelum:**
```go
Run: func(cmd *cobra.Command, args []string) {
    cleanup.Logger = types.Deps.Logger
    cleanup.SetConfig(types.Deps.Config)
    if err := cleanup.CleanupOldBackups(); err != nil {
        types.Deps.Logger.Errorf("Cleanup gagal: %v", err)
    }
}
```

**Sesudah:**
```go
Run: func(cmd *cobra.Command, args []string) {
    if err := cleanup.ExecuteCleanup(cmd, types.Deps, "run"); err != nil {
        types.Deps.Logger.Error("cleanup gagal: " + err.Error())
    }
}
```

**Benefits:**
- Eliminasi redundant injection code
- Consistent error handling
- Cleaner command implementation

### 6. General Purpose Cleanup Functions
**File:** `internal/cleanup/cleanup_backup.go`

**Updated:**
```go
func CleanupOldBackupsFromBackup(config *appconfig.Config, logger applog.Logger) error {
    // Membuat temporary service untuk cleanup
    opts := types.CleanupOptions{
        Enabled: true,
        Days:    retentionDays,
        Pattern: "",
    }
    svc := NewCleanupService(config, logger, opts)
    
    // Use service methods
    filesToDelete, err := svc.scanFiles(baseDir, cutoffTime, "")
    svc.performDeletion(filesToDelete)
}
```

**Benefits:**
- Backup module dapat menggunakan cleanup service secara general
- Tidak ada code duplication
- Reusable untuk module lain

### 7. Enhanced Parsing Support
**File:** `pkg/parsing/parsing_cleanup.go`

**Added support for:**
- `days` flag
- `pattern` flag
- `background` flag

## Architecture Alignment

### Similarities with Backup Module
✅ Service struct dengan proper DI
✅ Unified command execution (ExecuteCleanup vs ExecuteBackup)
✅ Entry point dengan config pattern (ExecuteCleanupCommand)
✅ command.go untuk command layer logic
✅ Simplified command files

### Differences (By Design)
❌ No BaseService embedding - cleanup tidak perlu mutex/cancel func
❌ No subfolder structure - logic terlalu simple
❌ No context parameter - operations terlalu fast
❌ No factory pattern - hanya 3 simple modes

## Testing Considerations

### Manual Testing
```bash
# Test cleanup help
./bin/sfdbtools cleanup --help

# Test cleanup run
./bin/sfdbtools cleanup run --ticket TEST123

# Test dry-run
./bin/sfdbtools cleanup dry-run

# Test pattern
./bin/sfdbtools cleanup pattern --pattern "**/*.sql.gz"
```

### Build Verification
```bash
go build -o bin/sfdbtools main.go
# Success - no compile errors
```

## Migration Impact

### Breaking Changes
None - internal package only

### API Changes
- Removed public package-level variables
- Added `ExecuteCleanup()` public function
- Service constructor signature changed (added opts parameter)

### Backward Compatibility
- `CleanupOldBackupsFromBackup()` tetap compatible
- `CleanupFailedBackup()` tetap compatible
- `CleanupPartialBackup()` tetap compatible

## Future Enhancements

### Possible Improvements
1. Add context support untuk cancellation (optional)
2. Extract display logic ke subfolder (optional)
3. Add more cleanup modes (by size, by count, etc.)
4. Add cleanup metrics/statistics
5. Add cleanup scheduling support

### Not Needed (Keep Simple)
- BaseService embedding
- Complex factory patterns
- Multiple executor implementations

## Conclusion

Refactoring berhasil dilakukan dengan benefits:
- ✅ Consistency dengan backup module pattern
- ✅ Better testability (no package-level vars)
- ✅ Reduced code duplication
- ✅ Cleaner command layer
- ✅ General purpose cleanup service
- ✅ Maintainable architecture

Cleanup module sekarang mengikuti best practices project dan dapat digunakan secara general oleh module lain.
