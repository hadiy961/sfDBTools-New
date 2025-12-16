# Developer Manual - Backup Feature sfDBTools

## Daftar Isi

1. [Arsitektur Sistem](#arsitektur-sistem)
2. [Struktur Kode](#struktur-kode)
3. [Core Components](#core-components)
4. [Mode Executors](#mode-executors)
5. [Helper Packages](#helper-packages)
6. [Data Flow](#data-flow)
7. [Dependency Injection](#dependency-injection)
8. [Extension Guide](#extension-guide)
9. [Testing Strategy](#testing-strategy)

---

## Arsitektur Sistem

### Layer Architecture

```
┌─────────────────────────────────────────────────────────┐
│                     CLI Layer (Cobra)                    │
│  cmd/cmd_backup/*.go                                     │
│  - Parsing flags                                         │
│  - Command registration                                  │
│  - User input validation                                 │
└────────────────┬────────────────────────────────────────┘
                 │
                 ▼
┌─────────────────────────────────────────────────────────┐
│              Command Execution Layer                     │
│  internal/backup/command.go                              │
│  - ExecuteBackup(cmd, deps, mode)                       │
│  - Unified entry point                                   │
│  - Options parsing                                       │
└────────────────┬────────────────────────────────────────┘
                 │
                 ▼
┌─────────────────────────────────────────────────────────┐
│                  Service Layer                           │
│  internal/backup/service.go                              │
│  - Business logic orchestration                          │
│  - State management                                      │
│  - Interface implementation                              │
└────────────────┬────────────────────────────────────────┘
                 │
                 ▼
┌─────────────────────────────────────────────────────────┐
│              Mode Executor Factory                       │
│  internal/backup/modes/factory.go                        │
│  - GetExecutor(mode, service)                           │
│  - Returns ModeExecutor interface                        │
└────────────────┬────────────────────────────────────────┘
                 │
                 ├─────────────┬──────────────┐
                 ▼             ▼              ▼
         ┌──────────┐  ┌────────────┐  ┌─────────────┐
         │Combined  │  │ Iterative  │  │  Future     │
         │Executor  │  │ Executor   │  │  Modes...   │
         └──────────┘  └────────────┘  └─────────────┘
```

### Pattern: Strategy + Factory

Mode backup menggunakan **Strategy Pattern** dengan **Factory Pattern**:

1. **ModeExecutor Interface**: Contract untuk semua mode
2. **Factory**: `modes.GetExecutor()` returns implementasi yang sesuai
3. **Concrete Strategies**: `CombinedExecutor`, `IterativeExecutor`

**Keuntungan**:
- Easy to add new modes
- Mode logic terisolasi
- Testable independently
- No conditional hell

---

## Struktur Kode

### Directory Layout

```
internal/backup/
├── command.go              # Command layer entry points
├── executor.go             # Main execution logic
├── service.go              # Service struct & interface impl
├── service_helpers.go      # Service helper methods
├── setup.go                # Setup & preparation logic
├── writer.go               # Writer pipeline & mysqldump
├── mode_config.go          # Mode configuration
├── display/                # UI display components
│   ├── options_display.go  # Display backup options
│   └── result_display.go   # Display backup results
├── filehelper/             # File operations
│   └── file_helper.go      # Path generation, file ops
├── metadata/               # Metadata generation
│   ├── backup_metadata.go  # Metadata CRUD
│   └── user.go            # User grants export
└── modes/                  # Mode executors
    ├── interface.go        # ModeExecutor & BackupService
    ├── factory.go          # GetExecutor factory
    ├── combined.go         # Combined mode executor
    └── iterative.go        # Iterative mode executor

cmd/cmd_backup/
├── cmd_backup_main.go      # Parent command
├── cmd_backup_all.go       # All mode command
├── cmd_backup_filter.go    # Filter mode command
├── cmd_backup_single.go    # Single mode command
├── cmd_backup_primary.go   # Primary mode command
└── cmd_backup_secondary.go # Secondary mode command
```

### Package Dependencies

```
cmd/cmd_backup
    ↓
internal/backup (command.go)
    ↓
internal/backup (service.go)
    ↓
internal/backup/modes
    ↓
pkg/backuphelper
pkg/database
pkg/compress
pkg/encrypt
pkg/profilehelper
```

---

## Core Components

### 1. Service Struct

**File**: `internal/backup/service.go`

```go
type Service struct {
    servicehelper.BaseService  // Embed: mutex, cancel func
    
    Config          *appconfig.Config
    Log             applog.Logger
    ErrorLog        *errorlog.ErrorLogger
    DBInfo          *types.DBInfo
    Profile         *types.ProfileInfo
    BackupDBOptions *types_backup.BackupDBOptions
    BackupEntry     *types_backup.BackupEntryConfig
    Client          *database.Client
    
    // State management
    currentBackupFile string
    backupInProgress  bool
    gtidInfo          *database.GTIDInfo
    excludedDatabases []string
}
```

**Key Methods**:

| Method | Purpose |
|--------|---------|
| `NewBackupService()` | Constructor dengan type switching |
| `ExecuteBackup()` | Main entry point |
| `ExecuteBackupCommand()` | Command execution wrapper |
| `SetCurrentBackupFile()` | Track backup in progress |
| `HandleShutdown()` | Graceful shutdown (CTRL+C) |

**Interface Implementation**:
Service implements `modes.BackupService` interface untuk decoupling.

---

### 2. Command Execution Flow

**File**: `internal/backup/command.go`

```go
// Unified entry point
func ExecuteBackup(cmd *cobra.Command, deps *types.Dependencies, mode string) error

// Internal flow
func executeBackupWithConfig(cmd, deps, config) error
    ↓ Parse options
    ↓ Create service
    ↓ Setup graceful shutdown
    ↓ Execute backup command
    ↓ Display results
```

**Graceful Shutdown**:
```go
// Setup signal handler
sigChan := make(chan os.Signal, 1)
signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

go func() {
    <-sigChan
    svc.HandleShutdown()  // Cleanup partial files
}()
```

---

### 3. Executor Logic

**File**: `internal/backup/executor.go`

```go
func (s *Service) ExecuteBackup(ctx, client, databases, mode) (*BackupResult, error) {
    // 1. Setup execution
    s.SetupBackupExecution()
    
    // 2. Route to mode executor
    result := s.executeBackupByMode(ctx, databases, mode)
    
    // 3. Auto cleanup old backups
    if s.Config.Backup.Cleanup.Enabled {
        cleanup.CleanupOldBackupsFromBackup()
    }
    
    // 4. Handle errors
    return s.handleBackupErrors(result)
}
```

**Mode Routing**:
```go
func (s *Service) executeBackupByMode(ctx, databases, mode) BackupResult {
    executor, _ := modes.GetExecutor(mode, s)
    return executor.Execute(ctx, databases)
}
```

---

### 4. Writer Pipeline

**File**: `internal/backup/writer.go`

```go
// Stream processing chain
mysqldump stdout → Compression → Encryption → Buffer → File

func (s *Service) createWriterPipeline(baseWriter, compress, encrypt) {
    writer := baseWriter
    
    // Layer 1: Encryption (closest to file)
    if encrypt {
        writer = encrypt.NewEncryptingWriter(writer, key)
    }
    
    // Layer 2: Compression
    if compress {
        writer = compress.NewWriter(writer, type, level)
    }
    
    return writer
}
```

**Mysqldump Execution**:
```go
func (s *Service) executeMysqldumpWithPipe(ctx, args, output, compress, type) {
    // Create output file with buffered writer
    file, bufWriter := s.createBufferedOutputFile(output)
    
    // Setup writer pipeline
    writer, closers := s.createWriterPipeline(bufWriter, compress, type, key)
    
    // Execute command
    cmd := exec.CommandContext(ctx, "mysqldump", args...)
    cmd.Stdout = writer
    cmd.Stderr = &stderrBuf
    
    cmd.Run()
}
```

---

## Mode Executors

### Interface Definition

**File**: `internal/backup/modes/interface.go`

```go
// ModeExecutor - Contract untuk semua mode
type ModeExecutor interface {
    Execute(ctx context.Context, databases []string) types_backup.BackupResult
}

// BackupService - Dependencies untuk executors
type BackupService interface {
    // Logging
    LogInfo(msg string)
    LogDebug(msg string)
    LogWarn(msg string)
    LogError(msg string)
    GetLogger() applog.Logger
    
    // Execution
    ExecuteAndBuildBackup(ctx, cfg) (DatabaseBackupInfo, error)
    ExecuteBackupLoop(ctx, dbs, config, pathFunc) BackupLoopResult
    
    // Helpers
    GetBackupOptions() *BackupDBOptions
    GenerateFullBackupPath(dbName, mode) (string, error)
    GetTotalDatabaseCount(ctx, dbs) int
    CaptureAndSaveGTID(ctx, backupFile) error
    ExportUserGrantsIfNeeded(ctx, backupFile, dbs) string
    UpdateMetadataUserGrantsPath(backupFile, userGrantsPath)
    ToBackupResult(loopResult) BackupResult
}
```

---

### Combined Executor

**File**: `internal/backup/modes/combined.go`

**Usage**: Mode `all`, `filter --mode=single-file`

**Characteristics**:
- Backup semua database dalam **satu file**
- GTID capture untuk snapshot point
- User grants export (semua user atau filtered)

**Flow**:
```go
func (e *CombinedExecutor) Execute(ctx, databases) BackupResult {
    // 1. Count total databases
    totalDBFound := e.service.GetTotalDatabaseCount(ctx, databases)
    
    // 2. Generate output path
    outputPath := filepath.Join(outputDir, filename)
    
    // 3. Capture GTID sebelum backup
    e.service.CaptureAndSaveGTID(ctx, outputPath)
    
    // 4. Execute backup (all databases in one file)
    backupInfo := e.service.ExecuteAndBuildBackup(ctx, BackupExecutionConfig{
        DBList:     databases,
        OutputPath: outputPath,
        BackupType: mode,
        IsMultiDB:  true,
    })
    
    // 5. Export user grants
    userGrantsPath := e.service.ExportUserGrantsIfNeeded(ctx, outputPath, databases)
    
    // 6. Update metadata
    e.service.UpdateMetadataUserGrantsPath(outputPath, userGrantsPath)
    
    return result
}
```

---

### Iterative Executor

**File**: `internal/backup/modes/iterative.go`

**Usage**: Mode `single`, `primary`, `secondary`, `filter --mode=multi-file`

**Characteristics**:
- Backup per database secara **berurutan**
- TIDAK capture GTID (tidak ada snapshot point global)
- User grants per database atau agregat

**Flow**:
```go
func (e *IterativeExecutor) Execute(ctx, databases) BackupResult {
    // 1. Create output path function
    outputPathFunc := e.createOutputPathFunc(databases)
    
    // 2. Execute backup loop
    loopResult := e.service.ExecuteBackupLoop(ctx, databases, BackupLoopConfig{
        Mode:       e.mode,
        TotalDBs:   len(databases),
        BackupType: e.mode,
    }, outputPathFunc)
    
    // 3. For primary/secondary: aggregate and export user grants
    if e.mode == "primary" || e.mode == "secondary" {
        userGrantsPath := e.service.ExportUserGrantsIfNeeded(ctx, loopResult.BackupInfos[0].OutputFile, databases)
        e.service.UpdateMetadataUserGrantsPath(loopResult.BackupInfos[0].OutputFile, userGrantsPath)
        e.generateCombinedMetadata(ctx, loopResult, databases)
    }
    
    // 4. Convert to BackupResult
    return e.service.ToBackupResult(loopResult)
}
```

**Output Path Logic**:
```go
func (e *IterativeExecutor) createOutputPathFunc(databases) func(string) (string, error) {
    return func(dbName string) (string, error) {
        // Single Mode Variant: first database can use custom filename
        if IsSingleModeVariant(e.mode) && databases[0] == dbName && customFilename != "" {
            return filepath.Join(outputDir, customFilename), nil
        }
        
        // Default: generate path
        return e.service.GenerateFullBackupPath(dbName, e.mode)
    }
}
```

---

### Factory Pattern

**File**: `internal/backup/modes/factory.go`

```go
func GetExecutor(mode string, svc BackupService) (ModeExecutor, error) {
    switch mode {
    case "combined", "all":
        return NewCombinedExecutor(svc), nil
    case "single", "primary", "secondary", "separated", "separate":
        return NewIterativeExecutor(svc, mode), nil
    default:
        return nil, fmt.Errorf("mode tidak dikenali: %s", mode)
    }
}
```

**Adding New Mode**:
1. Create new executor implementing `ModeExecutor`
2. Add case in `GetExecutor()`
3. Implement `Execute()` method
4. Add command in `cmd/cmd_backup/`

---

## Helper Packages

### 1. pkg/backuphelper

**Purpose**: Pure logic helpers untuk backup operations

**Files**:
- `mysqldump.go`: Mysqldump args builder
- `logic.go`: Database filtering, version extraction

**Key Functions**:

```go
// Build mysqldump arguments
func BuildMysqldumpArgs(baseDumpArgs string, conn DatabaseConn, filter FilterOptions, dbFiltered []string, singleDB string, totalDBFound int) []string

// Check fatal mysqldump errors
func IsFatalMysqldumpError(err error, stderrOutput string) bool

// Filter candidates by mode
func FilterCandidatesByMode(databases []string, mode string) []string

// Extract mysqldump version
func ExtractMysqldumpVersion(stderrOutput string) string
```

---

### 2. pkg/database

**Purpose**: Database client wrapper dengan connection pooling

**Key Files**:
- `database_connection.go`: Connection management
- `database_filter.go`: Database filtering logic
- `database_gtid.go`: GTID operations
- `database_user.go`: User grants export

**Client Struct**:
```go
type Client struct {
    db               *sql.DB
    config           Config
    logger           applog.Logger
    queryTimeout     time.Duration
    maxRetries       int
    retryDelay       time.Duration
}
```

**Key Methods**:

| Method | Purpose |
|--------|---------|
| `NewClient()` | Create client with connection pooling |
| `GetDatabases()` | List all databases |
| `FilterDatabases()` | Apply filters |
| `GetGTIDExecuted()` | Get GTID info |
| `ExportAllUserGrants()` | Export all user grants |
| `ExportUserGrantsForDatabases()` | Export filtered grants |

---

### 3. pkg/compress

**Purpose**: Multi-format compression support

**Supported Formats**:
- gzip
- zstd
- xz
- pgzip (parallel gzip)
- zlib

**Interface**:
```go
func NewWriter(w io.Writer, compressionType CompressionType, level int) (io.WriteCloser, error)

func NewReader(r io.Reader, compressionType CompressionType) (io.ReadCloser, error)
```

---

### 4. pkg/encrypt

**Purpose**: AES-256-GCM encryption (OpenSSL compatible)

**Key Functions**:
```go
// Create encrypting writer
func NewEncryptingWriter(w io.Writer, key []byte) (io.WriteCloser, error)

// Create decrypting reader
func NewDecryptingReader(r io.Reader, key []byte) (io.ReadCloser, error)
```

**Format**:
- Header: "Salted__" (8 bytes)
- Salt: 8 bytes
- PBKDF2: 100,000 iterations
- AES-256-GCM

---

### 5. pkg/profilehelper

**Purpose**: Unified profile loading and connection

**Key Functions**:
```go
// Load source profile with interactive selector
func LoadSourceProfile(profilePath, encryptionKey string, allowInteractive bool) (*types.ProfileInfo, error)

// Connect using profile
func ConnectWithProfile(profile *types.ProfileInfo, dbName string) (*database.Client, error)

// Connect to target (for restore)
func ConnectWithTargetProfile(profile *types.ProfileInfo, dbName string) (*database.Client, error)
```

**Eliminates**: 94 lines of duplication across backup/restore/dbscan

---

### 6. pkg/servicehelper

**Purpose**: Base service functionality

**BaseService**:
```go
type BaseService struct {
    mu         sync.Mutex
    cancelFunc context.CancelFunc
}

// Mutex locking
func (b *BaseService) WithLock(fn func())

// Cancel function management
func (b *BaseService) SetCancelFunc(cancel context.CancelFunc)
func (b *BaseService) Cancel()
```

**TrackProgress**:
```go
// Progress tracking with defer pattern
func TrackProgress(svc ProgressTracker) func() {
    start := time.Now()
    svc.IncrementProgress()
    
    return func() {
        svc.UpdateProgress(time.Since(start))
    }
}

// Usage in service
defer servicehelper.TrackProgress(s)()
```

---

### 7. pkg/ui

**Purpose**: Terminal UI components

**Key Functions**:
```go
// Headers with banner
func Headers(title string)

// Subheader
func PrintSubHeader(text string)

// Success/Error messages
func PrintSuccess(msg string)
func PrintError(msg string)

// Spinner with elapsed time
func NewSpinnerWithElapsed(text string) *SpinnerWithElapsed
```

---

## Data Flow

### Complete Backup Flow Diagram

```
┌─────────────────────────────────────────────────────────────────┐
│ 1. CLI Layer: cobra.Command.Run()                               │
│    - Parse flags                                                 │
│    - Validate input                                              │
└───────────────────────────┬─────────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────────┐
│ 2. backup.ExecuteBackup(cmd, deps, mode)                        │
│    - Get execution config                                        │
│    - Parse backup options                                        │
│    - Create backup service                                       │
└───────────────────────────┬─────────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────────┐
│ 3. service.ExecuteBackupCommand(ctx, config)                    │
│    - PrepareBackupSession()                                      │
│      • Load profile                                              │
│      • Connect to database                                       │
│      • Get filtered databases                                    │
│      • Generate backup paths                                     │
│      • Display options (if --force)                              │
└───────────────────────────┬─────────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────────┐
│ 4. service.ExecuteBackup(ctx, client, databases, mode)          │
│    - SetupBackupExecution()                                      │
│      • Prompt for ticket                                         │
│      • Create output directory                                   │
│      • Setup encryption                                          │
│    - executeBackupByMode()                                       │
│      • Get executor from factory                                 │
│      • executor.Execute()                                        │
└───────────────────────────┬─────────────────────────────────────┘
                            │
                ┌───────────┴────────────┐
                │                        │
                ▼                        ▼
    ┌──────────────────────┐  ┌──────────────────────┐
    │ CombinedExecutor     │  │ IterativeExecutor    │
    │ Execute()            │  │ Execute()            │
    └──────────┬───────────┘  └─────────┬────────────┘
               │                        │
               │                        │
               ▼                        ▼
┌────────────────────────────────────────────────────────────────┐
│ 5. Mode-specific Execution                                     │
│                                                                 │
│ Combined:                      Iterative:                      │
│  ├─ CaptureGTID()              ├─ Loop databases              │
│  ├─ ExecuteAndBuildBackup()    │   ├─ Generate output path   │
│  │   └─ mysqldump all DBs      │   ├─ ExecuteAndBuildBackup()│
│  ├─ ExportUserGrants()         │   │   └─ mysqldump single DB│
│  └─ UpdateMetadata()           │   ├─ ExportUserGrants()     │
│                                 │   └─ UpdateMetadata()       │
│                                 └─ AggregateResults()          │
└────────────────────────┬───────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────────────┐
│ 6. Backup Execution (executeAndBuildBackup)                     │
│    - BuildMysqldumpArgs()                                        │
│    - executeMysqldumpWithPipe()                                  │
│      • Create output file                                        │
│      • Setup writer pipeline (compress + encrypt)                │
│      • Execute mysqldump                                         │
│      • Stream: mysqldump → compress → encrypt → file            │
│    - buildBackupInfoFromResult()                                 │
└────────────────────────┬────────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────────────┐
│ 7. Post-Backup Operations                                       │
│    - Generate metadata JSON                                      │
│    - Export user grants (if not excluded)                        │
│    - Cleanup old backups (if enabled)                            │
└────────────────────────┬────────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────────────┐
│ 8. Display Results                                               │
│    - Display backup summary table                                │
│    - Show GTID info (if captured)                                │
│    - Show file sizes                                             │
│    - Show duration                                               │
└─────────────────────────────────────────────────────────────────┘
```

### Writer Pipeline Detail

```
mysqldump Process
      │
      │ stdout (SQL stream)
      ▼
┌─────────────────┐
│ Raw SQL Stream  │
└────────┬────────┘
         │
         ▼
┌──────────────────────────┐
│ Encryption Writer (opt)  │
│ - AES-256-GCM            │
│ - PBKDF2 key derivation  │
│ - Salted header          │
└────────┬─────────────────┘
         │
         ▼
┌──────────────────────────┐
│ Compression Writer (opt) │
│ - gzip/zstd/xz/pgzip     │
│ - Configurable level     │
└────────┬─────────────────┘
         │
         ▼
┌──────────────────────────┐
│ Buffered Writer          │
│ - 256KB buffer           │
│ - Reduce I/O operations  │
└────────┬─────────────────┘
         │
         ▼
┌──────────────────────────┐
│ File Writer              │
│ - os.Create()            │
│ - Write to disk          │
└──────────────────────────┘
```

---

## Dependency Injection

### Global Dependencies Pattern

**File**: `internal/types/types.go`

```go
// Global dependency injection
var Deps *Dependencies

type Dependencies struct {
    Config *appconfig.Config
    Logger applog.Logger
}
```

**Injection Flow**:
```go
// main.go
func main() {
    cfg := appconfig.LoadConfigFromEnv()
    logger := applog.NewLogger(cfg)
    
    types.Deps = &types.Dependencies{
        Config: cfg,
        Logger: logger,
    }
    
    cmd.Execute(types.Deps)
}

// cmd layer
func ExecuteBackup(cmd *cobra.Command, deps *types.Dependencies, mode string) error {
    if deps == nil {
        return errors.New("dependencies not available")
    }
    
    service := NewBackupService(deps.Logger, deps.Config, options)
    // ...
}
```

**Benefits**:
- Centralized configuration
- Easy testing (mock Deps)
- No global config variables
- Clear dependency graph

---

## Extension Guide

### Adding New Backup Mode

**Steps**:

1. **Create Executor**:
```go
// File: internal/backup/modes/mymode.go
type MyModeExecutor struct {
    service BackupService
}

func NewMyModeExecutor(svc BackupService) *MyModeExecutor {
    return &MyModeExecutor{service: svc}
}

func (e *MyModeExecutor) Execute(ctx context.Context, databases []string) types_backup.BackupResult {
    // Implement backup logic
    result := types_backup.BackupResult{}
    
    // Use service methods
    backupInfo := e.service.ExecuteAndBuildBackup(ctx, config)
    
    return result
}
```

2. **Register in Factory**:
```go
// File: internal/backup/modes/factory.go
func GetExecutor(mode string, svc BackupService) (ModeExecutor, error) {
    switch mode {
    // ... existing cases
    case "mymode":
        return NewMyModeExecutor(svc), nil
    default:
        return nil, fmt.Errorf("mode tidak dikenali: %s", mode)
    }
}
```

3. **Add Command**:
```go
// File: cmd/cmd_backup/cmd_backup_mymode.go
var CmdBackupMyMode = &cobra.Command{
    Use:   "mymode",
    Short: "My custom backup mode",
    Run: func(cmd *cobra.Command, args []string) {
        if types.Deps == nil {
            fmt.Println("✗ Dependencies tidak tersedia")
            return
        }
        
        if err := backup.ExecuteBackup(cmd, types.Deps, "mymode"); err != nil {
            types.Deps.Logger.Error("backup mymode gagal: " + err.Error())
        }
    },
}

func init() {
    defaultOpts := defaultVal.DefaultBackupOptions("mymode")
    flags.AddBackupFlags(CmdBackupMyMode, &defaultOpts)
}
```

4. **Register Command**:
```go
// File: cmd/cmd_backup/cmd_backup_main.go
func init() {
    CmdDBBackupMain.AddCommand(CmdBackupMyMode)
}
```

---

### Adding New Compression Type

**Steps**:

1. **Add Constant**:
```go
// File: pkg/compress/compress.go
const (
    CompressionTypeMyCompressor CompressionType = "mycompressor"
)
```

2. **Implement Writer**:
```go
func newMyCompressorWriter(w io.Writer, level int) (io.WriteCloser, error) {
    return mycompressor.NewWriter(w, level)
}
```

3. **Register in Factory**:
```go
func NewWriter(w io.Writer, compressionType CompressionType, level int) (io.WriteCloser, error) {
    switch compressionType {
    // ... existing cases
    case CompressionTypeMyCompressor:
        return newMyCompressorWriter(w, level)
    default:
        return nil, fmt.Errorf("compression type not supported")
    }
}
```

4. **Add to File Extension Mapping**:
```go
// File: pkg/helper/helper_compression.go
func GetCompressionExtension(compressionType compress.CompressionType) string {
    switch compressionType {
    // ... existing cases
    case compress.CompressionTypeMyCompressor:
        return ".mycomp"
    default:
        return ""
    }
}
```

---

### Extending Service Methods

**Pattern**: Add helper methods in `service_helpers.go`

```go
// File: internal/backup/service_helpers.go

// MyCustomOperation performs custom backup operation
func (s *Service) MyCustomOperation(ctx context.Context, params MyParams) error {
    s.Log.Info("Performing custom operation...")
    
    // Use existing helpers
    timer := pkghelper.NewTimer()
    
    // Your logic here
    
    duration := timer.Elapsed()
    s.Log.Infof("Operation completed in %v", duration)
    
    return nil
}
```

**Best Practices**:
- Keep business logic in service layer
- Use helper packages for pure functions
- Log important operations
- Use timer for duration tracking
- Return errors, don't panic

---

## Testing Strategy

### Unit Testing Structure

```
internal/backup/
├── executor_test.go
├── service_test.go
├── writer_test.go
└── modes/
    ├── combined_test.go
    ├── iterative_test.go
    └── factory_test.go
```

### Mock Service for Mode Testing

```go
// File: internal/backup/modes/mock_service_test.go
type MockBackupService struct {
    backupOptions *types_backup.BackupDBOptions
    executeError  error
    loopResult    types_backup.BackupLoopResult
}

func (m *MockBackupService) LogInfo(msg string) {}
func (m *MockBackupService) GetBackupOptions() *types_backup.BackupDBOptions {
    return m.backupOptions
}
func (m *MockBackupService) ExecuteAndBuildBackup(ctx context.Context, cfg types_backup.BackupExecutionConfig) (types.DatabaseBackupInfo, error) {
    if m.executeError != nil {
        return types.DatabaseBackupInfo{}, m.executeError
    }
    return types.DatabaseBackupInfo{DatabaseName: cfg.DBName}, nil
}
// ... implement other methods
```

### Example Test

```go
// File: internal/backup/modes/combined_test.go
func TestCombinedExecutor_Execute(t *testing.T) {
    tests := []struct {
        name      string
        databases []string
        wantErr   bool
    }{
        {
            name:      "backup multiple databases",
            databases: []string{"db1", "db2", "db3"},
            wantErr:   false,
        },
        {
            name:      "empty database list",
            databases: []string{},
            wantErr:   false,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            mockSvc := &MockBackupService{
                backupOptions: &types_backup.BackupDBOptions{
                    OutputDir: "/tmp/backup",
                    File: types_backup.FileOptions{
                        Path: "backup.sql",
                    },
                },
            }
            
            executor := NewCombinedExecutor(mockSvc)
            result := executor.Execute(context.Background(), tt.databases)
            
            if tt.wantErr && len(result.Errors) == 0 {
                t.Error("expected error but got none")
            }
            if !tt.wantErr && len(result.Errors) > 0 {
                t.Errorf("unexpected error: %v", result.Errors)
            }
        })
    }
}
```

### Integration Testing

```bash
# Test backup dengan dry-run
go run main.go backup all --dry-run --ticket TEST-001

# Test dengan mock database
docker run -d --name test-mysql -e MYSQL_ROOT_PASSWORD=test mysql:8.0
go test -v ./internal/backup/... -integration
```

---

## Best Practices

### 1. Error Handling

```go
// ✅ Good: Wrap errors with context
func (s *Service) DoSomething() error {
    if err := operation(); err != nil {
        return fmt.Errorf("failed to do something: %w", err)
    }
    return nil
}

// ❌ Bad: Lose error context
func (s *Service) DoSomething() error {
    if err := operation(); err != nil {
        return err
    }
    return nil
}
```

### 2. Logging

```go
// ✅ Good: Structured logging
s.Log.Infof("Backup started: database=%s, mode=%s", dbName, mode)

// ❌ Bad: Concatenated strings
s.Log.Info("Backup started: " + dbName + " mode: " + mode)
```

### 3. Resource Cleanup

```go
// ✅ Good: Defer cleanup
file, err := os.Create(path)
if err != nil {
    return err
}
defer file.Close()

// Track success for conditional cleanup
var success bool
defer func() {
    if !success {
        os.Remove(path)
    }
}()

// Do work...
success = true
```

### 4. Context Usage

```go
// ✅ Good: Pass context
func (s *Service) DoBackup(ctx context.Context, db string) error {
    cmd := exec.CommandContext(ctx, "mysqldump", args...)
    return cmd.Run()
}

// ❌ Bad: Ignore context
func (s *Service) DoBackup(db string) error {
    cmd := exec.Command("mysqldump", args...)
    return cmd.Run()
}
```

### 5. Configuration

```go
// ✅ Good: Use config struct
type BackupConfig struct {
    OutputDir   string
    Compression CompressionOptions
    Encryption  EncryptionOptions
}

// ❌ Bad: Many parameters
func Backup(outputDir string, compressType string, compressLevel int, encrypt bool, key string)
```

---

## Performance Considerations

### 1. Streaming vs Buffering

**Current**: Streaming architecture untuk memory efficiency
```go
// Streaming: constant memory usage
mysqldump → compress → encrypt → file
```

**Alternative**: In-memory buffering (faster but more memory)
```go
// Buffer all data first, then write
data := mysqldump.ReadAll()
compressed := compress(data)
encrypted := encrypt(compressed)
file.Write(encrypted)
```

### 2. Parallel Backup (Future)

```go
// Current: Sequential backup
for db in databases {
    backup(db)
}

// Future: Parallel with worker pool
pool := worker.NewPool(maxWorkers)
for db in databases {
    pool.Submit(func() {
        backup(db)
    })
}
pool.Wait()
```

### 3. Compression Benchmarks

| Type | Speed | Ratio | CPU | Use Case |
|------|-------|-------|-----|----------|
| gzip | Medium | Good | Medium | Default |
| zstd | Fast | Very Good | Low | Recommended |
| xz | Slow | Best | High | Archive |
| pgzip | Fast | Good | High | Large files |

---

## Troubleshooting Development

### Common Issues

**1. Dependencies Not Available**
```go
// Check Deps before use
if types.Deps == nil {
    return fmt.Errorf("dependencies not initialized")
}
```

**2. Interface Not Satisfied**
```bash
# Verify interface implementation at compile time
var _ modes.BackupService = (*Service)(nil)
```

**3. Memory Leaks**
```go
// Always close resources
defer client.Close()
defer file.Close()
defer writer.Close()
```

**4. Goroutine Leaks**
```go
// Use context for cancellation
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
defer cancel()
```

---

## API Reference

### Core Service Methods

```go
// Initialization
NewBackupService(logger, config, options) *Service

// Main execution
ExecuteBackup(ctx, client, databases, mode) (*BackupResult, error)
ExecuteBackupCommand(ctx, config) error

// Setup & preparation
SetupBackupExecution() error
PrepareBackupSession(ctx, title, showOptions) (client, databases, error)
CheckAndSelectConfigFile() error

// Backup operations
executeAndBuildBackup(ctx, config) (DatabaseBackupInfo, error)
executeBackupLoop(ctx, dbs, config, pathFunc) BackupLoopResult
executeMysqldumpWithPipe(ctx, args, output, compress, type) (*BackupWriteResult, error)

// Path generation
generateFullBackupPath(dbName, mode) (string, error)
GenerateFullBackupPath(dbName, mode) (string, error)

// GTID & metadata
captureAndSaveGTID(ctx, backupFile) error
exportUserGrantsIfNeeded(ctx, backupFile, databases) string
updateMetadataUserGrantsPath(backupFile, userGrantsPath)

// State management
SetCurrentBackupFile(filePath)
ClearCurrentBackupFile()
HandleShutdown()
```

### Mode Executor Interface

```go
type ModeExecutor interface {
    Execute(ctx context.Context, databases []string) BackupResult
}

type BackupService interface {
    // See modes/interface.go for full definition
}
```

---

## Glossary

| Term | Definition |
|------|------------|
| **Combined Mode** | Backup all databases in one file |
| **Iterative Mode** | Backup each database separately |
| **GTID** | Global Transaction Identifier for replication |
| **Writer Pipeline** | Chain of writers (compress → encrypt → file) |
| **Service Layer** | Business logic orchestration |
| **Mode Executor** | Strategy implementation for backup mode |
| **Graceful Shutdown** | Clean termination with partial file cleanup |
| **Companion Database** | Related databases (_dmart, _temp, _archive) |
| **Metadata** | JSON file describing backup details |
| **User Grants** | SQL file containing CREATE USER and GRANT statements |

---

**Version**: 1.0  
**Last Updated**: 2025-12-16  
**Author**: Hadiyatna Muflihun  
**Maintainer**: Development Team
