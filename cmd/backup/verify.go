package backupcmd

import (
	"fmt"
	"os"
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
			// Jika tidak ada argumen dan tidak ada flag, jalankan mode interaktif
			wizResult, err := verify.RunInteractiveVerify()
			if err != nil {
				fmt.Println(err.Error())
				return
			}
			
			fmt.Println("\nMemulai verifikasi...")
			if wizResult.IsBatch {
				executeBatchVerify(wizResult.TargetPath, wizResult.Opts, wizResult.Format)
			} else {
				executeSingleVerify(wizResult.TargetPath, wizResult.Opts, wizResult.Format)
			}
			return
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
	CmdBackupVerify.Flags().StringVar(&verifyAlgo, "algo", "sha256", "Override checksum algorithm (sha256, md5, xxhash)")
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

func buildCheckOptionsFromFlags() verify.CheckOptions {
	return verify.CheckOptions{
		Checksum:      true,
		ChecksumAlgo:  verifyAlgo,
		HeaderFooter:  !verifyChecksumOnly,
		SizeCheck:     verifyMinSize > 0,
		MinFileSize:   verifyMinSize,
		ExpectedHash:  verifyExpectedHash,
		EncryptionKey: verifyKey,
	}
}

func runSingleVerify(targetFile string) {
	opts := buildCheckOptionsFromFlags()
	executeSingleVerify(targetFile, opts, verifyFormat)
}

func runBatchVerify(dirPath string) {
	opts := buildCheckOptionsFromFlags()
	executeBatchVerify(dirPath, opts, verifyFormat)
}

func executeSingleVerify(targetFile string, opts verify.CheckOptions, format string) {
	result, err := verify.Check(targetFile, opts, getLogger())
	if err != nil && result == nil {
		fmt.Printf("Error during verification: %v\n", err)
		os.Exit(1)
	}

	verify.DisplayResult(result, targetFile, format)

	if result.VerifyStatus == "failed" {
		os.Exit(1)
	}
}

func executeBatchVerify(dirPath string, opts verify.CheckOptions, format string) {
	results, err := verify.CheckBatch(dirPath, opts, getLogger())
	if err != nil {
		fmt.Printf("Error running batch verify: %v\n", err)
		os.Exit(1)
	}

	verify.DisplayBatchResults(results, format)

	if verify.HasFailures(results) {
		os.Exit(1)
	}
}
