// File : pkg/parsing/parsing_restore.go
// Deskripsi : Parsing functions untuk restore options
// Author : Hadiyatna Muflihun
// Tanggal : 16 Desember 2025
// Last Modified : 6 Januari 2026

package parsing

import (
	"path/filepath"
	backupfile "sfdbtools/internal/app/backup/helpers/file"
	restoremodel "sfdbtools/internal/app/restore/model"
	"sfdbtools/internal/cli/resolver"
	"sfdbtools/pkg/consts"
	"strings"

	"github.com/spf13/cobra"
)

// ParsingRestoreSingleOptions melakukan parsing opsi untuk restore single
func ParsingRestoreSingleOptions(cmd *cobra.Command) (restoremodel.RestoreSingleOptions, error) {
	opts := restoremodel.RestoreSingleOptions{
		DropTarget:  true,  // Default true
		SkipBackup:  false, // Default false
		StopOnError: true,  // Default: stop pada error pertama
	}

	// Profile & key (target)
	if err := PopulateTargetProfileFlags(cmd, &opts.Profile); err != nil {
		return restoremodel.RestoreSingleOptions{}, err
	}

	// Encryption key untuk decrypt backup file
	if err := PopulateRestoreEncryptionKey(cmd, &opts.EncryptionKey); err != nil {
		return restoremodel.RestoreSingleOptions{}, err
	}

	// Safety flags
	PopulateRestoreSafetyFlags(cmd, &opts.DropTarget, &opts.SkipBackup, &opts.DryRun, &opts.Force)
	PopulateStopOnErrorFromContinueFlag(cmd, &opts.StopOnError)

	// File backup
	if v := resolver.GetStringFlagOrEnv(cmd, "file", ""); v != "" {
		opts.File = v
	}

	// Ticket number
	PopulateRestoreTicket(cmd, &opts.Ticket)

	// Target database
	if v := resolver.GetStringFlagOrEnv(cmd, "target-db", ""); v != "" {
		opts.TargetDB = v
	}

	// Grants
	PopulateRestoreGrantsFlags(cmd, &opts.GrantsFile, &opts.SkipGrants)

	// Backup options untuk pre-restore backup
	opts.BackupOptions = &restoremodel.RestoreBackupOptions{}

	// Backup directory
	PopulateRestoreBackupDir(cmd, opts.BackupOptions)

	return opts, nil
}

// ParsingRestorePrimaryOptions melakukan parsing opsi untuk restore primary
func ParsingRestorePrimaryOptions(cmd *cobra.Command) (restoremodel.RestorePrimaryOptions, error) {
	opts := restoremodel.RestorePrimaryOptions{
		DropTarget:         true,  // Default true
		SkipBackup:         false, // Default false
		IncludeDmart:       true,  // Default true
		AutoDetectDmart:    true,  // Default true
		ConfirmIfNotExists: true,  // Default true
		StopOnError:        true,  // Default: stop pada error pertama
	}

	// Profile & key (target)
	if err := PopulateTargetProfileFlags(cmd, &opts.Profile); err != nil {
		return restoremodel.RestorePrimaryOptions{}, err
	}

	// Encryption key untuk decrypt backup file
	if err := PopulateRestoreEncryptionKey(cmd, &opts.EncryptionKey); err != nil {
		return restoremodel.RestorePrimaryOptions{}, err
	}

	// Safety flags (force handled khusus di bawah)
	PopulateRestoreSafetyFlags(cmd, &opts.DropTarget, &opts.SkipBackup, &opts.DryRun, &opts.Force)
	PopulateStopOnErrorFromContinueFlag(cmd, &opts.StopOnError)

	// File backup primary
	if v := resolver.GetStringFlagOrEnv(cmd, "file", ""); v != "" {
		opts.File = v
	}

	// Companion file (dmart)
	if v := resolver.GetStringFlagOrEnv(cmd, "dmart-file", ""); v != "" {
		opts.CompanionFile = v
	} else if v := resolver.GetStringFlagOrEnv(cmd, "companion-file", ""); v != "" {
		opts.CompanionFile = v
	}

	// Ticket number
	PopulateRestoreTicket(cmd, &opts.Ticket)

	// Target database
	// UX baru: --client-code (akan membentuk nama DB primary)
	if v := resolver.GetStringFlagOrEnv(cmd, "client-code", ""); v != "" {
		clientCode := strings.TrimSpace(v)
		clientCodeLower := strings.ToLower(clientCode)
		if strings.HasPrefix(clientCodeLower, consts.PrimaryPrefixNBC) || strings.HasPrefix(clientCodeLower, consts.PrimaryPrefixBiznet) {
			opts.TargetDB = clientCode
		} else {
			prefix := consts.PrimaryPrefixNBC
			inferred := backupfile.ExtractDatabaseNameFromFile(filepath.Base(opts.File))
			inferredLower := strings.ToLower(inferred)
			if strings.HasPrefix(inferredLower, consts.PrimaryPrefixBiznet) {
				prefix = consts.PrimaryPrefixBiznet
			} else if strings.HasPrefix(inferredLower, consts.PrimaryPrefixNBC) {
				prefix = consts.PrimaryPrefixNBC
			}
			opts.TargetDB = prefix + clientCode
		}
	}

	// Grants
	PopulateRestoreGrantsFlags(cmd, &opts.GrantsFile, &opts.SkipGrants)

	// Force mode special: jika force, maka skip konfirmasi.
	if opts.Force {
		opts.ConfirmIfNotExists = false
	}

	// Include dmart
	if cmd.Flags().Changed("dmart-include") {
		opts.IncludeDmart = resolver.GetBoolFlagOrEnv(cmd, "dmart-include", "")
	} else if cmd.Flags().Changed("include-dmart") {
		opts.IncludeDmart = resolver.GetBoolFlagOrEnv(cmd, "include-dmart", "")
	}

	// Auto-detect dmart
	if cmd.Flags().Changed("dmart-detect") {
		opts.AutoDetectDmart = resolver.GetBoolFlagOrEnv(cmd, "dmart-detect", "")
	} else if cmd.Flags().Changed("auto-detect-dmart") {
		opts.AutoDetectDmart = resolver.GetBoolFlagOrEnv(cmd, "auto-detect-dmart", "")
	}

	// Confirm if not exists
	if cmd.Flags().Changed("skip-confirm") {
		opts.ConfirmIfNotExists = !resolver.GetBoolFlagOrEnv(cmd, "skip-confirm", "")
	}

	// Backup options untuk pre-restore backup
	opts.BackupOptions = &restoremodel.RestoreBackupOptions{}

	// Backup directory
	PopulateRestoreBackupDir(cmd, opts.BackupOptions)

	return opts, nil
}

