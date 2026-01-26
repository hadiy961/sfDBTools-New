// File : internal/app/dbcopy/parse.go
// Deskripsi : Parsing flags/env untuk db-copy
// Author : Hadiyatna Muflihun
// Tanggal : 26 Januari 2026
// Last Modified : 26 Januari 2026
package dbcopy

import (
	"fmt"
	"strings"

	"sfdbtools/internal/cli/resolver"
	"sfdbtools/internal/shared/consts"

	"github.com/spf13/cobra"
)

func parseCommon(cmd *cobra.Command) (CommonOptions, error) {
	var opts CommonOptions

	opts.SourceProfile = strings.TrimSpace(resolver.GetStringFlagOrEnv(cmd, "source-profile", consts.ENV_SOURCE_PROFILE))
	profileKey, err := resolver.GetSecretStringFlagOrEnv(cmd, "source-profile-key", consts.ENV_SOURCE_PROFILE_KEY)
	if err != nil {
		return CommonOptions{}, err
	}
	opts.SourceProfileKey = strings.TrimSpace(profileKey)

	opts.TargetProfile = strings.TrimSpace(resolver.GetStringFlagOrEnv(cmd, "target-profile", consts.ENV_TARGET_PROFILE))
	targetKey, err := resolver.GetSecretStringFlagOrEnv(cmd, "target-profile-key", consts.ENV_TARGET_PROFILE_KEY)
	if err != nil {
		return CommonOptions{}, err
	}
	opts.TargetProfileKey = strings.TrimSpace(targetKey)

	opts.Ticket = strings.TrimSpace(resolver.GetStringFlagOrEnv(cmd, "ticket", ""))

	opts.SkipConfirm = resolver.GetBoolFlagOrEnv(cmd, "skip-confirm", "")
	opts.ContinueOnError = resolver.GetBoolFlagOrEnv(cmd, "continue-on-error", "")
	opts.DryRun = resolver.GetBoolFlagOrEnv(cmd, "dry-run", "")
	opts.ExcludeData = resolver.GetBoolFlagOrEnv(cmd, "exclude-data", "")

	opts.IncludeDmart = resolver.GetBoolFlagOrEnv(cmd, "include-dmart", "")
	opts.PrebackupTarget = resolver.GetBoolFlagOrEnv(cmd, "prebackup-target", "")
	opts.Workdir = strings.TrimSpace(resolver.GetStringFlagOrEnv(cmd, "workdir", ""))

	// Validasi minimal untuk automation.
	if opts.SourceProfile == "" {
		return CommonOptions{}, fmt.Errorf("source-profile wajib diisi: gunakan --source-profile atau env %s", consts.ENV_SOURCE_PROFILE)
	}
	if opts.SourceProfileKey == "" {
		return CommonOptions{}, fmt.Errorf("source-profile-key wajib diisi: gunakan --source-profile-key atau env %s", consts.ENV_SOURCE_PROFILE_KEY)
	}
	if opts.TargetProfile != "" && opts.TargetProfileKey == "" {
		return CommonOptions{}, fmt.Errorf("target-profile-key wajib diisi jika --target-profile diisi: gunakan --target-profile-key atau env %s", consts.ENV_TARGET_PROFILE_KEY)
	}
	if opts.Ticket == "" {
		return CommonOptions{}, fmt.Errorf("ticket wajib diisi: gunakan --ticket")
	}

	return opts, nil
}

