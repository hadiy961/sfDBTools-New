package cmddbscan

import (
	"errors"
	"fmt"
	"sfDBTools/internal/dbscan"
	"sfDBTools/internal/types"
	"sfDBTools/pkg/flags"
	"sfDBTools/pkg/parsing"

	"github.com/spf13/cobra"
)

// CmdDBScanFilter mengimplementasikan perintah `dbscan filter`
// untuk melakukan scan database dengan opsi include/exclude yang fleksibel.
var CmdDBScanFilter = &cobra.Command{
	Use:   "filter",
	Short: "Scan database dengan filter include/exclude",
	Long: `Command untuk melakukan scan pada database tertentu dengan opsi untuk mengecualikan database.

Contoh penggunaan:
  sfdbtools dbscan filter --exclude-system --exclude-db db1,db2,db3
  sfdbtools dbscan filter --exclude-system --exclude-db-file path/to/file.txt
  sfdbtools dbscan filter --exclude-system --exclude-db db1,db2,db3 --exclude-db-file path/to/file.txt
  sfdbtools dbscan filter --db db1,db2,db3
  sfdbtools dbscan filter --db-file path/to/file.txt

Catatan:
  - Jika kedua flag exclude-db/exclude-db-file dan db/db-file digunakan bersamaan, maka proses scan akan dilakukan pada database yang ada di flag db namun tidak ada di flag exclude-db.
  - Jika tidak ada flag exclude-db/exclude-db-file dan db/db-file yang digunakan (dan tidak ada fallback di konfigurasi), tampilkan error yang menjelaskan bahwa minimal salah satu flag harus digunakan.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if types.Deps == nil {
			return fmt.Errorf("dependencies tidak tersedia. Pastikan aplikasi diinisialisasi dengan benar")
		}

		logger := types.Deps.Logger
		cfg := types.Deps.Config

		// Parsing terpisah untuk merapikan command
		scanOpts, err := parsing.ParsingScanFilterOptions(cmd, cfg)
		if err != nil {
			return err
		}

		// Inisialisasi service dan jalankan
		svc := dbscan.NewDBScanService(logger, cfg)
		svc.SetScanOptions(scanOpts)

		scanConfig := types.ScanEntryConfig{
			HeaderTitle: "Database Scanning - Filter",
			ShowOptions: true,
			SuccessMsg:  "Proses scanning database (filter) selesai.",
			LogPrefix:   "Proses database scan (filter)",
			Mode:        "database",
		}

		if err := svc.ExecuteScanCommand(scanConfig); err != nil {
			if errors.Is(err, types.ErrUserCancelled) {
				logger.Warn("Proses dibatalkan oleh pengguna.")
				return nil
			}
			return err
		}
		return nil
	},
}

func init() {
	// Delegasikan pendaftaran flags ke paket flags agar konsisten dengan command lain
	// (mengikuti pattern seperti AddDbScanFlags)
	// Catatan: flags ini tidak di-bind ke struct langsung; parsing dilakukan di pkg/parsing
	flags.AddDbScanFilterFlags(CmdDBScanFilter)
}
