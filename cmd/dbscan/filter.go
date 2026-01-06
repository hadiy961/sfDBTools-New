package dbscancmd

import (
	"errors"
	"fmt"
	"sfdbtools/internal/app/dbscan"
	dbscanmodel "sfdbtools/internal/app/dbscan/model"
	appdeps "sfdbtools/internal/cli/deps"
	"sfdbtools/internal/cli/flags"
	"sfdbtools/internal/cli/parsing"
	"sfdbtools/pkg/validation"

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
  sfdbtools dbscan filter --exclude-system --exclude-file path/to/file.txt
  sfdbtools dbscan filter --exclude-system --exclude-db db1,db2 --exclude-file path/to/file.txt
  sfdbtools dbscan filter --db db1,db2,db3
  sfdbtools dbscan filter --db-file path/to/file.txt
  sfdbtools dbscan filter --db db1,db2 --db-file path/to/file.txt

Catatan:
  - Flag --exclude-db dan --exclude-file dapat dikombinasikan (hasil akan di-merge).
  - Flag --db dan --db-file dapat dikombinasikan (hasil akan di-merge).
  - Jika kedua include dan exclude digunakan, proses scan akan dilakukan pada database yang ada di include namun tidak ada di exclude.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if appdeps.Deps == nil {
			return fmt.Errorf("dependencies tidak tersedia. Pastikan aplikasi diinisialisasi dengan benar")
		}

		logger := appdeps.Deps.Logger
		cfg := appdeps.Deps.Config

		// Parsing terpisah untuk merapikan command
		scanOpts, err := parsing.ParsingScanFilterOptions(cmd, cfg)
		if err != nil {
			return err
		}

		// Resolve lists from files (baca file include/exclude dan merge)
		if err := dbscan.ResolveScanLists(&scanOpts); err != nil {
			return err
		}

		// Inisialisasi service dengan pattern baru
		svc := dbscan.NewDBScanService(cfg, logger, scanOpts)

		scanConfig := dbscanmodel.ScanEntryConfig{
			HeaderTitle: "Database Scanning - Filter",
			ShowOptions: true,
			SuccessMsg:  "Proses scanning database (filter) selesai.",
			LogPrefix:   "Proses database scan (filter)",
			Mode:        "database",
		}

		if err := dbscan.ExecuteScanCommand(svc, scanConfig); err != nil {
			if errors.Is(err, validation.ErrUserCancelled) {
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