func parseP2S(cmd *cobra.Command) (P2SOptions, error) {
	common, err := parseCommon(cmd)
	if err != nil {
		return P2SOptions{}, err
	}
	opts := P2SOptions{Common: common}

	opts.ClientCode = strings.TrimSpace(resolver.GetStringFlagOrEnv(cmd, "client-code", ""))
	opts.Instance = strings.TrimSpace(resolver.GetStringFlagOrEnv(cmd, "instance", ""))

	opts.SourceDB = strings.TrimSpace(resolver.GetStringFlagOrEnv(cmd, "source-db", ""))
	opts.TargetDB = strings.TrimSpace(resolver.GetStringFlagOrEnv(cmd, "target-db", ""))

	// Rule-based atau explicit.
	if opts.SourceDB != "" || opts.TargetDB != "" {
		if opts.SourceDB == "" || opts.TargetDB == "" {
			return P2SOptions{}, fmt.Errorf("mode eksplisit butuh --source-db dan --target-db")
		}
		return opts, nil
	}

	if opts.ClientCode == "" {
		return P2SOptions{}, fmt.Errorf("client-code wajib diisi pada mode rule-based: gunakan --client-code")
	}
	if opts.Instance == "" {
		return P2SOptions{}, fmt.Errorf("instance wajib diisi pada mode rule-based: gunakan --instance")
	}

	return opts, nil
}

func parseP2P(cmd *cobra.Command) (P2POptions, error) {
	common, err := parseCommon(cmd)
	if err != nil {
		return P2POptions{}, err
	}
	opts := P2POptions{Common: common}

	opts.ClientCode = strings.TrimSpace(resolver.GetStringFlagOrEnv(cmd, "client-code", ""))
	opts.TargetClientCode = strings.TrimSpace(resolver.GetStringFlagOrEnv(cmd, "target-client-code", ""))
	opts.SourceDB = strings.TrimSpace(resolver.GetStringFlagOrEnv(cmd, "source-db", ""))
	opts.TargetDB = strings.TrimSpace(resolver.GetStringFlagOrEnv(cmd, "target-db", ""))

	if opts.SourceDB != "" || opts.TargetDB != "" {
		if opts.SourceDB == "" || opts.TargetDB == "" {
			return P2POptions{}, fmt.Errorf("mode eksplisit butuh --source-db dan --target-db")
		}
		return opts, nil
	}

	if opts.ClientCode == "" {
		return P2POptions{}, fmt.Errorf("client-code wajib diisi pada mode rule-based: gunakan --client-code")
	}
	if opts.TargetClientCode == "" {
		opts.TargetClientCode = opts.ClientCode
	}
	return opts, nil
}

func parseS2S(cmd *cobra.Command) (S2SOptions, error) {
	common, err := parseCommon(cmd)
	if err != nil {
		return S2SOptions{}, err
	}
	opts := S2SOptions{Common: common}

	opts.ClientCode = strings.TrimSpace(resolver.GetStringFlagOrEnv(cmd, "client-code", ""))
	opts.SourceInstance = strings.TrimSpace(resolver.GetStringFlagOrEnv(cmd, "source-instance", ""))
	opts.TargetInstance = strings.TrimSpace(resolver.GetStringFlagOrEnv(cmd, "target-instance", ""))
	opts.SourceDB = strings.TrimSpace(resolver.GetStringFlagOrEnv(cmd, "source-db", ""))
	opts.TargetDB = strings.TrimSpace(resolver.GetStringFlagOrEnv(cmd, "target-db", ""))

	if opts.SourceDB != "" || opts.TargetDB != "" {
		if opts.SourceDB == "" || opts.TargetDB == "" {
			return S2SOptions{}, fmt.Errorf("mode eksplisit butuh --source-db dan --target-db")
		}
		return opts, nil
	}

	if opts.ClientCode == "" {
		return S2SOptions{}, fmt.Errorf("client-code wajib diisi pada mode rule-based: gunakan --client-code")
	}
	if opts.SourceInstance == "" {
		return S2SOptions{}, fmt.Errorf("source-instance wajib diisi pada mode rule-based: gunakan --source-instance")
	}
	if opts.TargetInstance == "" {
		return S2SOptions{}, fmt.Errorf("target-instance wajib diisi pada mode rule-based: gunakan --target-instance")
	}

	return opts, nil
}
