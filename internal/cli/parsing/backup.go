package parsing

import (
	"fmt"
	"sfDBTools/internal/app/backup/model/types_backup"
	defaultVal "sfDBTools/internal/cli/defaults"
	"sfDBTools/pkg/compress"
	"sfDBTools/pkg/consts"
	"sfDBTools/pkg/helper"
	"sfDBTools/pkg/runtimecfg"
	"sfDBTools/pkg/validation"
	"strings"

	"github.com/spf13/cobra"
)

// ParsingBackupOptions melakukan parsing opsi untuk backup combined
func ParsingBackupOptions(cmd *cobra.Command, mode string) (types_backup.BackupDBOptions, error) {
	// Mulai dari default untuk mode combined
	opts := defaultVal.DefaultBackupOptions(mode)

	// Deteksi apakah ini command filter (untuk multi-select logic)
	// Command filter memiliki Use="filter", sedangkan all memiliki Use="all"
	isFilterCommand := cmd.Use == "filter"

	// Profile & key (Shared Helper)
	PopulateProfileFlags(cmd, &opts.Profile)

	// Filters (Shared Helper)
	PopulateFilterFlags(cmd, &opts.Filter)

	// Set flag untuk command filter agar bisa tampilkan multi-select jika tidak ada include/exclude
	// Ini digunakan di setup.go untuk menentukan apakah perlu multi-select atau tidak
	if isFilterCommand {
		opts.Filter.IsFilterCommand = true
	}

	// Encryption (Shared Helper)
	PopulateEncryptionFlags(cmd, &opts.Encryption)

	// CaptureGTID & ExcludeUser berasal dari config file (defaultval), tidak di-override via flag.

	// Dry Run
	opts.DryRun = helper.GetBoolFlagOrEnv(cmd, "dry-run", "")

	// Non Interactive (global): --quiet / --daemon
	opts.NonInteractive = runtimecfg.IsQuiet() || runtimecfg.IsDaemon()

	// Backup Directory
	if v := helper.GetStringFlagOrEnv(cmd, "backup-dir", ""); v != "" {
		opts.OutputDir = v
	}

	// Filename (optional; jika kosong akan auto dari config/pattern)
	if v := helper.GetStringFlagOrEnv(cmd, "filename", ""); v != "" {
		if err := validation.ValidateCustomFilenameBase(v); err != nil {
			return types_backup.BackupDBOptions{}, fmt.Errorf("filename tidak valid: %w", err)
		}
		opts.File.Filename = v
	}

	// Compression flags
	skipCompress := helper.GetBoolFlagOrEnv(cmd, "skip-compress", "")
	if skipCompress {
		opts.Compression.Enabled = false
		opts.Compression.Type = consts.CompressionTypeNone
	} else {
		// Jika user eksplisit set --skip-compress=false, aktifkan kompresi berdasarkan type (jika bukan none)
		if cmd.Flags().Changed("skip-compress") {
			if strings.TrimSpace(opts.Compression.Type) != "" && strings.ToLower(strings.TrimSpace(opts.Compression.Type)) != consts.CompressionTypeNone {
				opts.Compression.Enabled = true
			}
		}

		// compress type
		if v := helper.GetStringFlagOrEnv(cmd, "compress", ""); v != "" {
			ct, err := compress.ValidateCompressionType(v)
			if err != nil {
				return types_backup.BackupDBOptions{}, err
			}
			if string(ct) == consts.CompressionTypeNone {
				opts.Compression.Enabled = false
				opts.Compression.Type = consts.CompressionTypeNone
			} else {
				opts.Compression.Enabled = true
				opts.Compression.Type = string(ct)
			}
		}

		// compress level
		if cmd.Flags().Lookup("compress-level") != nil {
			lvl, err := cmd.Flags().GetInt("compress-level")
			if err != nil {
				return types_backup.BackupDBOptions{}, fmt.Errorf("gagal membaca compress-level: %w", err)
			}
			if opts.Compression.Enabled {
				if _, err := compress.ValidateCompressionLevel(lvl); err != nil {
					return types_backup.BackupDBOptions{}, err
				}
				opts.Compression.Level = lvl
			}
		}
	}

	// Encryption skip flag
	skipEncrypt := helper.GetBoolFlagOrEnv(cmd, "skip-encrypt", "")
	if skipEncrypt {
		opts.Encryption.Enabled = false
		opts.Encryption.Key = ""
	} else if cmd.Flags().Changed("skip-encrypt") {
		// Jika user eksplisit set --skip-encrypt=false, anggap enkripsi ingin dipakai.
		opts.Encryption.Enabled = true
	}

	// Mode-specific options
	if mode == "single" {
		if v := helper.GetStringFlagOrEnv(cmd, "database", ""); v != "" {
			opts.DBName = v
		}
		opts.IncludeDmart = helper.GetBoolFlagOrEnv(cmd, "include-dmart", "")
	} else if mode == "primary" {
		// Mode primary sama seperti single, hanya tanpa --database flag
		opts.IncludeDmart = helper.GetBoolFlagOrEnv(cmd, "include-dmart", "")
		if v := helper.GetStringFlagOrEnv(cmd, "client-code", ""); v != "" {
			opts.ClientCode = v
		}
	} else if mode == "secondary" {
		// Mode secondary sama seperti primary, hanya untuk database dengan suffix _secondary
		opts.IncludeDmart = helper.GetBoolFlagOrEnv(cmd, "include-dmart", "")
		if v := helper.GetStringFlagOrEnv(cmd, "client-code", ""); v != "" {
			opts.ClientCode = v
		}
		if v := helper.GetStringFlagOrEnv(cmd, "instance", ""); v != "" {
			opts.Instance = v
		}
	}

	// Ticket (wajib untuk semua mode)
	if v := helper.GetStringFlagOrEnv(cmd, "ticket", ""); v != "" {
		opts.Ticket = v
	}

	// Mode
	opts.Mode = mode

	// Validasi mode non-interaktif (fail-fast)
	if opts.NonInteractive {
		if strings.TrimSpace(opts.Ticket) == "" {
			return types_backup.BackupDBOptions{}, fmt.Errorf("ticket wajib diisi pada mode non-interaktif (--quiet): gunakan --ticket")
		}
		if strings.TrimSpace(opts.Profile.Path) == "" {
			return types_backup.BackupDBOptions{}, fmt.Errorf("profile wajib diisi pada mode non-interaktif (--quiet): gunakan --profile")
		}
		if strings.TrimSpace(opts.Profile.EncryptionKey) == "" {
			return types_backup.BackupDBOptions{}, fmt.Errorf("profile-key wajib diisi pada mode non-interaktif (--quiet): gunakan --profile-key atau env %s", consts.ENV_SOURCE_PROFILE_KEY)
		}
		if opts.Encryption.Enabled && strings.TrimSpace(opts.Encryption.Key) == "" {
			return types_backup.BackupDBOptions{}, fmt.Errorf("backup-key wajib diisi saat enkripsi aktif pada mode non-interaktif (--quiet): gunakan --backup-key atau env %s (atau set --skip-encrypt)", consts.ENV_BACKUP_ENCRYPTION_KEY)
		}

		// Mode-specific non-interactive requirements
		if mode == consts.ModeSingle && strings.TrimSpace(opts.DBName) == "" {
			return types_backup.BackupDBOptions{}, fmt.Errorf("database wajib diisi pada mode backup single saat non-interaktif: gunakan --database")
		}
		if isFilterCommand {
			hasInclude := len(opts.Filter.IncludeDatabases) > 0 || strings.TrimSpace(opts.Filter.IncludeFile) != ""
			if !hasInclude {
				return types_backup.BackupDBOptions{}, fmt.Errorf("mode backup filter non-interaktif membutuhkan include list: gunakan --db atau --db-file")
			}
		}
	}

	return opts, nil
}
