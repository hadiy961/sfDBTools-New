package backupimpl

import (
	"sfDBTools/internal/types/types_backup"

	"github.com/spf13/cobra"
)

func SecondaryBackupFlags(cmd *cobra.Command, defaultOpts *types_backup.BackupDBOptions) {
	// Menambahkan flag spesifik untuk backup secondary database (tanpa --database flag)
	cmd.Flags().StringVarP(&defaultOpts.OutputDir, "output-dir", "o", defaultOpts.OutputDir, "Direktori output untuk menyimpan file backup")
	cmd.Flags().StringVarP(&defaultOpts.File.Filename, "filename", "f", "", "Nama file backup")
	cmd.Flags().BoolVarP(&defaultOpts.ExcludeUser, "exclude-user", "e", defaultOpts.ExcludeUser, "Exclude user grants dari export")
	cmd.Flags().BoolVar(&defaultOpts.Filter.ExcludeData, "exclude-data", defaultOpts.Filter.ExcludeData, "Backup hanya struktur database tanpa data")
	cmd.Flags().BoolVar(&defaultOpts.IncludeDmart, "include-dmart", defaultOpts.IncludeDmart, "Backup juga database <database>_dmart jika tersedia")
	cmd.Flags().BoolVar(&defaultOpts.IncludeTemp, "include-temp", defaultOpts.IncludeTemp, "Backup juga database <database>_temp jika tersedia")
	cmd.Flags().BoolVar(&defaultOpts.IncludeArchive, "include-archive", defaultOpts.IncludeArchive, "Backup juga database <database>_archive jika tersedia")
	cmd.Flags().StringVar(&defaultOpts.ClientCode, "client-code", defaultOpts.ClientCode, "Filter database secondary berdasarkan client code (contoh: adaro)")
	cmd.Flags().StringVar(&defaultOpts.Instance, "instance", defaultOpts.Instance, "Filter database secondary berdasarkan instance (contoh: 1, 2, 3)")
	cmd.Flags().StringVar(&defaultOpts.Ticket, "ticket", defaultOpts.Ticket, "Ticket number untuk request backup")
}
