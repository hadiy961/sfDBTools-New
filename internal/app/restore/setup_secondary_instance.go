// File : internal/restore/setup_secondary_instance.go
// Deskripsi : Helper untuk penentuan instance pada restore secondary
// Author : Hadiyatna Muflihun
// Tanggal : 30 Desember 2025
// Last Modified : 14 Januari 2026
package restore

import (
	"context"
	"fmt"
	backupfile "sfdbtools/internal/app/backup/helpers/file"
	restoremodel "sfdbtools/internal/app/restore/model"
	"sfdbtools/internal/shared/naming"
	"sfdbtools/internal/ui/prompt"
	"strings"
)

func (s *Service) resolveSecondaryInstance(ctx context.Context, opts *restoremodel.RestoreSecondaryOptions, primaryDB string, allowInteractive bool) error {
	inst := strings.TrimSpace(opts.Instance)
	if inst != "" {
		return applyValidatedSecondaryInstance(opts, primaryDB, inst)
	}

	if !allowInteractive {
		return fmt.Errorf("instance wajib diisi (--instance) pada mode non-interaktif (--force)")
	}

	instances, err := s.fetchExistingSecondaryInstances(ctx, primaryDB)
	if err != nil {
		return err
	}

	chosen, err := s.pickSecondaryInstance(instances, primaryDB)
	if err != nil {
		return err
	}

	opts.Instance = chosen
	return nil
}

func (s *Service) fetchExistingSecondaryInstances(ctx context.Context, primaryDB string) ([]string, error) {
	databases, err := s.TargetClient.GetNonSystemDatabases(ctx)
	if err != nil {
		return nil, fmt.Errorf("gagal mendapatkan list database: %w", err)
	}

	secondaryPrefix := primaryDB + "_secondary_"
	return extractSecondaryInstances(databases, secondaryPrefix), nil
}

func (s *Service) pickSecondaryInstance(instances []string, primaryDB string) (string, error) {
	options := buildSecondaryInstanceOptions(instances)

	selected, _, err := prompt.SelectOne("Pilih instance secondary", options, 0)
	if err != nil {
		return "", fmt.Errorf("gagal memilih instance: %w", err)
	}

	switch selected {
	case manualInstanceOption():
		return askSecondaryInstanceManual(primaryDB)
	case separatorOption():
		return "", fmt.Errorf("pilihan tidak valid")
	default:
		return strings.TrimSpace(selected), nil
	}
}

func askSecondaryInstanceManual(primaryDB string) (string, error) {
	val, err := prompt.AskText(
		"Masukkan instance secondary: ",
		prompt.WithDefault("training"),
		prompt.WithValidator(func(ans interface{}) error { return validateSecondaryInstanceInput(primaryDB, ans) }),
	)
	if err != nil {
		return "", fmt.Errorf("gagal mendapatkan instance: %w", err)
	}
	trimmed := strings.TrimSpace(val)
	if err := validateSecondaryInstanceName(primaryDB, trimmed); err != nil {
		return "", err
	}
	return trimmed, nil
}

func applyValidatedSecondaryInstance(opts *restoremodel.RestoreSecondaryOptions, primaryDB, inst string) error {
	if err := validateSecondaryInstanceName(primaryDB, inst); err != nil {
		return err
	}
	opts.Instance = strings.TrimSpace(inst)
	return nil
}

func validateSecondaryInstanceInput(primaryDB string, ans interface{}) error {
	v, ok := ans.(string)
	if !ok {
		return fmt.Errorf("input tidak valid")
	}
	return validateSecondaryInstanceName(primaryDB, v)
}

func validateSecondaryInstanceName(primaryDB, inst string) error {
	inst = strings.TrimSpace(inst)
	if inst == "" {
		return fmt.Errorf("instance tidak boleh kosong")
	}
	target := naming.BuildSecondaryDBName(primaryDB, inst)
	if !backupfile.IsValidDatabaseName(target) {
		return fmt.Errorf("instance menghasilkan nama database tidak valid: %s", target)
	}
	return nil
}

func buildSecondaryInstanceOptions(instances []string) []string {
	options := []string{manualInstanceOption()}
	if len(instances) > 0 {
		options = append(options, separatorOption())
		options = append(options, instances...)
	}
	return options
}

func manualInstanceOption() string {
	return "⌨️  [ Input instance baru secara manual ]"
}

func separatorOption() string {
	return "────────────────────────"
}
