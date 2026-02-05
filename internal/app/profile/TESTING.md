# Testing Guide - Profile Package

Panduan lengkap untuk menjalankan dan menulis tests untuk package `profile`.

## Quick Start

```bash
# Run semua tests di package profile
go test ./internal/app/profile/...

# Run tests dengan verbose output
go test -v ./internal/app/profile/...

# Run tests dengan coverage
go test -cover ./internal/app/profile/...

# Run tests untuk subpackage tertentu
go test ./internal/app/profile/model/
go test ./internal/app/profile/validation/
go test ./internal/app/profile/helpers/...
```

## Test Structure

```
profile/
├── model/
│   ├── types_profile.go
│   └── types_profile_test.go       ✓ Unit tests
├── validation/
│   ├── profile.go
│   ├── profile_test.go            ✓ Unit tests
│   ├── database.go
│   ├── database_test.go           ✓ Unit tests
│   ├── input.go
│   └── input_test.go              ✓ Unit tests
└── helpers/
    ├── paths/
    │   ├── resolver.go
    │   └── resolver_test.go        ✓ Unit tests
    └── keys/
        ├── profile_key.go
        └── profile_key_test.go     ✓ Unit tests
```

## Test Coverage

### Priority 1 - COMPLETED ✓

**Model & State Management (model/)**
- ✓ `types_profile_test.go` - 15 test cases
  - ProfileOptions implementations (Mode, IsInteractive)
  - ProfileState.HasMeaningfulChanges()
  - ProfileState helpers (CreateOptions, EditOptions, etc.)
  - State mode detection

**Validation (validation/)**
- ✓ `input_test.go` - 13 test cases
  - ValidateNoLeadingTrailingSpace
  - ValidateNoControlChars
  - ValidateNoSpaces
  - ValidateNotEmpty
  - ValidateIntInRange
  - ValidateFileAccessible
  - ValidateConfigName
  - ValidateHost
  - ValidateUsername

- ✓ `database_test.go` - 18 test cases
  - ValidateDBInfo with various scenarios
  - Port validation (edge cases)
  - Empty fields handling
  - Nil checks

- ✓ `profile_test.go` - 12 test cases
  - ValidateProfileInfo comprehensive
  - SSH tunnel validation
  - Complex scenarios
  - Full featured profiles

**Total Priority 1: ~58 test cases**

### Priority 2 - COMPLETED ✓

**Path Resolution (helpers/paths/)**
- ✓ `resolver_test.go` - 7 test cases
  - PathResolver construction
  - Resolve logic
  - State independence
  - Edge cases

**Key Resolution (helpers/keys/)**
- ✓ `profile_key_test.go` - 9 test cases
  - ResolveProfileEncryptionKey
  - Priority order (flag > env > prompt)
  - Fallback mechanisms
  - Error scenarios

**Total Priority 2: ~16 test cases**

**Grand Total: ~74 test cases untuk Priority 1-2**

## Running Tests

### Run All Tests

```bash
# Semua tests
go test ./internal/app/profile/...

# Dengan race detector (detect race conditions)
go test -race ./internal/app/profile/...

# Dengan timeout
go test -timeout 30s ./internal/app/profile/...
```

### Run Specific Tests

```bash
# Run single test file
go test ./internal/app/profile/model/

# Run specific test by name
go test -run TestProfileState_HasMeaningfulChanges ./internal/app/profile/model/

# Run tests matching pattern
go test -run ".*State.*" ./internal/app/profile/model/
```

### Verbose Output

```bash
# Verbose: show all test names
go test -v ./internal/app/profile/...

# Very verbose: with test output
go test -v -count=1 ./internal/app/profile/...

# Show only failures
go test ./internal/app/profile/... 2>&1 | grep -A 5 FAIL
```

### Coverage Reports

```bash
# Simple coverage percentage
go test -cover ./internal/app/profile/...

# Detailed coverage per package
go test -coverprofile=coverage.out ./internal/app/profile/...
go tool cover -func=coverage.out

# HTML coverage report
go test -coverprofile=coverage.out ./internal/app/profile/...
go tool cover -html=coverage.out

# Coverage for specific package
go test -cover ./internal/app/profile/model/
go test -cover ./internal/app/profile/validation/
```

### Watch Mode (dengan external tool)

Install `entr` atau `watchexec`:

```bash
# Using watchexec
watchexec -e go -r go test ./internal/app/profile/...

# Using entr
find . -name "*.go" | entr -c go test ./internal/app/profile/...
```

## Test Patterns

### Table-Driven Tests

Digunakan untuk testing multiple input scenarios:

