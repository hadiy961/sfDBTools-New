package helper

import (
	"sfDBTools/pkg/compress"
	"strings"
	"testing"
)

func TestGenerateBackupFilename(t *testing.T) {
	tests := []struct {
		name            string
		pattern         string
		database        string
		mode            string
		hostname        string
		compressionType compress.CompressionType
		encrypted       bool
		contains        []string // Substring yang harus ada
	}{
		{
			name:            "Combined mode with default pattern",
			pattern:         "{database}_{year}{month}{day}_{hour}{minute}{second}_{hostname}.sql",
			database:        "",
			mode:            "combined",
			hostname:        "db-server-01",
			compressionType: compress.CompressionNone,
			encrypted:       false,
			contains:        []string{"all_databases_", "db-server-01", ".sql"},
		},
		{
			name:            "Separate mode with gzip compression",
			pattern:         "{database}_{timestamp}.sql",
			database:        "mydb",
			mode:            "separate",
			hostname:        "mysql-prod",
			compressionType: compress.CompressionGzip,
			encrypted:       false,
			contains:        []string{"mydb_", ".sql.gz"},
		},
		{
			name:            "With encryption only",
			pattern:         "{database}_{timestamp}.sql",
			database:        "testdb",
			mode:            "separate",
			hostname:        "dbhost",
			compressionType: compress.CompressionNone,
			encrypted:       true,
			contains:        []string{"testdb_", ".sql.enc"},
		},
		{
			name:            "With compression and encryption",
			pattern:         "{database}_{year}{month}{day}.sql",
			database:        "proddb",
			mode:            "separate",
			hostname:        "prod-mysql-01",
			compressionType: compress.CompressionZstd,
			encrypted:       true,
			contains:        []string{"proddb_", ".sql.zst.enc"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := GenerateBackupFilename(tt.pattern, tt.database, tt.mode, tt.hostname, tt.compressionType, tt.encrypted)
			if err != nil {
				t.Errorf("GenerateBackupFilename() error = %v", err)
				return
			}

			for _, substr := range tt.contains {
				if !strings.Contains(result, substr) {
					t.Errorf("GenerateBackupFilename() = %v, want to contain %v", result, substr)
				}
			}
		})
	}
}

func TestGenerateBackupDirectory(t *testing.T) {
	tests := []struct {
		name             string
		baseDir          string
		structurePattern string
		hostname         string
		contains         []string
	}{
		{
			name:             "With year/month/day pattern",
			baseDir:          "/mnt/nfs/backup",
			structurePattern: "{year}/{month}/{day}/",
			hostname:         "db-server-01",
			contains:         []string{"/mnt/nfs/backup/", "/"},
		},
		{
			name:             "With hostname in path",
			baseDir:          "/backups",
			structurePattern: "{hostname}/{year}/{month}/",
			hostname:         "mysql-prod",
			contains:         []string{"/backups", "mysql-prod"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := GenerateBackupDirectory(tt.baseDir, tt.structurePattern, tt.hostname)
			if err != nil {
				t.Errorf("GenerateBackupDirectory() error = %v", err)
				return
			}

			for _, substr := range tt.contains {
				if !strings.Contains(result, substr) {
					t.Errorf("GenerateBackupDirectory() = %v, want to contain %v", result, substr)
				}
			}
		})
	}
}

func TestGenerateFullBackupPath(t *testing.T) {
	baseDir := "/mnt/nfs/backup"
	structurePattern := "{year}/{month}/{day}/"
	filenamePattern := "{database}_{timestamp}_{hostname}.sql"
	database := "mydb"
	mode := "separate"
	hostname := "db-server-01"
	compressionType := compress.CompressionGzip
	encrypted := true

	result, err := GenerateFullBackupPath(baseDir, structurePattern, filenamePattern, database, mode, hostname, compressionType, encrypted)
	if err != nil {
		t.Errorf("GenerateFullBackupPath() error = %v", err)
		return
	}

	// Check if result contains expected parts
	if !strings.Contains(result, baseDir) {
		t.Errorf("GenerateFullBackupPath() = %v, want to contain %v", result, baseDir)
	}

	if !strings.Contains(result, "mydb_") {
		t.Errorf("GenerateFullBackupPath() = %v, want to contain mydb_", result)
	}

	if !strings.Contains(result, "db-server-01") {
		t.Errorf("GenerateFullBackupPath() = %v, want to contain db-server-01", result)
	}

	if !strings.HasSuffix(result, ".sql.gz.enc") {
		t.Errorf("GenerateFullBackupPath() = %v, want to end with .sql.gz.enc", result)
	}
}
