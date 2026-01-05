package ui

import (
	"fmt"
	"sfDBTools/internal/domain"
	applog "sfDBTools/internal/services/log"
	"sfDBTools/pkg/consts"
)

// DisplayFilterStats menampilkan statistik hasil pemfilteran database secara reusable.
// Context parameter untuk menyesuaikan label (contoh: "Akan di-backup" atau "Akan di-scan")
// Logger parameter opsional untuk logging (bisa nil)
func DisplayFilterStats(stats *domain.FilterStats, konteks string, logger applog.Logger) {
	PrintSubHeader("Statistik Filtering Database")

	// Hitung total excluded
	totalExcluded := stats.ExcludedSystem + stats.ExcludedByList + stats.ExcludedByFile + stats.ExcludedEmpty

	// Tentukan label action berdasarkan context
	actionLabel := "Akan diproses"
	if konteks == "backup" {
		actionLabel = "Akan di-backup"
	} else if konteks == "scan" {
		actionLabel = "Akan di-scan"
	}

	// Log statistik filtering
	if logger != nil {
		logger.WithFields(map[string]interface{}{
			"context":          konteks,
			"total_found":      stats.TotalFound,
			"total_included":   stats.TotalIncluded,
			"total_excluded":   totalExcluded,
			"excluded_system":  stats.ExcludedSystem,
			"excluded_by_list": stats.ExcludedByList,
			"excluded_by_file": stats.ExcludedByFile,
			"excluded_empty":   stats.ExcludedEmpty,
		}).Info("Filter statistics")
	}

	data := [][]string{
		{"Total Ditemukan", fmt.Sprintf("%d", stats.TotalFound)},
		{actionLabel, ColorText(fmt.Sprintf("%d", stats.TotalIncluded), consts.UIColorGreen)},
		{"Total Dikecualikan", ColorText(fmt.Sprintf("%d", totalExcluded), consts.UIColorYellow)},
	}

	// Tampilkan detail exclusion jika ada yang dikecualikan
	if totalExcluded > 0 {
		data = append(data, []string{"", ""}) // Empty row for separation
		data = append(data, []string{ColorText("Detail Exclusion:", consts.UIColorCyan), ""})

		if stats.ExcludedSystem > 0 {
			data = append(data, []string{"  - Sistem Database", fmt.Sprintf("%d", stats.ExcludedSystem)})
		}
		if stats.ExcludedByList > 0 {
			data = append(data, []string{"  - Exclude List", fmt.Sprintf("%d", stats.ExcludedByList)})
		}
		if stats.ExcludedByFile > 0 {
			data = append(data, []string{"  - Tidak di Include List", fmt.Sprintf("%d", stats.ExcludedByFile)})
		}
		if stats.ExcludedEmpty > 0 {
			data = append(data, []string{"  - Nama Kosong", fmt.Sprintf("%d", stats.ExcludedEmpty)})
		}
	}

	FormatTable([]string{"Kategori", "Jumlah"}, data)

	// Tampilkan warning untuk database yang tidak ditemukan di server
	hasWarnings := len(stats.NotFoundInInclude) > 0 || len(stats.NotFoundInExclude) > 0 || len(stats.NotFoundInWhitelist) > 0 || len(stats.NotFoundInBlacklist) > 0

	if hasWarnings {
		PrintWarning("\n⚠️  Peringatan: Ada database yang tidak ditemukan di server")

		// Log warning tentang database tidak ditemukan
		if logger != nil {
			logger.WithFields(map[string]interface{}{
				"not_found_in_include":   stats.NotFoundInInclude,
				"not_found_in_exclude":   stats.NotFoundInExclude,
				"not_found_in_whitelist": stats.NotFoundInWhitelist,
				"not_found_in_blacklist": stats.NotFoundInBlacklist,
			}).Warn("Databases not found on server")
		}

		if len(stats.NotFoundInInclude) > 0 {
			PrintWarning(fmt.Sprintf("\nDatabase di Include List yang tidak ditemukan (%d):", len(stats.NotFoundInInclude)))
			for _, db := range stats.NotFoundInInclude {
				PrintWarning(fmt.Sprintf("  - %s", db))
			}
		}

		if len(stats.NotFoundInExclude) > 0 {
			PrintWarning(fmt.Sprintf("\nDatabase di Exclude List yang tidak ditemukan (%d):", len(stats.NotFoundInExclude)))
			for _, db := range stats.NotFoundInExclude {
				PrintWarning(fmt.Sprintf("  - %s", db))
			}
		}

		if len(stats.NotFoundInWhitelist) > 0 {
			PrintWarning(fmt.Sprintf("\nDatabase di Whitelist File yang tidak ditemukan (%d):", len(stats.NotFoundInWhitelist)))
			for _, db := range stats.NotFoundInWhitelist {
				PrintWarning(fmt.Sprintf("  - %s", db))
			}
		}

		if len(stats.NotFoundInBlacklist) > 0 {
			PrintWarning(fmt.Sprintf("\nDatabase di Exclude File yang tidak ditemukan (%d):", len(stats.NotFoundInBlacklist)))
			for _, db := range stats.NotFoundInBlacklist {
				PrintWarning(fmt.Sprintf("  - %s", db))
			}
		}

		fmt.Println() // Empty line for spacing
	}
}
