// File : internal/app/dbcopy/helpers/validation.go
// Deskripsi : Helper functions untuk validasi copy operations
// Author : Hadiyatna Muflihun
// Tanggal : 26 Januari 2026
// Last Modified : 26 Januari 2026
package helpers

import (
	"fmt"
	"strings"

	"sfdbtools/internal/app/dbcopy/model"
	"sfdbtools/internal/shared/consts"
)

// ValidateCommonOptions memvalidasi options yang wajib ada untuk semua mode
func ValidateCommonOptions(opts *model.CommonCopyOptions) error {
	if opts == nil {
		return fmt.Errorf("options tidak boleh nil")
	}

	if strings.TrimSpace(opts.SourceProfile) == "" {
		return fmt.Errorf("source-profile wajib diisi: gunakan --source-profile atau env %s", consts.ENV_SOURCE_PROFILE)
	}
	if strings.TrimSpace(opts.SourceProfileKey) == "" {
		return fmt.Errorf("source-profile-key wajib diisi: gunakan --source-profile-key atau env %s", consts.ENV_SOURCE_PROFILE_KEY)
	}
	if strings.TrimSpace(opts.TargetProfile) != "" && strings.TrimSpace(opts.TargetProfileKey) == "" {
		return fmt.Errorf("target-profile-key wajib diisi jika --target-profile diisi: gunakan --target-profile-key atau env %s", consts.ENV_TARGET_PROFILE_KEY)
	}

	if opts.Ticket == "" {
		return fmt.Errorf("ticket wajib diisi: gunakan --ticket")
	}

	return nil
}

// ValidateP2POptions memvalidasi opsi spesifik untuk P2P mode
func ValidateP2POptions(opts *model.P2POptions) error {
	if err := ValidateCommonOptions(&opts.CommonCopyOptions); err != nil {
		return err
	}

	// P2P wajib punya target profile (tidak boleh default sama dengan source).
	if strings.TrimSpace(opts.TargetProfile) == "" {
		return fmt.Errorf("target-profile wajib diisi untuk db-copy p2p: gunakan --target-profile atau env SFDB_TARGET_PROFILE")
	}

	// Mode eksplisit: cukup --source-db, target-db HARUS sama dengan source-db.
	if strings.TrimSpace(opts.SourceDB) != "" {
		if strings.TrimSpace(opts.TargetDB) != "" && !strings.EqualFold(strings.TrimSpace(opts.TargetDB), strings.TrimSpace(opts.SourceDB)) {
			return fmt.Errorf("untuk db-copy p2p, --target-db harus sama dengan --source-db (target akan selalu = source)")
		}
		return nil
	}

	// Rule-based: client-code wajib (target-client-code tidak digunakan di p2p).
	if strings.TrimSpace(opts.ClientCode) == "" {
		return fmt.Errorf("untuk db-copy p2p (non-interaktif), butuh --source-db atau --client-code")
	}
	if strings.TrimSpace(opts.TargetClientCode) != "" && !strings.EqualFold(strings.TrimSpace(opts.TargetClientCode), strings.TrimSpace(opts.ClientCode)) {
		return fmt.Errorf("untuk db-copy p2p, --target-client-code tidak didukung (target database selalu sama dengan source)")
	}

	return nil
}

// ValidateP2SOptions memvalidasi opsi spesifik untuk P2S mode
func ValidateP2SOptions(opts *model.P2SOptions) error {
	if err := ValidateCommonOptions(&opts.CommonCopyOptions); err != nil {
		return err
	}

	// Explicit mode
	if opts.SourceDB != "" || opts.TargetDB != "" {
		if opts.SourceDB == "" || opts.TargetDB == "" {
			return fmt.Errorf("mode eksplisit butuh --source-db dan --target-db")
		}
		return nil
	}

	// Rule-based mode
	if opts.ClientCode == "" {
		return fmt.Errorf("client-code wajib diisi pada mode rule-based: gunakan --client-code")
	}
	if opts.Instance == "" {
		return fmt.Errorf("instance wajib diisi pada mode rule-based: gunakan --instance")
	}

	return nil
}

// ValidateS2SOptions memvalidasi opsi spesifik untuk S2S mode
func ValidateS2SOptions(opts *model.S2SOptions) error {
	if err := ValidateCommonOptions(&opts.CommonCopyOptions); err != nil {
		return err
	}

	// Explicit mode
	if opts.SourceDB != "" || opts.TargetDB != "" {
		if opts.SourceDB == "" || opts.TargetDB == "" {
			return fmt.Errorf("mode eksplisit butuh --source-db dan --target-db")
		}
		return nil
	}

	// Rule-based mode
	if opts.ClientCode == "" {
		return fmt.Errorf("client-code wajib diisi pada mode rule-based: gunakan --client-code")
	}
	if opts.SourceInstance == "" {
		return fmt.Errorf("source-instance wajib diisi pada mode rule-based: gunakan --source-instance")
	}
	if opts.TargetInstance == "" {
		return fmt.Errorf("target-instance wajib diisi pada mode rule-based: gunakan --target-instance")
	}

	return nil
}

// IsExplicitMode mengembalikan true jika menggunakan mode eksplisit (source-db + target-db)
func IsExplicitMode(sourceDB, targetDB string) bool {
	return strings.TrimSpace(sourceDB) != "" || strings.TrimSpace(targetDB) != ""
}
