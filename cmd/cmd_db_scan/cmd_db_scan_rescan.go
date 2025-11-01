package cmddbscan

import (
	"errors"
	"fmt"
	"sfDBTools/internal/dbscan"
	"sfDBTools/internal/types"
	defaultVal "sfDBTools/pkg/defaultval"
	"sfDBTools/pkg/flags"
	"sfDBTools/pkg/parsing"

	"github.com/spf13/cobra"
)

var scanRescanOpts types.ScanOptions

var CmdDBScanRescan = &cobra.Command{
	Use:   "rescan",
	Short: "Scan semua database dan collect informasi detail",
	Long: `Scan semua database dari server yang dikonfigurasi dan mengumpulkan informasi detail.

Hasil scanning dapat disimpan ke database_details dan database_detail_history table untuk tracking dan monitoring.

Contoh penggunaan:
	sfdbtools dbscan rescan --profile /path/to/profile.cnf.enc --profile-key SECRET
	sfdbtools dbscan rescan --profile /path/to/profile.cnf.enc --exclude-system=false
	sfdbtools dbscan rescan --profile /path/to/profile.cnf.enc --background
`,
	Run: func(cmd *cobra.Command, args []string) {
		// Pastikan dependencies tersedia
		if types.Deps == nil {
			fmt.Println("✗ Dependencies tidak tersedia. Pastikan aplikasi diinisialisasi dengan benar.")
			return
		}

		// Akses config dan logger dari dependency injection
		logger := types.Deps.Logger

		// Log dimulainya proses scan all database
		logger.Info("Memulai scanning semua database")

		// Parsing opsi memakai parser baru agar konsisten dengan command filter
		parsedOpts, err := parsing.ParsingScanAllOptions(cmd)
		if err != nil {
			logger.Error("gagal parsing opsi: " + err.Error())
			return
		}

		// Inisialisasi service dbscan
		svc := dbscan.NewDBScanService(logger, types.Deps.Config)
		svc.SetScanOptions(parsedOpts)

		// Execute scan
		scanConfig := types.ScanEntryConfig{
			HeaderTitle: "Database Scanning - Rescan Database (Yang gagal Sebelumnya)",
			ShowOptions: parsedOpts.ShowOptions,
			SuccessMsg:  "Proses scanning database selesai.",
			LogPrefix:   "Proses database scan",
			Mode:        "rescan",
		}

		if err := svc.ExecuteScanCommand(scanConfig); err != nil {
			if errors.Is(err, types.ErrUserCancelled) {
				logger.Warn("Proses dibatalkan oleh pengguna.")
				return
			}
			logger.Error("db-scan all gagal: " + err.Error())
		}
	},
}

func init() {
	// Set default values
	defaultOpts := defaultVal.GetDefaultScanOptions("all")
	scanAllOpts = defaultOpts

	// Tambahkan hanya flags yang diminta untuk 'all'
	flags.AddDbScanAllFlags(CmdDBScanRescan, &scanRescanOpts)
}
