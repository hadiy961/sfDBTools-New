package verify

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sfdbtools/internal/app/backup/model/types_backup"
	"strings"
	"text/tabwriter"
)

// humanizeBytes formats bytes as human readable string
func humanizeBytes(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}

// DisplayResult menampilkan hasil verifikasi dalam format tabel atau JSON
func DisplayResult(result *types_backup.VerificationResult, filePath string, format string) {
	if format == "json" {
		out := map[string]interface{}{
			"file":   filePath,
			"result": result,
		}
		b, _ := json.MarshalIndent(out, "", "  ")
		fmt.Println(string(b))
		return
	}

	fmt.Println("📋 Hasil Verifikasi Backup")
	fmt.Println(strings.Repeat("-", 80))

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "Parameter\tNilai")
	fmt.Fprintln(w, "---------\t-----")

	fmt.Fprintf(w, "File\t%s\n", filepath.Base(filePath))

	statusSymbol := "❌"
	if result.VerifyStatus == "passed" {
		statusSymbol = "✓"
	} else if result.VerifyStatus == "partial" {
		statusSymbol = "⚠️"
	}
	fmt.Fprintf(w, "Status\t%s %s\n", statusSymbol, strings.ToUpper(result.VerifyStatus))

	if result.FailureReason != "" {
		fmt.Fprintf(w, "Reason\t%s\n", result.FailureReason)
	}

	if result.ChecksumAlgo != "" && result.ChecksumHash != "" {
		fmt.Fprintf(w, "Checksum (%s)\t%s\n", strings.ToUpper(result.ChecksumAlgo), result.ChecksumHash)
	}

	if result.SizeValid != nil {
		validStr := "valid"
		if !*result.SizeValid {
			validStr = "INVALID"
		}
		fmt.Fprintf(w, "File Size\t%s (%s)\n", humanizeBytes(result.FileSizeBytes), validStr)
	} else {
		fmt.Fprintf(w, "File Size\t%s\n", humanizeBytes(result.FileSizeBytes))
	}

	if result.HeaderValid != nil {
		if *result.HeaderValid {
			fmt.Fprintln(w, "SQL Header\t✓ Valid")
		} else {
			fmt.Fprintln(w, "SQL Header\t❌ Invalid")
		}
	} else {
		fmt.Fprintln(w, "SQL Header\tNot checked")
	}

	if result.FooterValid != nil {
		if *result.FooterValid {
			fmt.Fprintln(w, "SQL Footer\t✓ Valid")
		} else {
			fmt.Fprintln(w, "SQL Footer\t❌ Invalid")
		}
	} else {
		fmt.Fprintln(w, "SQL Footer\tNot checked")
	}

	if result.VerifiedAt != nil {
		fmt.Fprintf(w, "Verified At\t%s\n", result.VerifiedAt.Format("2006-01-02 15:04:05"))
	}

	w.Flush()
}

// DisplayBatchResults menampilkan hasil verifikasi batch (directory scan)
func DisplayBatchResults(results map[string]*types_backup.VerificationResult, format string) {
	if format == "json" {
		b, _ := json.MarshalIndent(results, "", "  ")
		fmt.Println(string(b))
		return
	}

	fmt.Println("📋 Batch Verification Results")
	fmt.Println(strings.Repeat("-", 80))

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "File\tStatus\tSize\tChecksum\tHeader/Footer")
	fmt.Fprintln(w, "----\t------\t----\t--------\t-------------")

	for file, res := range results {
		status := "❌ FAILED"
		if res.VerifyStatus == "passed" {
			status = "✓ PASSED"
		} else if res.VerifyStatus == "partial" {
			status = "⚠️ PARTIAL"
		}

		hfStatus := "N/A"
		if res.HeaderValid != nil && res.FooterValid != nil {
			if *res.HeaderValid && *res.FooterValid {
				hfStatus = "✓ Valid"
			} else {
				hfStatus = "❌ Invalid"
			}
		} else if res.HeaderValid != nil {
			if *res.HeaderValid {
				hfStatus = "Header OK"
			} else {
				hfStatus = "Header Fail"
			}
		}

		hash := res.ChecksumHash
		if len(hash) > 16 {
			hash = hash[:16] + "..."
		} else if hash == "" {
			hash = "N/A"
		}

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
			filepath.Base(file),
			status,
			humanizeBytes(res.FileSizeBytes),
			hash,
			hfStatus,
		)
	}

	w.Flush()
}
