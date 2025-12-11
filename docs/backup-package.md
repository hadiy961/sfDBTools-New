# Dokumentasi Package Backup

## Deskripsi
Package `internal/backup` adalah modul inti untuk operasi backup database MariaDB/MySQL di sfDBTools. Package ini menyediakan berbagai mode backup (single, separated, combined, primary, secondary) dengan dukungan kompresi, enkripsi, dan graceful shutdown.

**Path**: `internal/backup`  
**Author**: Hadiyatna Muflihun  
**Last Modified**: 2025-12-05

---

## Daftar Isi
- [Arsitektur](#arsitektur)
- [Struktur File](#struktur-file)
- [Komponen Utama](#komponen-utama)
- [Mode Backup](#mode-backup)
- [Pipeline Eksekusi](#pipeline-eksekusi)
- [Error Handling](#error-handling)
- [Metadata Management](#metadata-management)
- [Helper Functions](#helper-functions)
- [Interface dan Patterns](#interface-dan-patterns)
- [Contoh Penggunaan](#contoh-penggunaan)

---

## Arsitektur

### Diagram Alur
```
cmd layer (Cobra)
    ↓
command.go (ExecuteBackup)
    ↓
service.go (ExecuteBackupCommand)
    ↓
executor.go (ExecuteBackup)
    ↓
modes/* (CombinedExecutor/SeparatedExecutor/SingleExecutor)
    ↓
helpers.go (executeBackupLoop/executeAndBuildBackup)
    ↓
writer.go (executeMysqldumpWithPipe)
    ↓
metadata/* (GenerateBackupMetadata/SaveGTID/ExportUserGrants)
```

### Dependency Injection Pattern
Package ini menggunakan **Service Pattern** dengan interface-based design:
- `Service` struct mengimplementasikan `modes.BackupService` interface
- Mode executors (`CombinedExecutor`, `SeparatedExecutor`, `SingleExecutor`) bergantung pada interface, bukan konkrit implementation
- Memungkinkan testing dan extensibility yang lebih baik

---

## Struktur File

```
internal/backup/
├── command.go              # Entry point dari cmd layer (Cobra)
├── service.go              # Service struct dan interface implementations
├── executor.go             # Main execution logic
├── setup.go                # Preparation dan setup functions
├── helpers.go              # Helper functions (loops, path generation, filtering)
├── writer.go               # Writer pipeline (compression, encryption)
├── errors.go               # Error handling dan logging
├── display/
│   ├── options_display.go  # Display backup options dan konfirmasi
│   └── result_display.go   # Display backup results
├── helper/
│   ├── compression_helper.go  # Compression utilities
│   └── file_helper.go         # File path generation utilities
├── metadata/
│   ├── backup_metadata.go     # Metadata generation dan persistence
│   ├── gtid.go                # GTID capture dan save
│   └── user.go                # User grants export
├── modes/
│   ├── interface.go        # Interface definitions untuk mode executors
│   ├── combined.go         # Combined mode executor
│   ├── separated.go        # Separated mode executor
│   └── single.go           # Single mode executor
└── strategy/
    └── strategy.go         # Strategy pattern untuk backup modes
```

---

## Komponen Utama

### 1. Service (`service.go`)

**Struct Definition**:
```go
type Service struct {
    servicehelper.BaseService
    
    Config          *appconfig.Config
    Log             applog.Logger
    ErrorLog        *errorlog.ErrorLogger
    DBInfo          *types.DBInfo
    Profile         *types.ProfileInfo
    BackupDBOptions *types_backup.BackupDBOptions
    BackupEntry     *types_backup.BackupEntryConfig
    Client          *database.Client
    
    // Backup-specific state
    currentBackupFile string
    backupInProgress  bool
}
```

**Key Methods**:
- `NewBackupService(logs, cfg, backup)` - Constructor dengan type switching untuk options
- `ExecuteBackup(ctx, sourceClient, dbFiltered, backupMode)` - Main entry point untuk backup execution
- `SetCurrentBackupFile(filePath)` / `ClearCurrentBackupFile()` - Track backup in progress untuk graceful shutdown
- `HandleShutdown()` - Cleanup partial files saat interrupt

**Interface Implementations**:
Implements `modes.BackupService` interface untuk digunakan oleh mode executors:
- `LogInfo/LogDebug/LogWarn/LogError` - Logging methods
- `ExecuteAndBuildBackup` - Execute single database backup
- `ExecuteBackupLoop` - Execute multiple databases backup
- `GetBackupOptions` - Access configuration
- `GenerateFullBackupPath` - Path generation
- `CaptureAndSaveGTID` - GTID handling
- `ExportUserGrantsIfNeeded` - User grants export

---

### 2. Command Layer (`command.go`)

**Public API**:
```go
func ExecuteBackup(cmd *cobra.Command, deps *types.Dependencies, mode string) error
```

**Mode Configurations**:
```go
modeConfigs := map[string]types_backup.ExecutionConfig{
    "single":    { Mode: "single", HeaderTitle: "Database Backup - Single", ... },
    "separated": { Mode: "separated", HeaderTitle: "Database Backup - Separated", ... },
    "combined":  { Mode: "combined", HeaderTitle: "Database Backup - Combined", ... },
    "primary":   { Mode: "primary", HeaderTitle: "Database Backup - Primary", ... },
    "secondary": { Mode: "secondary", HeaderTitle: "Database Backup - Secondary", ... },
}
```

**Execution Flow**:
1. Parse mode dari input string
2. Load configuration dari `modeConfigs`
3. Parse backup options via `parsing.ParsingBackupOptions()`
4. Initialize backup service
5. Setup context dengan cancellation untuk graceful shutdown
6. Setup signal handler untuk CTRL+C / SIGTERM
7. Execute backup command via `svc.ExecuteBackupCommand()`

**Graceful Shutdown Handling**:
```go
sigChan := make(chan os.Signal, 1)
signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

go func() {
    sig := <-sigChan
    logger.Warnf("Menerima signal %v, menghentikan backup...", sig)
    svc.HandleShutdown()  // Cleanup partial files
    cancel()              // Cancel context
}()
```

---

### 3. Executor (`executor.go`)

**Main Entry Point**:
```go
func (s *Service) ExecuteBackup(
    ctx context.Context, 
    sourceClient *database.Client, 
    dbFiltered []string, 
    backupMode string
) (*types_backup.BackupResult, error)
```

**Execution Steps**:
1. Setup backup execution configuration
2. Execute backup sesuai mode (combined/separated/single)
3. Handle cleanup old backups (jika enabled)
4. Handle backup errors
5. Return backup result

**Mode Selection Logic**:
```go
switch backupMode {
case "separate", "separated":
    return s.ExecuteBackupSeparated(ctx, dbFiltered)
case "single", "primary", "secondary":
    return s.ExecuteBackupSingle(ctx, dbFiltered)
default:
    return s.ExecuteBackupCombined(ctx, dbFiltered)
}
```

---

### 4. Writer Pipeline (`writer.go`)

**Main Function**:
```go
func (s *Service) executeMysqldumpWithPipe(
    ctx context.Context, 
    mysqldumpArgs []string, 
    outputPath string, 
    compressionRequired bool, 
    compressionType string
) (*types_backup.BackupWriteResult, error)
```

**Writer Pipeline Layers**:
```
mysqldump stdout 
    ↓
Compression Writer (optional: gzip/zstd/xz/pgzip)
    ↓
Encryption Writer (optional: AES-256-GCM)
    ↓
Buffered Writer (256KB buffer)
    ↓
File Writer
```

**Pipeline Construction**:
```go
func (s *Service) createWriterPipeline(
    baseWriter io.Writer, 
    compressionRequired bool, 
    compressionType string, 
    encryptionKey string
) (io.Writer, []io.Closer, error)
```

**Key Features**:
- **Buffered I/O**: 256KB buffer untuk performance
- **Streaming**: Process data in chunks, tidak load semua ke memory
- **Encryption First**: Layer enkripsi paling dekat dengan file (encrypted data yang di-compress)
- **Error Recovery**: Cleanup file jika mysqldump gagal
- **Stderr Capture**: Capture output error dari mysqldump untuk logging

---

### 5. Setup & Preparation (`setup.go`)

**Profile Loading**:
```go
func (s *Service) CheckAndSelectConfigFile() error
```
- Load profile menggunakan `profilehelper.LoadSourceProfile()`
- Interactive selector jika profile tidak dispesifikasi (untuk single/primary/secondary mode)
- Decrypt profile dengan encryption key

**Session Preparation**:
```go
func (s *Service) PrepareBackupSession(
    ctx context.Context, 
    headerTitle string, 
    showOptions bool
) (client *database.Client, dbFiltered []string, err error)
```

**Steps**:
1. Check dan select config file (profile)
2. Connect ke database via `profilehelper.ConnectWithProfile()`
3. Get server hostname dari MySQL server
4. Filter databases sesuai rules (include/exclude lists)
5. Generate backup paths (output directory & filename)
6. Display options dan minta konfirmasi user

**Database Filtering**:
```go
func (s *Service) GetFilteredDatabases(
    ctx context.Context, 
    client *database.Client
) ([]string, *types.DatabaseFilterStats, error)
```
- Menggunakan `database.FilterFromBackupOptions()` dari `pkg/database`
- Apply exclude system databases
- Apply exclude empty databases
- Apply include/exclude lists
- Apply include/exclude files (whitelist/blacklist)
- Return filtered list + statistics

---

### 6. Error Handling (`errors.go`)

**Error Handler Struct**:
```go
type BackupErrorHandler struct {
    Logger   applog.Logger
    ErrorLog *errorlog.ErrorLogger
    ShowUI   bool
}
```

**Error Types**:

1. **Database Backup Error** (per-database failure):
```go
func (h *BackupErrorHandler) HandleDatabaseBackupError(
    filePath string,
    dbName string,
    err error,
    stderrOutput string,
    logMetadata map[string]interface{},
) string
```
- Cleanup failed backup file
- Log error dengan stderr output
- Display error di UI
- Return formatted error message

2. **Combined Backup Error** (all databases failure):
```go
func (h *BackupErrorHandler) HandleCombinedBackupError(
    filePath string,
    err error,
    stderrOutput string,
    logMetadata map[string]interface{},
) error
```
- Cleanup failed backup file
- Log error dengan detail mysqldump output
- Display error header dan detail di UI
- Return wrapped error

**Error Logging**:
- Log disimpan di directory yang dikonfigurasi (`/var/log/sfDBTools` default)
- Format: JSON dengan metadata (database, type, file path, timestamp)
- Include stderr output dari mysqldump

---

## Mode Backup

### 1. Combined Mode (`modes/combined.go`)

**Deskripsi**: Backup semua database dalam satu file SQL.

**Executor**:
```go
type CombinedExecutor struct {
    service BackupService
}
```

**Execution Flow**:
1. Generate single output path untuk semua databases
2. Execute backup dengan `IsMultiDB=true`
3. Capture GTID (jika enabled)
4. Export user grants
5. Format display name dengan list databases

**Output**:
- Single file: `combined_hostname_20250101_120000.sql.gz`
- GTID file (optional): `combined_hostname_20250101_120000_gtid.info`
- User grants file (optional): `combined_hostname_20250101_120000_users.cnf`
- Metadata file: `combined_hostname_20250101_120000.meta.json`

---

### 2. Separated Mode (`modes/separated.go`)

**Deskripsi**: Backup setiap database dalam file terpisah.

**Executor**:
```go
type SeparatedExecutor struct {
    service BackupService
}
```

**Execution Flow**:
1. Loop semua filtered databases
2. Generate output path per database
3. Execute backup untuk setiap database
4. Collect success/failed statistics
5. Export user grants dari first successful backup

**Output** (per database):
- `database1_separated_hostname_20250101_120000.sql.gz`
- `database2_separated_hostname_20250101_120000.sql.gz`
- `database3_separated_hostname_20250101_120000.sql.gz`
- User grants file (single): `database1_separated_hostname_20250101_120000_users.cnf`

---

### 3. Single Mode (`modes/single.go`)

**Deskripsi**: Backup satu database utama dengan optional companion databases (_dmart, _temp, _archive).

**Executor**:
```go
type SingleExecutor struct {
    service BackupService
}
```

**Companion Databases**:
- `database_dmart` - Data mart (jika `--include-dmart`)
- `database_temp` - Temporary tables (jika `--include-temp`)
- `database_archive` - Archive data (jika `--include-archive`)

**Execution Flow**:
1. Select database utama (interactive atau via flag)
2. Check companion databases availability
3. Loop backup untuk primary + companions
4. Use custom filename untuk primary database (jika dispesifikasi)
5. Export user grants dari first successful backup

**Output**:
- `database_single_hostname_20250101_120000.sql.gz` (primary)
- `database_dmart_single_hostname_20250101_120000.sql.gz` (companion)
- `database_temp_single_hostname_20250101_120000.sql.gz` (companion)

---

### 4. Primary Mode

**Deskripsi**: Sama seperti Single Mode, tapi otomatis filter hanya database primary (exclude `_secondary`, `_dmart`, `_temp`, `_archive`).

**Filter Rules**:
```go
if strings.Contains(dbLower, "_secondary") || 
   strings.HasSuffix(dbLower, "_dmart") ||
   strings.HasSuffix(dbLower, "_temp") || 
   strings.HasSuffix(dbLower, "_archive") {
    continue // Skip
}
```

---

### 5. Secondary Mode

**Deskripsi**: Sama seperti Single Mode, tapi otomatis filter hanya database secondary (include `_secondary`, exclude lainnya).

**Filter Rules**:
```go
if !strings.Contains(dbLower, "_secondary") || 
   strings.HasSuffix(dbLower, "_dmart") ||
   strings.HasSuffix(dbLower, "_temp") || 
   strings.HasSuffix(dbLower, "_archive") {
    continue // Skip
}
```

---

## Pipeline Eksekusi

### Backup Loop Flow

```go
func (s *Service) executeBackupLoop(
    ctx context.Context,
    databases []string,
    config types_backup.BackupLoopConfig,
    outputPathFunc func(dbName string) (string, error),
) types_backup.BackupLoopResult
```

**Steps per Database**:
1. Check context cancellation (graceful shutdown)
2. Log progress `[n/total]`
3. Generate output path via `outputPathFunc`
4. Execute backup via `executeAndBuildBackup()`
5. Handle error (add to failed list)
6. Collect backup info (add to success list)

**Return Result**:
```go
type BackupLoopResult struct {
    BackupInfos []types.DatabaseBackupInfo
    FailedDBs   []types_backup.FailedDatabaseInfo
    Errors      []string
    Success     int
    Failed      int
}
```

---

### Single Backup Execution

```go
func (s *Service) executeAndBuildBackup(
    ctx context.Context, 
    cfg types_backup.BackupExecutionConfig
) (types.DatabaseBackupInfo, error)
```

**Steps**:
1. Start timer untuk duration tracking
2. Set current backup file (untuk graceful shutdown)
3. Build mysqldump arguments
4. Execute mysqldump dengan writer pipeline
5. Handle mysqldump errors (fatal vs warning)
6. Generate metadata
7. Save metadata to file
8. Build backup info dengan throughput calculation
9. Clear current backup file
10. Return backup info

**Configuration**:
```go
type BackupExecutionConfig struct {
    DBName       string   // Single database name
    DBList       []string // Multiple databases (combined mode)
    OutputPath   string   // Full output path
    BackupType   string   // "single" | "separated" | "combined"
    TotalDBFound int      // Total databases di server
    IsMultiDB    bool     // True untuk combined mode
}
```

---

## Metadata Management

### 1. Backup Metadata (`metadata/backup_metadata.go`)

**Struct Definition**:
```go
type BackupMetadata struct {
    BackupFile      string
    BackupType      string
    DatabaseNames   []string
    Hostname        string
    BackupStartTime time.Time
    BackupEndTime   time.Time
    BackupDuration  string
    FileSize        int64
    FileSizeHuman   string
    Compressed      bool
    CompressionType string
    Encrypted       bool
    BackupStatus    string
    Warnings        []string
    GeneratedBy     string
    GeneratedAt     time.Time
}
```

**Generation**:
```go
func GenerateBackupMetadata(cfg types_backup.MetadataConfig) *types_backup.BackupMetadata
```

**Persistence**:
```go
func SaveBackupMetadata(meta *types_backup.BackupMetadata, logger applog.Logger) (string, error)
```
- Atomic write dengan temporary file + rename pattern
- Format: JSON dengan indentation
- Output: `backup_file.meta.json`

---

### 2. GTID Information (`metadata/gtid.go`)

**Purpose**: Capture binary log position untuk point-in-time recovery.

**Capture Function**:
```go
func CaptureAndSaveGTID(
    ctx context.Context, 
    client *database.Client, 
    logger applog.Logger, 
    backupFilePath string, 
    captureGTID bool
) error
```

**GTID Info Structure**:
```go
type GTIDInfo struct {
    MasterLogFile string
    MasterLogPos  int64
    GTIDBinlog    string
}
```

**File Format**:
```
# GTID Information
# Generated at: 2025-01-01 12:00:00

MASTER_LOG_FILE = mysql-bin.000123
MASTER_LOG_POS = 154567890
gtid_binlog = 123e4567-e89b-12d3-a456-426614174000:1-100
```

**Output**: `backup_file_gtid.info`

---

### 3. User Grants Export (`metadata/user.go`)

**Purpose**: Export semua user privileges untuk restore users.

**Export Function**:
```go
func ExportAndSaveUserGrants(
    ctx context.Context, 
    client *database.Client, 
    logger applog.Logger, 
    backupFilePath string, 
    excludeUser bool
) error
```

**Query Execution**:
```go
userGrantsSQL, err := client.ExportAllUserGrants(ctx)
```
- Get all users (exclude system users: 'root'@'localhost', 'mysql.sys', dll)
- Generate CREATE USER statements
- Generate GRANT statements

**Output**: `backup_file_users.cnf`

**File Format**:
```sql
-- User Grants Export
-- Generated at: 2025-01-01 12:00:00

CREATE USER 'user1'@'%' IDENTIFIED BY PASSWORD '*hash';
GRANT SELECT, INSERT ON db1.* TO 'user1'@'%';

CREATE USER 'user2'@'localhost' IDENTIFIED BY PASSWORD '*hash';
GRANT ALL PRIVILEGES ON db2.* TO 'user2'@'localhost';
```

---

## Helper Functions

### 1. Compression Helper (`helper/compression_helper.go`)

```go
// Convert string compression type ke compress.CompressionType
func ConvertCompressionType(enabled bool, compressionType string) compress.CompressionType

// Create CompressionSettings dari konfigurasi
func NewCompressionSettings(enabled bool, compressionType string, level int) types_backup.CompressionSettings
```

**Supported Compression Types**:
- `gzip` - Standard gzip compression (balanced)
- `zstd` - Zstandard (faster, better ratio)
- `xz` - LZMA2 (slower, best ratio)
- `pgzip` - Parallel gzip (faster dengan multi-core)

---

### 2. File Helper (`helper/file_helper.go`)

```go
// Generate path untuk user grants file
func GenerateUserFilePath(backupFilePath string) string
// Input:  /backup/db_20250101.sql.gz
// Output: /backup/db_20250101_users.cnf

// Generate path untuk GTID info file
func GenerateGTIDFilePath(backupFilePath string) string
// Input:  /backup/db_20250101.sql.gz
// Output: /backup/db_20250101_gtid.info

// Generate path untuk metadata file
func GenerateBackupMetadataFilePath(backupFilePath string) string
// Input:  /backup/db_20250101.sql.gz
// Output: /backup/db_20250101.metadata.json
```

**Logic**:
- Extract directory dan basename
- Remove extensions (handle double extension seperti `.sql.gz`)
- Append suffix sesuai tipe file
- Return full path

---

### 3. Database Filtering (`helpers.go`)

```go
// Filter databases berdasarkan mode (primary/secondary/single)
func (s *Service) filterCandidatesByMode(dbFiltered []string, mode string) []string
```

**Mode-Specific Filters**:
- **Primary**: Exclude `_secondary`, `_dmart`, `_temp`, `_archive`
- **Secondary**: Include only `_secondary`, exclude others
- **Single**: Exclude `_dmart`, `_temp`, `_archive`

```go
// Add companion databases (_dmart, _temp, _archive)
func (s *Service) addCompanionDatabases(
    selectedDB string, 
    companionDbs *[]string,
    companionStatus map[string]bool, 
    allDatabases []string
)
```

---

### 4. Path Generation (`helpers.go`)

```go
// Generate full path untuk backup file
func (s *Service) generateFullBackupPath(dbName string, mode string) (string, error)
```

**Filename Pattern**:
```
{dbName}_{mode}_{hostname}_{timestamp}.sql{.compression_ext}{.enc}
```

**Examples**:
- `mydb_single_server1_20250101_120000.sql.gz`
- `mydb_separated_server1_20250101_120000.sql.zst.enc`
- `combined_server1_20250101_120000.sql.xz`

---

## Interface dan Patterns

### 1. BackupService Interface (`modes/interface.go`)

**Purpose**: Decouple mode executors dari Service implementation.

```go
type BackupService interface {
    // Logging
    LogInfo(msg string)
    LogDebug(msg string)
    LogWarn(msg string)
    LogError(msg string)

    // Backup execution
    ExecuteAndBuildBackup(ctx context.Context, cfg types_backup.BackupExecutionConfig) (types.DatabaseBackupInfo, error)
    ExecuteBackupLoop(ctx context.Context, databases []string, config types_backup.BackupLoopConfig, outputPathFunc func(dbName string) (string, error)) types_backup.BackupLoopResult

    // Helpers
    GetBackupOptions() *types_backup.BackupDBOptions
    GenerateFullBackupPath(dbName string, mode string) (string, error)
    GetTotalDatabaseCount(ctx context.Context, dbFiltered []string) int
    CaptureAndSaveGTID(ctx context.Context, backupFilePath string) error
    ExportUserGrantsIfNeeded(ctx context.Context, referenceBackupFile string)
    ToBackupResult(loopResult types_backup.BackupLoopResult) types_backup.BackupResult
}
```

**Benefits**:
- Testing: Mock interface untuk unit testing
- Extensibility: Mudah tambah method baru
- Separation of Concerns: Mode executors tidak perlu tahu internal Service

---

### 2. ModeExecutor Interface (`modes/interface.go`)

**Purpose**: Unified interface untuk semua backup modes.

```go
type ModeExecutor interface {
    Execute(ctx context.Context, databases []string) types_backup.BackupResult
}
```

**Implementations**:
- `CombinedExecutor`
- `SeparatedExecutor`
- `SingleExecutor`

**Usage Pattern**:
```go
var executor modes.ModeExecutor

switch mode {
case "combined":
    executor = modes.NewCombinedExecutor(service)
case "separated":
    executor = modes.NewSeparatedExecutor(service)
default:
    executor = modes.NewSingleExecutor(service)
}

result := executor.Execute(ctx, databases)
```

---

### 3. Strategy Pattern (`strategy/strategy.go`)

**Purpose**: Alternative pattern untuk mode selection (currently not actively used, reserved for future).

```go
type BackupStrategy interface {
    Execute(ctx context.Context, dbList []string) types_backup.BackupResult
}

type BackupMode string

const (
    BackupModeSingle    BackupMode = "single"
    BackupModeSeparated BackupMode = "separated"
    BackupModeCombined  BackupMode = "combined"
    BackupModePrimary   BackupMode = "primary"
    BackupModeSecondary BackupMode = "secondary"
)
```

---

### 4. Builder Pattern (Metadata)

**DatabaseBackupInfoBuilder**:
```go
type DatabaseBackupInfoBuilder struct {
    DatabaseName string
    OutputFile   string
    FileSize     int64
    Duration     time.Duration
    Status       string
    Warnings     string
    StartTime    time.Time
    EndTime      time.Time
    ManifestFile string
}

func (b *DatabaseBackupInfoBuilder) Build() types.DatabaseBackupInfo
```

**Benefits**:
- Clean construction process
- Automatic calculations (throughput, backup ID)
- Consistent formatting (file size, duration)

---

## Display Components

### 1. Options Display (`display/options_display.go`)

**Purpose**: Display backup configuration dan minta konfirmasi user.

```go
type OptionsDisplayer struct {
    options *types_backup.BackupDBOptions
}

func (d *OptionsDisplayer) Display() (bool, error)
```

**Display Sections**:
1. General Information (mode, output dir, filename)
2. Mode-Specific (database selection, companion status)
3. Profile Information (hostname, host, user)
4. Filter Options (exclude/include lists)
5. Compression Settings (type, level)
6. Encryption Settings (status)
7. Cleanup Settings (enabled, days, pattern)

**Table Format** (via `pkg/ui`):
```
┌─────────────────────┬─────────────────────────────────┐
│ Parameter           │ Value                           │
├─────────────────────┼─────────────────────────────────┤
│ Mode Backup         │ single                          │
│ Output Directory    │ /backup/20250101                │
│ Filename Pattern    │ {db}_{mode}_{host}_{time}.sql   │
│ ...                 │ ...                             │
└─────────────────────┴─────────────────────────────────┘
```

---

### 2. Result Display (`display/result_display.go`)

**Purpose**: Display hasil backup (success/failed databases).

```go
type ResultDisplayer struct {
    result             *types_backup.BackupResult
    compressionEnabled bool
    compressionType    string
    encryptionEnabled  bool
}

func (d *ResultDisplayer) Display()
```

**Display Sections**:
1. **Summary Statistics**: Total found, success, failed, duration
2. **Success Details**: Per-database details (file, size, duration, throughput)
3. **Failures**: List failed databases dengan error messages

**Example Output**:
```
=== Hasil Backup Database ===

┌────────────────────────┬─────────┐
│ Kategori               │ Jumlah  │
├────────────────────────┼─────────┤
│ Total Database         │ 5       │
│ Total Berhasil         │ 4       │
│ Total Gagal            │ 1       │
│ Total Waktu Proses     │ 2m 30s  │
└────────────────────────┴─────────┘

=== Detail Backup yang Berhasil ===

┌─────────────────┬──────────────────────────────────┐
│ Parameter       │ Nilai                            │
├─────────────────┼──────────────────────────────────┤
│ Database        │ mydb                             │
│ Status          │ success                          │
│ File Output     │ /backup/mydb_single_...sql.gz    │
│ Ukuran File     │ 1.5 GB                           │
│ Durasi Backup   │ 45s                              │
│ Throughput      │ 34.12 MB/s                       │
│ Kompresi        │ Enabled (gzip)                   │
│ Enkripsi        │ Enabled                          │
└─────────────────┴──────────────────────────────────┘

=== Daftar Database Gagal Dibackup ===
1. failed_db
   Error: connection timeout
```

---

## Contoh Penggunaan

### 1. Combined Mode (dari cmd layer)

```go
// cmd/cmd_backup/cmd_backup_combined.go
func init() {
    combinedCmd.Flags().StringP("profile", "p", "", "Path ke database profile")
    combinedCmd.Flags().Bool("compression", true, "Enable compression")
    combinedCmd.Flags().String("compression-type", "gzip", "Compression type")
    combinedCmd.Flags().Bool("encryption", false, "Enable encryption")
}

var combinedCmd = &cobra.Command{
    Use:   "combined",
    Short: "Backup semua database dalam satu file",
    RunE: func(cmd *cobra.Command, args []string) error {
        return backup.ExecuteBackup(cmd, types.Deps, "combined")
    },
}
```

**CLI Usage**:
```bash
sfdbtools backup combined \
    --profile /path/to/profile.cnf.enc \
    --compression \
    --compression-type zstd \
    --encryption \
    --exclude-system \
    --exclude-empty
```

---

### 2. Single Mode (dari cmd layer)

```go
var singleCmd = &cobra.Command{
    Use:   "single",
    Short: "Backup satu database dengan optional companions",
    RunE: func(cmd *cobra.Command, args []string) error {
        return backup.ExecuteBackup(cmd, types.Deps, "single")
    },
}
```

**CLI Usage**:
```bash
sfdbtools backup single \
    --profile /path/to/profile.cnf.enc \
    --database mydb \
    --include-dmart \
    --include-temp \
    --compression \
    --encryption
```

**Interactive Mode** (tanpa `--database` flag):
```bash
sfdbtools backup single --profile /path/to/profile.cnf.enc
# Will show interactive menu to select database
```

---

### 3. Separated Mode

```bash
sfdbtools backup separated \
    --profile /path/to/profile.cnf.enc \
    --include db1,db2,db3 \
    --compression \
    --compression-type xz
```

**Output**:
```
/backup/20250101/
├── db1_separated_server1_20250101_120000.sql.xz
├── db2_separated_server1_20250101_120000.sql.xz
├── db3_separated_server1_20250101_120000.sql.xz
└── db1_separated_server1_20250101_120000_users.cnf
```

---

### 4. Programmatic Usage (dari code)

```go
import (
    "context"
    "sfDBTools/internal/backup"
    "sfDBTools/internal/types/types_backup"
)

// Setup options
opts := &types_backup.BackupDBOptions{
    Mode: "combined",
    Profile: types.ProfileInfo{
        Path: "/path/to/profile.cnf.enc",
    },
    Compression: types_backup.CompressionOptions{
        Enabled: true,
        Type:    "zstd",
        Level:   3,
    },
    Encryption: types_backup.EncryptionOptions{
        Enabled: true,
        Key:     "your-encryption-key",
    },
    Filter: types.FilterOptions{
        ExcludeSystem: true,
        ExcludeEmpty:  true,
    },
}

// Create service
svc := backup.NewBackupService(logger, config, opts)

// Setup context
ctx, cancel := context.WithCancel(context.Background())
defer cancel()

// Execute backup
backupConfig := types_backup.BackupEntryConfig{
    HeaderTitle: "Database Backup - Combined",
    Force:       false,
    BackupMode:  "combined",
}

if err := svc.ExecuteBackupCommand(ctx, backupConfig); err != nil {
    log.Fatalf("Backup failed: %v", err)
}
```

---

## Best Practices

### 1. Graceful Shutdown
- Selalu setup signal handler untuk SIGINT/SIGTERM
- Gunakan context cancellation untuk stop proses
- Cleanup partial files via `HandleShutdown()`

### 2. Error Handling
- Catch dan log semua errors
- Distinguish fatal vs non-fatal mysqldump errors
- Cleanup resources di defer blocks

### 3. Resource Management
- Close database connections di defer
- Close writer pipeline dari tail ke head (reverse order)
- Flush buffered writers sebelum close

### 4. Performance
- Gunakan buffered I/O (256KB buffer)
- Parallel compression dengan pgzip untuk large files
- Stream data, avoid load semua ke memory

### 5. Security
- Enkripsi passwords di profile files
- Use AES-256-GCM untuk backup encryption
- Prompt encryption key jika tidak tersedia di env var

### 6. Testing
- Mock `BackupService` interface untuk unit testing mode executors
- Test dengan berbagai kombinasi flags
- Test graceful shutdown scenarios

---

## Troubleshooting

### Common Issues

**1. "gagal mendapatkan kunci enkripsi"**
- Set `SFDB_BACKUP_ENCRYPTION_KEY` environment variable
- Atau gunakan flag `--encryption-key`
- Atau input interactively saat prompt

**2. "database tidak ditemukan di server"**
- Check spelling database name
- Check database exists: `sfdbtools db-scan all`
- Check profile credentials

**3. "mysqldump gagal: exit status 2"**
- Check mysqldump installed: `which mysqldump`
- Check database connectivity
- Check user permissions: `GRANT SELECT, LOCK TABLES ON *.* TO 'user'@'host'`

**4. "Tidak ada database yang tersedia setelah filtering"**
- Review exclude/include lists
- Check exclude-system/exclude-empty flags
- Check filter stats output untuk detail

**5. Backup timeout**
- Increase context timeout
- Check database size vs network speed
- Consider separated mode untuk large databases

---

## Dependencies

### Internal Dependencies
- `internal/appconfig` - Configuration management
- `internal/applog` - Logging interface
- `internal/cleanup` - Cleanup old backups
- `internal/types` - Shared type definitions
- `internal/types/types_backup` - Backup-specific types

### Package Dependencies
- `pkg/database` - Database client wrapper
- `pkg/compress` - Compression utilities
- `pkg/encrypt` - Encryption utilities
- `pkg/helper` - General utilities
- `pkg/profilehelper` - Profile loading
- `pkg/servicehelper` - Base service functionality
- `pkg/ui` - Terminal UI components
- `pkg/validation` - Input validation
- `pkg/backuphelper` - Backup-specific helpers

### External Dependencies
- `github.com/spf13/cobra` - CLI framework
- `github.com/briandowns/spinner` - Progress spinners
- `github.com/olekukonko/tablewriter` - Table rendering
- `github.com/AlecAivazis/survey/v2` - Interactive prompts

---

## Changelog

### 2025-12-05
- Konsolidasi execution logic ke mode executors (combined/separated/single)
- Refactor duplicate code ke helper functions
- Implement BackupService interface untuk better separation of concerns
- Add graceful shutdown handling dengan partial file cleanup
- Improve error handling dengan BackupErrorHandler
- Add metadata generation dengan builder pattern
- Modular display components (options & results)

---

## Future Enhancements

1. **Parallel Backup**: Backup multiple databases simultaneously
2. **Incremental Backup**: Track changes dengan binary log position
3. **Cloud Upload**: Auto-upload ke S3/GCS after backup
4. **Notification**: Email/Slack notification on completion
5. **Backup Verification**: Verify backup integrity dengan checksum
6. **Restore Testing**: Auto-test restore dari backup file
7. **Metrics**: Prometheus metrics export untuk monitoring
8. **Web UI**: Dashboard untuk backup management

---

## Lihat Juga

- [Profile Management Documentation](./profile-package.md)
- [Database Package Documentation](./database-package.md)
- [Restore Package Documentation](./restore-package.md)
- [Cleanup Package Documentation](./cleanup-package.md)

---

**Dokumentasi ini dibuat**: 2025-12-08  
**Versi sfDBTools**: Latest  
**Author**: AI Assistant dengan review dari Hadiyatna Muflihun
