package copy

import (
	"context"
	"fmt"
	"sfdbtools/internal/shared/database"
)

// VerifyChecksum membandingkan CRC32 checksum antara tabel sumber dan tabel tujuan.
func (s *Service) VerifyChecksum(ctx context.Context, client *database.Client, sourceDB, sourceTable, targetDB, targetTable string) (bool, error) {
	// Query checksum untuk source
	sourceSum, err := s.getChecksum(ctx, client, sourceDB, sourceTable)
	if err != nil {
		return false, fmt.Errorf("gagal ambil checksum sumber: %w", err)
	}

	// Query checksum untuk target
	targetSum, err := s.getChecksum(ctx, client, targetDB, targetTable)
	if err != nil {
		return false, fmt.Errorf("gagal ambil checksum tujuan: %w", err)
	}

	return sourceSum == targetSum, nil
}

func (s *Service) getChecksum(ctx context.Context, client *database.Client, dbName, tableName string) (int64, error) {
	query := fmt.Sprintf("CHECKSUM TABLE `%s`.`%s` ", dbName, tableName)
	row := client.DB().QueryRowContext(ctx, query)

	var name string
	var checksum int64
	if err := row.Scan(&name, &checksum); err != nil {
		// Beberapa engine mungkin tidak support, fallback ke count jika perlu
		// tapi kita ikuti spek CHECKSUM TABLE dulu.
		return 0, err
	}
	return checksum, nil
}
