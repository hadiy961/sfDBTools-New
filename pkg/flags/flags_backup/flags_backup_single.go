package flagsbackup

import (
	"sfDBTools/internal/types/types_backup"

	"github.com/spf13/cobra"
)

func SingleBackupFlags(cmd *cobra.Command, defaultOpts *types_backup.BackupDBOptions) {
	// Menambahkan flag spesifik untuk backup single database
	cmd.Flags().StringVarP(&defaultOpts.OutputDir, "output-dir", "o", defaultOpts.OutputDir, "Direktori output untuk menyimpan file backup")
	cmd.Flags().StringVarP(&defaultOpts.File.Filename, "filename", "f", "", "Nama file backup")
	cmd.Flags().StringVarP(&defaultOpts.DBName, "database", "d", defaultOpts.DBName, "Nama database yang akan di-backup")
	cmd.Flags().BoolVarP(&defaultOpts.ExcludeUser, "exclude-user", "e", defaultOpts.ExcludeUser, "Exclude user grants dari export")
	cmd.Flags().BoolVar(&defaultOpts.Filter.ExcludeData, "exclude-data", defaultOpts.Filter.ExcludeData, "Backup hanya struktur database tanpa data")
	cmd.Flags().StringVar(&defaultOpts.Ticket, "ticket", defaultOpts.Ticket, "Ticket number untuk request backup")
}
