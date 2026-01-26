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
)

// ValidateCommonOptions memvalidasi options yang wajib ada untuk semua mode
func ValidateCommonOptions(opts *model.CommonCopyOptions) error {
	if opts == nil {
		return fmt.Errorf("options tidak boleh nil")
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

	// Explicit mode: source-db dan target-db harus ada keduanya
	if opts.SourceDB != "" || opts.TargetDB != "" {
		if opts.SourceDB == "" || opts.TargetDB == "" {
			return fmt.Errorf("mode eksplisit butuh --source-db dan --target-db")
		}
		return nil
	}

	// Rule-based mode: client-code wajib
	if opts.ClientCode == "" {
		return fmt.Errorf("client-code wajib diisi pada mode rule-based: gunakan --client-code")
	}

	if opts.TargetClientCode == "" {
		opts.TargetClientCode = opts.ClientCode
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
