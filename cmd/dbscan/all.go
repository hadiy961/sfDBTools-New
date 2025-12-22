package dbscancmd

import (
	"errors"
	"fmt"
	"sfDBTools/internal/dbscan"
	appdeps "sfDBTools/internal/deps"
	"sfDBTools/internal/flags"
	"sfDBTools/internal/parsing"
	"sfDBTools/internal/types"
	"sfDBTools/pkg/validation"

	"github.com/spf13/cobra"
)

var CmdDBScanAll = &cobra.Command{
	Use:   "all",
	Short: "Scan semua database dan collect informasi detail",
	Long: `Scan semua database dari server yang dikonfigurasi dan mengumpulkan informasi detail.

Hasil scanning dapat disimpan ke database_details dan database_detail_history table untuk tracking dan monitoring.

Contoh penggunaan:
	sfdbtools dbscan all --profile /path/to/profile.cnf.enc --profile-key SECRET
	sfdbtools dbscan all --profile /path/to/profile.cnf.enc --exclude-system=false
	sfdbtools dbscan all --profile /path/to/profile.cnf.enc --background
`,
	Run: func(cmd *cobra.Command, args []string) {
		// Pastikan dependencies tersedia
		if appdeps.Deps == nil {
			fmt.Println("✗ Dependencies tidak tersedia. Pastikan aplikasi diinisialisasi dengan benar.")
			return
		}

		// Akses config dan logger dari dependency injection
		logger := appdeps.Deps.Logger

		// Log dimulainya proses scan all database
		logger.Info("Memulai scanning semua database")

		// Parsing opsi memakai parser baru agar konsisten dengan command filter
		parsedOpts, err := parsing.ParsingScanAllOptions(cmd)
		if err != nil {
			logger.Error("gagal parsing opsi: " + err.Error())
			return
		}

		// Inisialisasi service dbscan dengan pattern baru
		svc := dbscan.NewDBScanService(appdeps.Deps.Config, logger, parsedOpts)

		// Execute scan
		scanConfig := types.ScanEntryConfig{
			HeaderTitle: "Database Scanning - Semua Database",
			ShowOptions: parsedOpts.ShowOptions,
			SuccessMsg:  "Proses scanning database selesai.",
			LogPrefix:   "Proses database scan",
			Mode:        "all",
		}

		if err := dbscan.ExecuteScanCommand(svc, scanConfig); err != nil {
			if errors.Is(err, validation.ErrUserCancelled) {
				logger.Warn("Proses dibatalkan oleh pengguna.")
				return
			}
			logger.Error("db-scan all gagal: " + err.Error())
		}
	},
}

func init() {
	// Tambahkan hanya flags yang diminta untuk 'all'
	flags.AddDbScanAllFlags(CmdDBScanAll)
}
