package verify

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sfdbtools/internal/app/backup/model/types_backup"
	applog "sfdbtools/internal/services/log"
	"strings"
	"time"
)

// CheckOptions mengontrol check mana yang dijalankan
type CheckOptions struct {
	Checksum      bool   // generate/compare checksum
	ChecksumAlgo  string // "sha256" atau "md5"
	HeaderFooter  bool   // validasi SQL header/footer (butuh streaming)
	SizeCheck     bool   // validasi minimum file size
	MinFileSize   int64  // minimum size dalam bytes
	ExpectedHash  string // jika non-empty, compare dengan hash ini
	EncryptionKey string // key untuk decrypt (jika header/footer check pada .enc file)
}

// Check menjalankan verifikasi pada satu backup file
func Check(filePath string, opts CheckOptions, logger applog.Logger) (*types_backup.VerificationResult, error) {
	result := &types_backup.VerificationResult{
		VerifyStatus: "passed",
	}

	now := time.Now()
	result.VerifiedAt = &now

	// 1. Size check
	if opts.SizeCheck {
		valid, size, err := ValidateSize(filePath, opts.MinFileSize)
		if err != nil {
			result.VerifyStatus = "failed"
			result.FailureReason = fmt.Sprintf("Failed to get file size: %v", err)
			return result, err
		}
		result.FileSizeBytes = size
		sizeValid := valid
		result.SizeValid = &sizeValid

		if !valid {
			result.VerifyStatus = "failed"
			result.FailureReason = fmt.Sprintf("File size %d is below minimum %d", size, opts.MinFileSize)
			return result, nil
		}
	} else {
		_, size, err := ValidateSize(filePath, 0)
		if err == nil {
			result.FileSizeBytes = size
		}
	}

	if result.FileSizeBytes == 0 {
		result.VerifyStatus = "failed"
		result.FailureReason = "File is empty"
		return result, nil
	}

	// 2. Checksum
	if opts.Checksum {
		algo := opts.ChecksumAlgo
		if algo == "" {
			algo = "sha256"
		}
		result.ChecksumAlgo = algo

		if opts.ExpectedHash != "" {
			match, actual, err := CompareChecksum(filePath, algo, opts.ExpectedHash)
			if err != nil {
				result.VerifyStatus = "failed"
				result.FailureReason = fmt.Sprintf("Failed to compare checksum: %v", err)
				return result, err
			}
			result.ChecksumHash = actual
			if !match {
				result.VerifyStatus = "failed"
				result.FailureReason = fmt.Sprintf("Checksum mismatch: expected %s, got %s", opts.ExpectedHash, actual)
				return result, nil
			}
		} else {
			hash, err := GenerateChecksum(filePath, algo)
			if err != nil {
				result.VerifyStatus = "failed"
				result.FailureReason = fmt.Sprintf("Failed to generate checksum: %v", err)
				return result, err
			}
			result.ChecksumHash = hash
		}
	}

	// 3. Header/Footer validation
	if opts.HeaderFooter {
		reader, closers, err := OpenVerifyReader(filePath, opts.EncryptionKey)
		if err != nil {
			result.VerifyStatus = "failed"
			result.FailureReason = fmt.Sprintf("Failed to open verify reader: %v", err)
			return result, err
		}
		defer CloseReaders(closers)

		headerOK, footerOK, err := ValidateHeaderFooter(reader)
		if err != nil {
			result.VerifyStatus = "failed"
			result.FailureReason = fmt.Sprintf("Error during header/footer validation: %v", err)
			return result, err
		}

		result.HeaderValid = &headerOK
		result.FooterValid = &footerOK

		if !headerOK || !footerOK {
			result.VerifyStatus = "failed"
			var reasons []string
			if !headerOK {
				reasons = append(reasons, "Invalid SQL Header")
			}
			if !footerOK {
				reasons = append(reasons, "Invalid SQL Footer (Dump not completed)")
			}
			result.FailureReason = strings.Join(reasons, ", ")
		}
	}

	return result, nil
}

// CheckFromMetadata menjalankan verifikasi menggunakan info dari .meta.json
func CheckFromMetadata(metaPath string, opts CheckOptions, logger applog.Logger) (*types_backup.VerificationResult, error) {
	data, err := os.ReadFile(metaPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read metadata file: %w", err)
	}

	var meta types_backup.BackupMetadata
	if err := json.Unmarshal(data, &meta); err != nil {
		return nil, fmt.Errorf("failed to parse metadata: %w", err)
	}

	// backup file path might be relative to metadata or absolute
	// generally it's absolute, but just in case:
	backupFilePath := meta.BackupFile
	if !filepath.IsAbs(backupFilePath) {
		backupFilePath = filepath.Join(filepath.Dir(metaPath), filepath.Base(meta.BackupFile))
	}

	if opts.ExpectedHash == "" && meta.Verification != nil && meta.Verification.ChecksumHash != "" {
		opts.ExpectedHash = meta.Verification.ChecksumHash
		if opts.ChecksumAlgo == "" {
			opts.ChecksumAlgo = meta.Verification.ChecksumAlgo
		}
	}

	return Check(backupFilePath, opts, logger)
}
