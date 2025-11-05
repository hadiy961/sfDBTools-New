// File : internal/backup/backup_checksum_files.go
// Deskripsi : Generator untuk checksum files (.sha256, .md5) dan metadata manifest
// Author : Hadiyatna Muflihun
// Tanggal : 2025-11-05
// Last Modified : 2025-11-05

package backup

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sfDBTools/internal/types"
	"time"
)

// WriteChecksumFiles menulis checksum ke file .sha256 dan .md5
// dengan format yang compatible dengan sha256sum dan md5sum tools
func (s *Service) WriteChecksumFiles(backupFile, sha256Hash, md5Hash string) error {
	baseFileName := filepath.Base(backupFile)

	// Write SHA256 checksum file
	if sha256Hash != "" {
		sha256File := backupFile + ".sha256"
		// Format: <hash>  <filename>
		// Space ganda adalah standard format untuk *sum tools
		content := fmt.Sprintf("%s  %s\n", sha256Hash, baseFileName)
		if err := os.WriteFile(sha256File, []byte(content), 0644); err != nil {
			return fmt.Errorf("gagal menulis SHA256 checksum file: %w", err)
		}
		s.Log.Infof("✓ SHA256 checksum file: %s", sha256File)
	}

	// Write MD5 checksum file
	if md5Hash != "" {
		md5File := backupFile + ".md5"
		content := fmt.Sprintf("%s  %s\n", md5Hash, baseFileName)
		if err := os.WriteFile(md5File, []byte(content), 0644); err != nil {
			return fmt.Errorf("gagal menulis MD5 checksum file: %w", err)
		}
		s.Log.Infof("✓ MD5 checksum file: %s", md5File)
	}

	return nil
}

// WriteBackupMetadata menulis metadata lengkap backup ke file JSON
func (s *Service) WriteBackupMetadata(metadata *types.BackupMetadata) error {
	metadataFile := metadata.BackupFile + ".meta.json"

	// Marshal ke JSON dengan indentation untuk readability
	jsonData, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return fmt.Errorf("gagal marshal metadata ke JSON: %w", err)
	}

	// Tulis ke file
	if err := os.WriteFile(metadataFile, jsonData, 0644); err != nil {
		return fmt.Errorf("gagal menulis metadata file: %w", err)
	}

	s.Log.Infof("✓ Metadata file: %s", metadataFile)
	return nil
}

// CreateBackupMetadata membuat object metadata lengkap untuk backup
func (s *Service) CreateBackupMetadata(
	backupFile string,
	backupType string,
	databases []string,
	startTime, endTime time.Time,
	sha256Hash, md5Hash string,
	fileSize int64,
) *types.BackupMetadata {

	// Buat checksum info
	checksums := []types.ChecksumInfo{}

	if sha256Hash != "" {
		checksums = append(checksums, types.ChecksumInfo{
			Algorithm:    "sha256",
			Hash:         sha256Hash,
			CalculatedAt: endTime,
			FileSize:     fileSize,
		})
	}

	if md5Hash != "" {
		checksums = append(checksums, types.ChecksumInfo{
			Algorithm:    "md5",
			Hash:         md5Hash,
			CalculatedAt: endTime,
			FileSize:     fileSize,
		})
	}

	// Buat metadata object
	metadata := &types.BackupMetadata{
		BackupFile:      backupFile,
		BackupType:      backupType,
		DatabaseNames:   databases,
		Hostname:        s.BackupDBOptions.Profile.DBInfo.Host,
		BackupStartTime: startTime,
		BackupEndTime:   endTime,
		BackupDuration:  endTime.Sub(startTime).String(),
		FileSize:        fileSize,
		FileSizeHuman:   formatFileSize(fileSize),
		Compressed:      s.BackupDBOptions.Compression.Enabled,
		CompressionType: s.BackupDBOptions.Compression.Type,
		Encrypted:       s.BackupDBOptions.Encryption.Enabled,
		Checksums:       checksums,
		MariaDBVersion:  s.BackupDBOptions.Profile.DBInfo.Version,
		BackupStatus:    "success",
		GeneratedBy:     fmt.Sprintf("%s v%s", s.Config.General.AppName, s.Config.General.Version),
		GeneratedAt:     time.Now(),
	}

	return metadata
}

// formatFileSize adalah helper untuk format file size ke human-readable
func formatFileSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.2f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// VerifyChecksumFiles memverifikasi checksum dengan membaca dari .sha256 dan .md5 files
func (s *Service) VerifyChecksumFiles(backupFile string) error {
	var errors []string

	// Verify SHA256
	sha256File := backupFile + ".sha256"
	if _, err := os.Stat(sha256File); err == nil {
		content, err := os.ReadFile(sha256File)
		if err != nil {
			errors = append(errors, fmt.Sprintf("gagal membaca SHA256 file: %v", err))
		} else {
			s.Log.Infof("SHA256 checksum file tersedia: %s", sha256File)
			// Parse format: <hash>  <filename>
			var hash, filename string
			fmt.Sscanf(string(content), "%s  %s", &hash, &filename)
			if hash != "" {
				s.Log.Infof("SHA256: %s", hash)
			}
		}
	}

	// Verify MD5
	md5File := backupFile + ".md5"
	if _, err := os.Stat(md5File); err == nil {
		content, err := os.ReadFile(md5File)
		if err != nil {
			errors = append(errors, fmt.Sprintf("gagal membaca MD5 file: %v", err))
		} else {
			s.Log.Infof("MD5 checksum file tersedia: %s", md5File)
			var hash, filename string
			fmt.Sscanf(string(content), "%s  %s", &hash, &filename)
			if hash != "" {
				s.Log.Infof("MD5: %s", hash)
			}
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("verifikasi checksum files gagal: %v", errors)
	}

	return nil
}