// ParsingRestoreAllOptions melakukan parsing opsi untuk restore all
func ParsingRestoreAllOptions(cmd *cobra.Command) (restoremodel.RestoreAllOptions, error) {
	opts := restoremodel.RestoreAllOptions{
		SkipBackup:    false,
		SkipSystemDBs: true, // Default true demi keamanan
		StopOnError:   true,
	}

	// 1. Basic configs (Profile, Keys, File)
	if err := PopulateTargetProfileFlags(cmd, &opts.Profile); err != nil {
		return restoremodel.RestoreAllOptions{}, err
	}
	if v := resolver.GetStringFlagOrEnv(cmd, "file", ""); v != "" {
		opts.File = v
	}
	if err := PopulateRestoreEncryptionKey(cmd, &opts.EncryptionKey); err != nil {
		return restoremodel.RestoreAllOptions{}, err
	}

	// 2. Safety Flags
	PopulateRestoreSafetyFlags(cmd, &opts.DropTarget, &opts.SkipBackup, &opts.DryRun, &opts.Force)
	PopulateStopOnErrorFromContinueFlag(cmd, &opts.StopOnError)

	// 4. Ticket & Backup Dir
	PopulateRestoreTicket(cmd, &opts.Ticket)
	opts.BackupOptions = &restoremodel.RestoreBackupOptions{}
	PopulateRestoreBackupDir(cmd, opts.BackupOptions)
	PopulateRestoreGrantsFlags(cmd, &opts.GrantsFile, &opts.SkipGrants)

	return opts, nil
}

// ParsingRestoreSecondaryOptions melakukan parsing opsi untuk restore secondary
func ParsingRestoreSecondaryOptions(cmd *cobra.Command) (restoremodel.RestoreSecondaryOptions, error) {
	opts := restoremodel.RestoreSecondaryOptions{
		DropTarget:      true,
		SkipBackup:      false,
		StopOnError:     true,
		From:            "",
		IncludeDmart:    true,
		AutoDetectDmart: true,
	}

	// Profile & key (target)
	if err := PopulateTargetProfileFlags(cmd, &opts.Profile); err != nil {
		return restoremodel.RestoreSecondaryOptions{}, err
	}

	// Encryption key (for decrypt source file / encrypt pre-backup primary)
	if err := PopulateRestoreEncryptionKey(cmd, &opts.EncryptionKey); err != nil {
		return restoremodel.RestoreSecondaryOptions{}, err
	}

	// Safety flags
	PopulateRestoreSafetyFlags(cmd, &opts.DropTarget, &opts.SkipBackup, &opts.DryRun, &opts.Force)
	PopulateStopOnErrorFromContinueFlag(cmd, &opts.StopOnError)

	// Source
	// Important UX: jika user tidak mengisi --from dan mode interaktif aktif,
	// pemilihan mode (file/primary) akan diprompt paling awal pada setup.
	if cmd.Flags().Changed("from") {
		if v := resolver.GetStringFlagOrEnv(cmd, "from", ""); v != "" {
			opts.From = strings.ToLower(strings.TrimSpace(v))
		}
	}

	// File (used when From=file)
	if v := resolver.GetStringFlagOrEnv(cmd, "file", ""); v != "" {
		opts.File = v
	}

	// Companion file (dmart)
	if v := resolver.GetStringFlagOrEnv(cmd, "dmart-file", ""); v != "" {
		opts.CompanionFile = v
	}
	if cmd.Flags().Changed("dmart-include") {
		opts.IncludeDmart = resolver.GetBoolFlagOrEnv(cmd, "dmart-include", "")
	}
	if cmd.Flags().Changed("dmart-detect") {
		opts.AutoDetectDmart = resolver.GetBoolFlagOrEnv(cmd, "dmart-detect", "")
	}

	// Ticket
	PopulateRestoreTicket(cmd, &opts.Ticket)

	// Secondary naming
	if v := resolver.GetStringFlagOrEnv(cmd, "client-code", ""); v != "" {
		opts.ClientCode = strings.TrimSpace(v)
	}
	if v := resolver.GetStringFlagOrEnv(cmd, "instance", ""); v != "" {
		opts.Instance = strings.TrimSpace(v)
	}

	// Backup options
	opts.BackupOptions = &restoremodel.RestoreBackupOptions{}
	PopulateRestoreBackupDir(cmd, opts.BackupOptions)

	return opts, nil
}
