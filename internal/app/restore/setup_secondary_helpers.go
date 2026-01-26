// File : internal/restore/setup_secondary_helpers.go
// Deskripsi : Helper functions untuk setup restore secondary operations
// Author : Hadiyatna Muflihun
// Tanggal : 30 Desember 2025
// Last Modified : 26 Januari 2026
package restore

import (
	"context"
	"fmt"
	restoremodel "sfdbtools/internal/app/restore/model"
	"sfdbtools/internal/shared/consts"
	"sfdbtools/internal/shared/naming"
	"sfdbtools/internal/ui/prompt"
	"strings"
)

// resolveSecondaryFrom resolve dari source (file atau primary)
func (s *Service) resolveSecondaryFrom(opts *restoremodel.RestoreSecondaryOptions, allowInteractive bool) error {
	from := normalizeRestoreSource(opts.From)
	if from == "" {
		if allowInteractive {
			selected, _, err := prompt.SelectOne("Pilih mode restore secondary (source)", validSecondarySources(), 0)
			if err != nil {
				return fmt.Errorf("gagal memilih mode restore secondary: %w", err)
			}
			from = selected
		} else {
			from = "file"
		}
	}

	if !isValidSecondarySource(from) {
		if !allowInteractive {
			return fmt.Errorf("nilai --from tidak valid: %s (gunakan: primary atau file)", from)
		}
		selected, _, err := prompt.SelectOne("Pilih sumber restore secondary", validSecondarySources(), 0)
		if err != nil {
			return fmt.Errorf("gagal memilih --from: %w", err)
		}
		from = selected
	}

	opts.From = from
	return nil
}

// resolveSecondaryClientCode resolve client code untuk secondary restore
func (s *Service) resolveSecondaryClientCode(opts *restoremodel.RestoreSecondaryOptions, allowInteractive bool) error {
	if strings.TrimSpace(opts.ClientCode) != "" {
		return nil
	}
	if !allowInteractive {
		return fmt.Errorf("client code wajib diisi (--client-code) pada mode non-interaktif (--skip-confirm/--quiet)")
	}
	cc, err := prompt.AskText("Masukkan client code: ", prompt.WithValidator(validateClientCodeInput))
	if err != nil {
		return fmt.Errorf("gagal mendapatkan client code: %w", err)
	}
	opts.ClientCode = strings.TrimSpace(cc)
	return nil
}

// resolveSecondaryEncryptionKey resolve encryption key untuk secondary restore
func (s *Service) resolveSecondaryEncryptionKey(opts *restoremodel.RestoreSecondaryOptions, allowInteractive bool) error {
	if opts.From == "file" {
		return s.resolveEncryptionKey(opts.File, &opts.EncryptionKey, allowInteractive)
	}

	if !s.Config.Backup.Encryption.Enabled {
		return nil
	}
	if strings.TrimSpace(opts.EncryptionKey) != "" {
		return nil
	}
	if !allowInteractive {
		return fmt.Errorf("encryption key wajib diisi (--encryption-key) karena backup encryption aktif")
	}

	k, err := prompt.AskText("Masukkan encryption key untuk file backup (SFDB_BACKUP_ENCRYPTION_KEY): ", prompt.WithValidator(validateClientCodeInput))
	if err != nil {
		return fmt.Errorf("gagal mendapatkan encryption key: %w", err)
	}
	opts.EncryptionKey = strings.TrimSpace(k)
	return nil
}

// resolveSecondaryPrimaryDB resolve primary database untuk mode "from=primary"
func (s *Service) resolveSecondaryPrimaryDB(ctx context.Context, opts *restoremodel.RestoreSecondaryOptions) error {
	cc := strings.TrimSpace(opts.ClientCode)
	if cc == "" {
		return fmt.Errorf("client code kosong")
	}

	ccLower := strings.ToLower(cc)
	if strings.HasPrefix(ccLower, consts.PrimaryPrefixNBC) || strings.HasPrefix(ccLower, consts.PrimaryPrefixBiznet) {
		opts.PrimaryDB = cc
		return nil
	}

	nbc := naming.BuildPrimaryDBName(consts.PrimaryPrefixNBC, cc)
	biz := naming.BuildPrimaryDBName(consts.PrimaryPrefixBiznet, cc)

	nbcExists, nbcErr := s.TargetClient.CheckDatabaseExists(ctx, nbc)
	if nbcErr != nil {
		return fmt.Errorf("gagal mengecek database primary (NBC): %w", nbcErr)
	}
	bizExists, bizErr := s.TargetClient.CheckDatabaseExists(ctx, biz)
	if bizErr != nil {
		return fmt.Errorf("gagal mengecek database primary (Biznet): %w", bizErr)
	}

	if nbcExists {
		opts.PrimaryDB = nbc
		if bizExists {
			s.Log.Warnf("Ditemukan 2 primary (%s dan %s); menggunakan %s", nbc, biz, nbc)
		}
		return nil
	}
	if bizExists {
		opts.PrimaryDB = biz
		return nil
	}

	return fmt.Errorf("database primary tidak ditemukan untuk client-code %q (coba: %s atau %s)", cc, nbc, biz)
}

// resolveSecondaryPrefixForFileMode menentukan prefix primary untuk mode "from=file"
func (s *Service) resolveSecondaryPrefixForFileMode(ctx context.Context, opts *restoremodel.RestoreSecondaryOptions) (string, error) {
	cc := strings.TrimSpace(opts.ClientCode)
	if cc == "" {
		return consts.PrimaryPrefixNBC, nil
	}

	dbs, err := s.TargetClient.GetNonSystemDatabases(ctx)
	if err == nil {
		needleNBC := consts.PrimaryPrefixNBC + cc + "_secondary_"
		needleBiz := consts.PrimaryPrefixBiznet + cc + "_secondary_"
		for _, db := range dbs {
			if strings.HasPrefix(db, needleNBC) {
				return consts.PrimaryPrefixNBC, nil
			}
		}
		for _, db := range dbs {
			if strings.HasPrefix(db, needleBiz) {
				return consts.PrimaryPrefixBiznet, nil
			}
		}
	}

	if strings.TrimSpace(opts.File) != "" {
		return naming.InferPrimaryPrefix("", opts.File), nil
	}

	return consts.PrimaryPrefixNBC, nil
}

// resolveSecondaryInstance dipindahkan ke setup_secondary_instance.go

func normalizeRestoreSource(val string) string {
	return strings.ToLower(strings.TrimSpace(val))
}

func validSecondarySources() []string {
	return []string{"file", "primary"}
}

func isValidSecondarySource(val string) bool {
	return val == "file" || val == "primary"
}
