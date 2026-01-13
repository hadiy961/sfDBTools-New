// File : internal/restore/setup_shared.go
// Deskripsi : Shared setup functions untuk restore operations
// Author : Hadiyatna Muflihun
// Tanggal : 30 Desember 2025
// Last Modified : 14 Januari 2026

package restore

import (
	backupfile "sfdbtools/internal/app/backup/helpers/file"
	"sfdbtools/internal/shared/fsops"
)

// resolveBackupFile resolve lokasi file backup
func (s *Service) resolveBackupFile(filePath *string, allowInteractive bool) error {
	return fsops.ResolveFileWithPrompt(fsops.FileResolverOptions{
		FilePath:         filePath,
		AllowInteractive: allowInteractive,
		ValidExtensions:  backupfile.ValidBackupFileExtensionsForSelection(),
		Purpose:          "file backup",
		PromptLabel:      "Masukkan path directory atau file backup",
		DefaultDir:       s.Config.Backup.Output.BaseDirectory,
	})
}

// resolveSelectionCSV resolve lokasi file CSV untuk restore selection
func (s *Service) resolveSelectionCSV(csvPath *string, allowInteractive bool) error {
	return fsops.ResolveFileWithPrompt(fsops.FileResolverOptions{
		FilePath:         csvPath,
		AllowInteractive: allowInteractive,
		ValidExtensions:  []string{".csv"},
		Purpose:          "file CSV",
		PromptLabel:      "Masukkan path CSV selection",
		DefaultDir:       s.Config.Backup.Output.BaseDirectory,
	})
}

// resolveEncryptionKey resolve encryption key untuk decrypt file
func (s *Service) resolveEncryptionKey(filePath string, encryptionKey *string, allowInteractive bool) error {
	if !backupfile.IsEncryptedFile(filePath) {
		s.Log.Debug("File backup tidak terenkripsi")
		return nil
	}

	return s.validateAndRetryEncryptionKey(filePath, encryptionKey, allowInteractive)
}
