package copy

import (
	"context"
	"fmt"

	profileconn "sfdbtools/internal/app/profile/connection"
	"sfdbtools/internal/domain"
	"sfdbtools/internal/ui/prompt"
)

// SelectDatabasesInteractive memunculkan picker multi-select untuk memilih database.
func (s *Service) SelectDatabasesInteractive(ctx context.Context, profile *domain.ProfileInfo) ([]string, error) {
	client, err := profileconn.ConnectWithProfile(s.cfg, profile, "")
	if err != nil {
		return nil, err
	}
	defer client.Close()

	dbs, err := client.GetNonSystemDatabases(ctx)
	if err != nil {
		return nil, err
	}

	if len(dbs) == 0 {
		return nil, fmt.Errorf("tidak ada database yang tersedia di server")
	}

	selected, _, err := prompt.SelectMany("Pilih database sumber (Spasi untuk memilih):", dbs, nil)
	return selected, err
}

// SelectTablesInteractive memunculkan picker untuk memilih database dan tabel (multi-select).
func (s *Service) SelectTablesInteractive(ctx context.Context, profile *domain.ProfileInfo) (dbName string, tableNames []string, err error) {
	client, err := profileconn.ConnectWithProfile(s.cfg, profile, "")
	if err != nil {
		return "", nil, err
	}
	defer client.Close()

	// 1. Pilih DB
	dbs, err := client.GetNonSystemDatabases(ctx)
	if err != nil {
		return "", nil, err
	}
	dbName, _, err = prompt.SelectOne("Pilih database sumber:", dbs, 0)
	if err != nil {
		return "", nil, err
	}

	// 2. Ambil list tabel
	query := fmt.Sprintf("SHOW TABLES FROM `%s` ", dbName)
	rows, err := client.QueryContextWithRetry(ctx, query)
	if err != nil {
		return "", nil, err
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var t string
		if err := rows.Scan(&t); err == nil {
			tables = append(tables, t)
		}
	}

	if len(tables) == 0 {
		return "", nil, fmt.Errorf("tidak ada tabel di database '%s'", dbName)
	}

	// 3. Pilih Tabel
	tableNames, _, err = prompt.SelectMany(fmt.Sprintf("Pilih tabel di %s (Spasi untuk memilih):", dbName), tables, nil)
	return dbName, tableNames, err
}
