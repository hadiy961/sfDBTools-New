// File : internal/app/dbcopy/helpers/flags.go
// Deskripsi : Helper functions untuk parsing command flags
// Author : Hadiyatna Muflihun
// Tanggal : 26 Januari 2026
// Last Modified : 26 Januari 2026
package helpers

import (
	"strings"

	"sfdbtools/internal/app/dbcopy/model"
	"sfdbtools/internal/cli/resolver"
	"sfdbtools/internal/shared/consts"

	"github.com/spf13/cobra"
)

// ParseCommonFlags mengekstrak flags yang common untuk semua mode
func ParseCommonFlags(cmd *cobra.Command) (*model.CommonCopyOptions, error) {
	opts := &model.CommonCopyOptions{}

	// Profile & Authentication
	opts.SourceProfile = strings.TrimSpace(resolver.GetStringFlagOrEnv(cmd, "source-profile", consts.ENV_SOURCE_PROFILE))
	profileKey, err := resolver.GetSecretStringFlagOrEnv(cmd, "source-profile-key", consts.ENV_SOURCE_PROFILE_KEY)
	if err != nil {
		return nil, err
	}
	opts.SourceProfileKey = strings.TrimSpace(profileKey)

	opts.TargetProfile = strings.TrimSpace(resolver.GetStringFlagOrEnv(cmd, "target-profile", consts.ENV_TARGET_PROFILE))
	targetKey, err := resolver.GetSecretStringFlagOrEnv(cmd, "target-profile-key", consts.ENV_TARGET_PROFILE_KEY)
	if err != nil {
		return nil, err
	}
	opts.TargetProfileKey = strings.TrimSpace(targetKey)

	// Audit & Control
	opts.Ticket = strings.TrimSpace(resolver.GetStringFlagOrEnv(cmd, "ticket", ""))

	// Behavior Flags
	opts.SkipConfirm = resolver.GetBoolFlagOrEnv(cmd, "skip-confirm", "")
	opts.ContinueOnError = resolver.GetBoolFlagOrEnv(cmd, "continue-on-error", "")
	opts.DryRun = resolver.GetBoolFlagOrEnv(cmd, "dry-run", "")
	opts.ExcludeData = resolver.GetBoolFlagOrEnv(cmd, "exclude-data", "")
	opts.IncludeDmart = resolver.GetBoolFlagOrEnv(cmd, "include-dmart", "")
	opts.PrebackupTarget = resolver.GetBoolFlagOrEnv(cmd, "prebackup-target", "")

	// Workdir
	opts.Workdir = strings.TrimSpace(resolver.GetStringFlagOrEnv(cmd, "workdir", ""))

	return opts, nil
}

// ParseP2PFlags parsing flags specific untuk P2P mode
func ParseP2PFlags(cmd *cobra.Command) (*model.P2POptions, error) {
	common, err := ParseCommonFlags(cmd)
	if err != nil {
		return nil, err
	}

	// P2P: pre-backup target WAJIB (safety). Flag tidak diekspos untuk p2p.
	common.PrebackupTarget = true

	opts := &model.P2POptions{
		CommonCopyOptions: *common,
	}

	opts.ClientCode = strings.TrimSpace(resolver.GetStringFlagOrEnv(cmd, "client-code", ""))
	opts.TargetClientCode = strings.TrimSpace(resolver.GetStringFlagOrEnv(cmd, "target-client-code", ""))
	opts.SourceDB = strings.TrimSpace(resolver.GetStringFlagOrEnv(cmd, "source-db", ""))
	opts.TargetDB = strings.TrimSpace(resolver.GetStringFlagOrEnv(cmd, "target-db", ""))

	return opts, nil
}

// ParseP2SFlags parsing flags specific untuk P2S mode
func ParseP2SFlags(cmd *cobra.Command) (*model.P2SOptions, error) {
	common, err := ParseCommonFlags(cmd)
	if err != nil {
		return nil, err
	}

	opts := &model.P2SOptions{
		CommonCopyOptions: *common,
	}

	opts.ClientCode = strings.TrimSpace(resolver.GetStringFlagOrEnv(cmd, "client-code", ""))
	opts.Instance = strings.TrimSpace(resolver.GetStringFlagOrEnv(cmd, "instance", ""))
	opts.SourceDB = strings.TrimSpace(resolver.GetStringFlagOrEnv(cmd, "source-db", ""))
	opts.TargetDB = strings.TrimSpace(resolver.GetStringFlagOrEnv(cmd, "target-db", ""))

	return opts, nil
}

// ParseS2SFlags parsing flags specific untuk S2S mode
func ParseS2SFlags(cmd *cobra.Command) (*model.S2SOptions, error) {
	common, err := ParseCommonFlags(cmd)
	if err != nil {
		return nil, err
	}

	opts := &model.S2SOptions{
		CommonCopyOptions: *common,
	}

	opts.ClientCode = strings.TrimSpace(resolver.GetStringFlagOrEnv(cmd, "client-code", ""))
	opts.SourceInstance = strings.TrimSpace(resolver.GetStringFlagOrEnv(cmd, "source-instance", ""))
	opts.TargetInstance = strings.TrimSpace(resolver.GetStringFlagOrEnv(cmd, "target-instance", ""))
	opts.SourceDB = strings.TrimSpace(resolver.GetStringFlagOrEnv(cmd, "source-db", ""))
	opts.TargetDB = strings.TrimSpace(resolver.GetStringFlagOrEnv(cmd, "target-db", ""))

	return opts, nil
}
