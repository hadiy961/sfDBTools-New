package grants

import (
	"context"

	"sfdbtools/internal/app/backup/metadata"
	applog "sfdbtools/internal/services/log"
	"sfdbtools/pkg/database"
)

// ExportUserGrantsIfNeeded exports user grants unless excluded or in dry-run.
func ExportUserGrantsIfNeeded(ctx context.Context, client *database.Client, log applog.Logger, referenceBackupFile string, excludeUser bool, dryRun bool, databases []string) string {
	if dryRun {
		log.Info("[DRY-RUN] Skip export user grants (dry-run mode aktif)")
		return ""
	}
	return metadata.ExportUserGrantsIfNeededWithLogging(ctx, client, log, referenceBackupFile, excludeUser, databases)
}

// UpdateMetadataUserGrantsPath updates backup metadata with the actual user grants file path.
func UpdateMetadataUserGrantsPath(log applog.Logger, backupFilePath string, userGrantsPath string) {
	if err := metadata.UpdateMetadataUserGrantsFile(backupFilePath, userGrantsPath, log); err != nil {
		log.Warnf("Gagal update metadata user grants path: %v", err)
	}
}
