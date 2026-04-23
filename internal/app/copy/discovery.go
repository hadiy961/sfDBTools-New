package copy

import (
	"context"
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

	var triggers []string
	for rows.Next() {
		cols, _ := rows.Columns()
		dest := make([]interface{}, len(cols))
		var triggerName string
		for i := range dest {
			if i == 0 {
				dest[i] = &triggerName
			} else {
				dest[i] = new(interface{})
			}
		}

		if err := rows.Scan(dest...); err == nil {
			triggers = append(triggers, triggerName)
		}
	}
	return triggers, nil
}

// DiscoverRoutines me-list Stored Procedures dan Functions.
func (s *Service) DiscoverRoutines(ctx context.Context, client *database.Client, dbName string) (procedures []string, functions []string, err error) {
	// Procedures
	procQuery := fmt.Sprintf("SHOW PROCEDURE STATUS WHERE Db = '%s' ", dbName)
	pRows, err := client.QueryContextWithRetry(ctx, procQuery)
	if err == nil {
		for pRows.Next() {
			var db, name string
			cols, _ := pRows.Columns()
			dest := make([]interface{}, len(cols))
			for i := range dest {
				if i == 0 { dest[i] = &db } else if i == 1 { dest[i] = &name } else { dest[i] = new(interface{}) }
			}
			if err := pRows.Scan(dest...); err == nil {
				procedures = append(procedures, name)
			}
		}
		pRows.Close()
	}

	// Functions
	funcQuery := fmt.Sprintf("SHOW FUNCTION STATUS WHERE Db = '%s' ", dbName)
	fRows, err := client.QueryContextWithRetry(ctx, funcQuery)
	if err == nil {
		for fRows.Next() {
			var db, name string
			cols, _ := fRows.Columns()
			dest := make([]interface{}, len(cols))
			for i := range dest {
				if i == 0 { dest[i] = &db } else if i == 1 { dest[i] = &name } else { dest[i] = new(interface{}) }
			}
			if err := fRows.Scan(dest...); err == nil {
				functions = append(functions, name)
			}
		}
		fRows.Close()
	}

	return procedures, functions, nil
}

// DiscoverEvents me-list seluruh Scheduled Events.
func (s *Service) DiscoverEvents(ctx context.Context, client *database.Client, dbName string) ([]string, error) {
	query := fmt.Sprintf("SHOW EVENTS FROM `%s` ", dbName)
	rows, err := client.QueryContextWithRetry(ctx, query)
	if err != nil {
		return nil, nil
	}
	defer rows.Close()

	var events []string
	for rows.Next() {
		var db, name string
		cols, _ := rows.Columns()
		dest := make([]interface{}, len(cols))
		for i := range dest {
			if i == 0 { dest[i] = &db } else if i == 1 { dest[i] = &name } else { dest[i] = new(interface{}) }
		}
		if err := rows.Scan(dest...); err == nil {
			events = append(events, name)
		}
	}
	return events, nil
}