```go
func TestValidateHost(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        wantErr bool
    }{
        {"valid hostname", "database.example.com", false},
        {"valid IP", "10.0.0.5", false},
        {"empty host", "", true},
        {"with space", "host name", true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := ValidateHost(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

### Subtests

Untuk grouping related tests:

```go
func TestProfileState(t *testing.T) {
    t.Run("HasMeaningfulChanges", func(t *testing.T) {
        // Test HasMeaningfulChanges logic
    })
    
    t.Run("Mode detection", func(t *testing.T) {
        // Test Mode() method
    })
    
    t.Run("Interactive check", func(t *testing.T) {
        // Test IsInteractive() method
    })
}
```

### Test Fixtures

Untuk setup/cleanup:

```go
func TestWithTempFile(t *testing.T) {
    // Setup
    tmpDir := t.TempDir() // Auto cleanup
    tmpFile := filepath.Join(tmpDir, "test.txt")
    os.WriteFile(tmpFile, []byte("test"), 0644)
    
    // Test
    // ... your test logic
    
    // Cleanup automatic via t.TempDir()
}
```

### Environment Variable Tests

```go
func TestFromEnv(t *testing.T) {
    // Setup
    original := os.Getenv("MY_VAR")
    os.Setenv("MY_VAR", "test-value")
    
    // Cleanup
    defer func() {
        if original != "" {
            os.Setenv("MY_VAR", original)
        } else {
            os.Unsetenv("MY_VAR")
        }
    }()
    
    // Test
    // ... your test logic
}
```

## Writing New Tests

### Checklist untuk Test Baru

1. **Naming Convention**
   - File: `*_test.go`
   - Function: `Test<FunctionName>`
   - Helper: `test<Helper>` (lowercase, not exported)

2. **Test Structure**
   ```go
   func TestFunction(t *testing.T) {
       // Setup (if needed)
       
       // Test cases (table-driven if multiple scenarios)
       tests := []struct {
           name    string
           input   Type
           want    Type
           wantErr bool
       }{
           // ... test cases
       }
       
       for _, tt := range tests {
           t.Run(tt.name, func(t *testing.T) {
               // Act
               got, err := Function(tt.input)
               
               // Assert
               if (err != nil) != tt.wantErr {
                   t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
               }
               if got != tt.want {
                   t.Errorf("got = %v, want %v", got, tt.want)
               }
           })
       }
   }
   ```

3. **Test Coverage**
   - Happy path (normal case)
   - Edge cases (boundary values)
   - Error cases (invalid input)
   - Nil/empty checks

4. **Test Independence**
   - Each test should run independently
   - No shared state between tests
   - Use `t.TempDir()` untuk temporary files
   - Cleanup environment variables

5. **Clear Error Messages**
   ```go
   // Good
   t.Errorf("HasMeaningfulChanges() = %v, want %v (name changed)", got, want)
   
   // Bad
   t.Errorf("wrong result")
   ```

## Continuous Integration

### GitHub Actions Example

```yaml
name: Tests
on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      
      - name: Run tests
        run: go test -race -coverprofile=coverage.out ./internal/app/profile/...
      
      - name: Upload coverage
        uses: codecov/codecov-action@v3
        with:
          file: ./coverage.out
```

## Common Issues & Solutions

### Issue: Tests fail in CI but pass locally

**Cause:** Environment differences (timezone, env vars, file paths)

**Solution:**
- Use `t.TempDir()` instead of hardcoded paths
- Clear environment variables in test setup
- Don't depend on local filesystem structure

### Issue: Flaky tests (sometimes pass, sometimes fail)

**Cause:** Race conditions, timing issues, shared state

**Solution:**
- Run with `-race` flag to detect races
- Use `-count=10` to repeat tests
- Ensure test independence (no shared global state)

### Issue: Tests are slow

**Cause:** Integration tests mixing with unit tests, no test categorization

**Solution:**
- Separate unit and integration tests
- Use `-short` flag for unit tests only
- Run integration tests separately in CI

## Performance Benchmarking

Untuk future performance testing:

```go
func BenchmarkValidateProfileInfo(b *testing.B) {
    profile := &domain.ProfileInfo{
        Name: "test-db",
        DBInfo: domain.DBInfo{
            Host: "10.0.0.5",
            Port: 3306,
            User: "admin",
            Password: "secret",
        },
    }
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        ValidateProfileInfo(profile)
    }
}
```

Run benchmarks:
```bash
go test -bench=. ./internal/app/profile/validation/
go test -bench=BenchmarkValidate ./internal/app/profile/validation/
```

## Next Steps (Priority 3)

Future test additions:

1. **Parser Integration Tests**
   - LoadAndParseProfile with real files
   - Encryption/Decryption round-trip
   - INI parsing edge cases

2. **Connection Tests**
   - ValidateConnectPreflight
   - Error description logic
   - Mock network tests

3. **Executor Tests** (with mocks)
   - CreateProfile flow
   - SaveProfile logic
   - Path resolution

4. **Import Tests**
   - Row processing
   - Conflict resolution
   - Bulk operations

## Resources

- [Go Testing Package](https://pkg.go.dev/testing)
- [Table Driven Tests](https://go.dev/wiki/TableDrivenTests)
- [Testify Package](https://github.com/stretchr/testify) (optional, for assertions)

## Questions?

Refer to:
- Existing test files for examples
- `ARCHITECTURE.md` untuk design patterns
- Package `doc.go` files untuk behavior specifications
