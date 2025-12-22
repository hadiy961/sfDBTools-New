package backup

import (
	"context"
	"sfDBTools/internal/backup/metadata"
)

// ExportUserGrantsIfNeeded export user grants jika diperlukan
// Delegates to metadata.ExportUserGrantsIfNeededWithLogging dengan BackupDBOptions.ExcludeUser
func (s *Service) ExportUserGrantsIfNeeded(ctx context.Context, referenceBackupFile string, databases []string) string {
	// Skip export user grants jika dry-run mode
	if s.BackupDBOptions.DryRun {
		s.Log.Info("[DRY-RUN] Skip export user grants (dry-run mode aktif)")
		return ""
	}
	return metadata.ExportUserGrantsIfNeededWithLogging(ctx, s.Client, s.Log, referenceBackupFile, s.BackupDBOptions.ExcludeUser, databases)
}

// UpdateMetadataUserGrantsPath update metadata dengan actual user grants path
func (s *Service) UpdateMetadataUserGrantsPath(backupFilePath string, userGrantsPath string) {
	if err := metadata.UpdateMetadataUserGrantsFile(backupFilePath, userGrantsPath, s.Log); err != nil {
		s.Log.Warnf("Gagal update metadata user grants path: %v", err)
	}
}
