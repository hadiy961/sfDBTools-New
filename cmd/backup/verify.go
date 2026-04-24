package backupcmd

import (
	"fmt"
	"os"
	"path/filepath"
	"sfdbtools/internal/app/backup/model/types_backup"
	"sfdbtools/internal/app/backup/verify"
	applog "sfdbtools/internal/services/log"

	"github.com/spf13/cobra"
)

var (
	verifyDir          string
	verifyLatest       bool
	verifyProfile      string
	verifyChecksumOnly bool
	verifyExpectedHash string
	verifyFormat       string
	verifyAlgo         string
	verifyKey          string
	verifyMinSize      int64
)

// CmdBackupVerify adalah command untuk verifikasi integritas backup
var CmdBackupVerify = &cobra.Command{
	Use:   "verify [file]",
	Short: "Verifikasi integritas file backup",
	Long: `Melakukan verifikasi file backup untuk memastikan integritas data.
Mendukung verifikasi ukuran file, checksum, dan validasi struktur SQL (header/footer).`,
	Run: func(cmd *cobra.Command, args []string) {
		if verifyDir != "" {
			runBatchVerify(verifyDir)
			return
		}

		if len(args) == 0 && !verifyLatest {
			fmt.Println("Error: Harus menspesifikasikan file backup atau menggunakan --dir / --latest")
			cmd.Help()
			os.Exit(1)
		}

		var targetFile string
		if verifyLatest {
			if verifyProfile == "" {
				fmt.Println("Error: --profile wajib diisi jika menggunakan --latest")
				os.Exit(1)
			}
			targetFile = getLatestBackup(verifyProfile)
		} else {
			targetFile = args[0]
		}

		if targetFile == "" {
			fmt.Println("Error: File backup tidak ditemukan")
			os.Exit(1)
		}

		runSingleVerify(targetFile)
	},
}

func init() {
	CmdBackupVerify.Flags().StringVar(&verifyDir, "dir", "", "Verify semua file di directory")
	CmdBackupVerify.Flags().BoolVar(&verifyLatest, "latest", false, "Verify backup terbaru")
	CmdBackupVerify.Flags().StringVar(&verifyProfile, "profile", "", "Profile untuk --latest lookup")
	CmdBackupVerify.Flags().BoolVar(&verifyChecksumOnly, "checksum-only", false, "Hanya generate checksum (tanpa header/footer)")
	CmdBackupVerify.Flags().StringVar(&verifyExpectedHash, "expected-hash", "", "Expected hash untuk comparison")
	CmdBackupVerify.Flags().StringVar(&verifyFormat, "format", "table", "Output format: table atau json")
	CmdBackupVerify.Flags().StringVar(&verifyAlgo, "algo", "sha256", "Override checksum algorithm (sha256, md5)")
	CmdBackupVerify.Flags().StringVar(&verifyKey, "encryption-key", "", "Kunci dekripsi jika file dienkripsi (dibutuhkan untuk header/footer)")
	CmdBackupVerify.Flags().Int64Var(&verifyMinSize, "min-size", 0, "Minimum file size in bytes")
}

func getLatestBackup(profileName string) string {
	fmt.Println("Notice: --latest feature is a stub in this version.")
	return ""
}

func getLogger() applog.Logger {
	return applog.NullLogger()
}

func runSingleVerify(targetFile string) {
	opts := verify.CheckOptions{
		Checksum:      true,
		ChecksumAlgo:  verifyAlgo,
		HeaderFooter:  !verifyChecksumOnly,
		SizeCheck:     verifyMinSize > 0,
		MinFileSize:   verifyMinSize,
		ExpectedHash:  verifyExpectedHash,
		EncryptionKey: verifyKey,
	}

	result, err := verify.Check(targetFile, opts, getLogger())
	if err != nil && result == nil {
		fmt.Printf("Error during verification: %v\n", err)
		os.Exit(1)
	}
	
	verify.DisplayResult(result, targetFile, verifyFormat)
	
	if result.VerifyStatus == "failed" {
		os.Exit(1)
	}
}

func runBatchVerify(dirPath string) {
	opts := verify.CheckOptions{
		Checksum:      true,
		ChecksumAlgo:  verifyAlgo,
		HeaderFooter:  !verifyChecksumOnly,
		SizeCheck:     verifyMinSize > 0,
		MinFileSize:   verifyMinSize,
		EncryptionKey: verifyKey,
	}

	files, err := os.ReadDir(dirPath)
	if err != nil {
		fmt.Printf("Error reading directory: %v\n", err)
		os.Exit(1)
	}

	results := make(map[string]*types_backup.VerificationResult)
	hasFailed := false

	logger := getLogger()

	for _, f := range files {
		if f.IsDir() {
			continue
		}
		
		ext := filepath.Ext(f.Name())
		if ext == ".sql" || ext == ".gz" || ext == ".zst" || ext == ".enc" || ext == ".zip" {
			filePath := filepath.Join(dirPath, f.Name())
			res, err := verify.Check(filePath, opts, logger)
			if err != nil && res == nil {
				fmt.Printf("Failed to verify %s: %v\n", filePath, err)
				hasFailed = true
				continue
			}
			results[filePath] = res
			
			if res.VerifyStatus == "failed" {
				hasFailed = true
			}
		}
	}

	verify.DisplayBatchResults(results, verifyFormat)
	
	if hasFailed {
		os.Exit(1)
	}
}
