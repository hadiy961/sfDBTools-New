package verify

import (
	"os"
	"path/filepath"
	"sfdbtools/internal/app/backup/model/types_backup"
	applog "sfdbtools/internal/services/log"
	"testing"

	"github.com/stretchr/testify/assert"
)

type MockStep struct {
	NameFunc    func() string
	ExecuteFunc func(ctx *VerifyContext) error
}

func (m *MockStep) Name() string {
	if m.NameFunc != nil {
		return m.NameFunc()
	}
	return "MockStep"
}

func (m *MockStep) Execute(ctx *VerifyContext) error {
	if m.ExecuteFunc != nil {
		return m.ExecuteFunc(ctx)
	}
	return nil
}

func TestEngineCheck(t *testing.T) {
	// Create a temporary file to act as the backup file
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test_backup.sql")
	err := os.WriteFile(filePath, []byte("dummy data"), 0644)
	assert.NoError(t, err)

	opts := CheckOptions{
		SizeCheck:   true,
		MinFileSize: 1, // small enough
	}

	logger := applog.NullLogger()

	// Test passing pipeline
	step1 := &MockStep{
		ExecuteFunc: func(ctx *VerifyContext) error {
			// Set some properties to simulate success
			size := int64(10)
			ctx.Result.FileSizeBytes = size
			valid := true
			ctx.Result.SizeValid = &valid
			return nil
		},
	}

	engine := &Engine{steps: []VerificationStep{step1}}
	result, err := engine.Check(filePath, opts, logger)
	assert.NoError(t, err)
	assert.Equal(t, "passed", result.VerifyStatus)
	assert.Equal(t, int64(10), result.FileSizeBytes)

	// Test failing pipeline (soft failure)
	step2 := &MockStep{
		ExecuteFunc: func(ctx *VerifyContext) error {
			ctx.Result.VerifyStatus = "failed"
			ctx.Result.FailureReason = "soft error occurred"
			return nil
		},
	}
	engine2 := &Engine{steps: []VerificationStep{step2}}
	result2, err2 := engine2.Check(filePath, opts, logger)
	assert.NoError(t, err2)
	assert.Equal(t, "failed", result2.VerifyStatus)
	assert.Equal(t, "soft error occurred", result2.FailureReason)
}

func TestIsBackupFile(t *testing.T) {
	tests := []struct {
		filename string
		expected bool
	}{
		{"backup.sql", true},
		{"backup.sql.gz", true},
		{"backup.sql.zst", true},
		{"backup.sql.zst.enc", true},
		{"backup.zip", true},
		{"backup.txt", false},
		{"script.sh", false},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			assert.Equal(t, tt.expected, IsBackupFile(tt.filename))
		})
	}
}

func TestHasFailures(t *testing.T) {
	results := map[string]*types_backup.VerificationResult{
		"file1.sql": {VerifyStatus: "passed"},
		"file2.sql": {VerifyStatus: "passed"},
	}
	assert.False(t, HasFailures(results))

	results["file3.sql"] = &types_backup.VerificationResult{VerifyStatus: "failed"}
	assert.True(t, HasFailures(results))
}
