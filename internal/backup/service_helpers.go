package backup

import (
	"sfDBTools/internal/types/types_backup"
	"sfDBTools/pkg/compress"
)

// buildCompressionSettings membuat CompressionSettings dari BackupDBOptions
func (s *Service) buildCompressionSettings() types_backup.CompressionSettings {
	compressionType := s.BackupDBOptions.Compression.Type
	if !s.BackupDBOptions.Compression.Enabled {
		compressionType = ""
	}
	return types_backup.CompressionSettings{
		Type:    compress.CompressionType(compressionType),
		Enabled: s.BackupDBOptions.Compression.Enabled,
		Level:   s.BackupDBOptions.Compression.Level,
	}
}
