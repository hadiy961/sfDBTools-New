package copy

import (
	"context"
	"database/sql"
	"fmt"
	"sfdbtools/internal/shared/database"
)

// TableType adalah enumerasi tipe objek tabel.
type TableType string

const (
	TableTypeBaseTable TableType = "BASE TABLE"
	TableTypeView      TableType = "VIEW"
)

// DBObject merepresentasikan objek dasar di dalam database (Tabel atau View).
type DBObject struct {
	Name string
	Type TableType
}

// DiscoverTablesAndViews mengambil daftar tabel dan view secara instan menggunakan SHOW FULL TABLES.
func (s *Service) DiscoverTablesAndViews(ctx context.Context, client *database.Client, dbName string) ([]DBObject, error) {
	query := fmt.Sprintf("SHOW FULL TABLES FROM `%s` ", dbName)
	rows, err := client.QueryContextWithRetry(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("gagal me-list tabel: %w", err)
	}
	defer rows.Close()

	var objects []DBObject
	for rows.Next() {
		var name string
		var tableType string
		if err := rows.Scan(&name, &tableType); err != nil {
			s.log.Warnf("Gagal scan baris tabel di %s: %v", dbName, err)
			continue
		}
		objects = append(objects, DBObject{
			Name: name,
			Type: TableType(tableType),
		})
	}

	return objects, nil
}

// DiscoverTriggers me-list seluruh trigger pada database tertentu.
func (s *Service) DiscoverTriggers(ctx context.Context, client *database.Client, dbName string) ([]string, error) {
	query := fmt.Sprintf("SHOW TRIGGERS FROM `%s` ", dbName)
	rows, err := client.QueryContextWithRetry(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("gagal me-list triggers: %w", err)
	}
	defer rows.Close()

	return s.scanColumnFromRows(rows, 0) // Trigger name is usually column 0
}

// DiscoverRoutines me-list Stored Procedures dan Functions.
func (s *Service) DiscoverRoutines(ctx context.Context, client *database.Client, dbName string) (procedures []string, functions []string, err error) {
	// Procedures
	procQuery := "SHOW PROCEDURE STATUS WHERE Db = ?"
	pRows, err := client.QueryContextWithRetry(ctx, procQuery, dbName)
	if err != nil {
		return nil, nil, fmt.Errorf("gagal me-list procedures: %w", err)
	}
	procedures, err = s.scanColumnFromRows(pRows, 1) // Name is column 1
	pRows.Close()
	if err != nil {
		return nil, nil, err
	}

	// Functions
	funcQuery := "SHOW FUNCTION STATUS WHERE Db = ?"
	fRows, err := client.QueryContextWithRetry(ctx, funcQuery, dbName)
	if err != nil {
		return procedures, nil, fmt.Errorf("gagal me-list functions: %w", err)
	}
	functions, err = s.scanColumnFromRows(fRows, 1) // Name is column 1
	fRows.Close()

	return procedures, functions, err
}

// DiscoverEvents me-list seluruh Scheduled Events.
func (s *Service) DiscoverEvents(ctx context.Context, client *database.Client, dbName string) ([]string, error) {
	query := fmt.Sprintf("SHOW EVENTS FROM `%s` ", dbName)
	rows, err := client.QueryContextWithRetry(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("gagal me-list events: %w", err)
	}
	defer rows.Close()

	return s.scanColumnFromRows(rows, 1) // Name is column 1
}

// scanColumnFromRows adalah helper untuk mengekstrak string dari kolom tertentu dalam hasil query.
func (s *Service) scanColumnFromRows(rows *sql.Rows, columnIndex int) ([]string, error) {
	var results []string
	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		dest := make([]interface{}, len(cols))
		var value string
		for i := range dest {
			if i == columnIndex {
				dest[i] = &value
			} else {
				dest[i] = new(interface{})
			}
		}

		if err := rows.Scan(dest...); err != nil {
			return nil, fmt.Errorf("gagal scan kolom %d: %w", columnIndex, err)
		}
		results = append(results, value)
	}
	return results, nil
}
