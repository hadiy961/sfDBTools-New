package copy

import (
	"context"
	"fmt"
	"sfdbtools/internal/shared/database"
)

// SmartDropDatabaseObjects menghapus seluruh isi database (tabel, view, routine, trigger)
// tanpa menghapus database itu sendiri, guna mempertahankan Character Set, Collation, dan Grants.
func (s *Service) SmartDropDatabaseObjects(ctx context.Context, client *database.Client, dbName string) error {
	s.log.Infof("Membersihkan objek lama di database target: %s", dbName)

	// 1. Disable Foreign Key Checks agar drop tabel tidak error urutan
	if _, err := client.ExecContextWithRetry(ctx, "SET FOREIGN_KEY_CHECKS = 0"); err != nil {
		return fmt.Errorf("gagal menonaktifkan foreign key checks: %w", err)
	}
	defer func() {
		_, _ = client.ExecContextWithRetry(ctx, "SET FOREIGN_KEY_CHECKS = 1")
	}()

	// 2. Discover & Drop Tables/Views
	objects, err := s.DiscoverTablesAndViews(ctx, client, dbName)
	if err == nil {
		for _, obj := range objects {
			typeName := "TABLE"
			if obj.Type == TableTypeView {
				typeName = "VIEW"
			}
			query := fmt.Sprintf("DROP %s IF EXISTS `%s`.`%s` ", typeName, dbName, obj.Name)
			if _, err := client.ExecContextWithRetry(ctx, query); err != nil {
				s.log.Warnf("Gagal drop %s %s: %v", typeName, obj.Name, err)
			}
		}
	}

	// 3. Discover & Drop Procedures/Functions
	procs, funcs, err := s.DiscoverRoutines(ctx, client, dbName)
	if err == nil {
		for _, p := range procs {
			query := fmt.Sprintf("DROP PROCEDURE IF EXISTS `%s`.`%s` ", dbName, p)
			_, _ = client.ExecContextWithRetry(ctx, query)
		}
		for _, f := range funcs {
			query := fmt.Sprintf("DROP FUNCTION IF EXISTS `%s`.`%s` ", dbName, f)
			_, _ = client.ExecContextWithRetry(ctx, query)
		}
	}

	// Triggers otomatis hilang saat tabelnya di-drop, tapi untuk amannya
	// kita biarkan logic drop table yang menanganinya.

	return nil
}
