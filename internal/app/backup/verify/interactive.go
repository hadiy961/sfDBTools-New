package verify

import (
	"fmt"
	"os"

	backupfile "sfdbtools/internal/app/backup/helpers/file"
	"sfdbtools/internal/shared/fsops"

	"github.com/AlecAivazis/survey/v2"
)

// VerifyWizardOptions menyimpan hasil dari input wizard user
type VerifyWizardOptions struct {
	TargetPath string
	IsBatch    bool
	Opts       CheckOptions
	Format     string
}

// RunInteractiveVerify menjalankan wizard interaktif untuk memverifikasi backup
func RunInteractiveVerify() (*VerifyWizardOptions, error) {
	fmt.Println("=== Interactive Backup Verification ===")

	var targetType string
	err := survey.AskOne(&survey.Select{
		Message: "Pilih target verifikasi:",
		Options: []string{"Single File", "Directory (Batch)"},
	}, &targetType)
	if err != nil {
		return nil, fmt.Errorf("dibatalkan")
	}

	result := &VerifyWizardOptions{
		IsBatch: targetType == "Directory (Batch)",
		Opts:    CheckOptions{},
	}

	if !result.IsBatch {
		err = fsops.ResolveFileWithPrompt(fsops.FileResolverOptions{
			FilePath:         &result.TargetPath,
			AllowInteractive: true,
			ValidExtensions:  []string{".sql", ".sql.gz", ".sql.zst", ".enc", ".zip"},
			Purpose:          "file backup",
			PromptLabel:      "Pilih file backup untuk diverifikasi",
			DefaultDir:       ".",
		})
		if err != nil {
			return nil, fmt.Errorf("dibatalkan atau error: %w", err)
		}
	} else {
		err = survey.AskOne(&survey.Input{
			Message: "Masukkan path directory backup:",
			Default: ".",
		}, &result.TargetPath)
		if err != nil || result.TargetPath == "" {
			return nil, fmt.Errorf("dibatalkan")
		}

		fi, err := os.Stat(result.TargetPath)
		if err != nil || !fi.IsDir() {
			return nil, fmt.Errorf("directory tidak valid atau tidak ditemukan")
		}
	}

	var doHeaderFooter bool
	err = survey.AskOne(&survey.Confirm{
		Message: "Lakukan validasi struktur SQL (Header/Footer)?",
		Default: true,
	}, &doHeaderFooter)
	if err != nil {
		return nil, err
	}
	result.Opts.HeaderFooter = doHeaderFooter

	if doHeaderFooter {
		var hasEncryption bool

		if !result.IsBatch {
			hasEncryption = backupfile.IsEncryptedFile(result.TargetPath)
			if hasEncryption {
				fmt.Println("? [Auto-Detect] File terenkripsi (.enc)")
			}
		} else {
			err = survey.AskOne(&survey.Confirm{
				Message: "Apakah ada file backup yang dienkripsi (.enc) dalam direktori ini?",
				Default: false,
			}, &hasEncryption)
			if err != nil {
				return nil, err
			}
		}

		if hasEncryption {
			err = survey.AskOne(&survey.Password{
				Message: "Masukkan Encryption Key:",
			}, &result.Opts.EncryptionKey)
			if err != nil {
				return nil, err
			}
		}
	}

	var algo string
	err = survey.AskOne(&survey.Select{
		Message: "Pilih Algoritma Checksum:",
		Options: []string{"xxhash", "sha256", "md5", "skip"},
		Default: "sha256",
	}, &algo)
	if err != nil {
		return nil, err
	}

	if algo != "skip" {
		result.Opts.Checksum = true
		result.Opts.ChecksumAlgo = algo
	}

	if result.Opts.Checksum && !result.IsBatch {
		var hasExpectedHash bool
		survey.AskOne(&survey.Confirm{
			Message: "Apakah Anda memiliki expected hash untuk dibandingkan?",
			Default: false,
		}, &hasExpectedHash)

		if hasExpectedHash {
			survey.AskOne(&survey.Input{
				Message: "Masukkan Expected Hash:",
			}, &result.Opts.ExpectedHash)
		}
	}

	var doMinSize bool
	survey.AskOne(&survey.Confirm{
		Message: "Lakukan validasi ukuran file minimum?",
		Default: false,
	}, &doMinSize)

	if doMinSize {
		var sizeStr string
		err = survey.AskOne(&survey.Input{
			Message: "Masukkan ukuran minimum (dalam bytes):",
			Default: "1024",
		}, &sizeStr)
		if err == nil {
			fmt.Sscanf(sizeStr, "%d", &result.Opts.MinFileSize)
			if result.Opts.MinFileSize > 0 {
				result.Opts.SizeCheck = true
			}
		}
	}

	var format string
	survey.AskOne(&survey.Select{
		Message: "Pilih format output:",
		Options: []string{"table", "json"},
		Default: "table",
	}, &format)
	result.Format = format

	return result, nil
}
