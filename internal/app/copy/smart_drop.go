package copy

import (
	"context"
	"fmt"
	"sfdbtools/internal/shared/database"
)

// SmartDropDatabaseObjects menghapus seluruh isi database (tabel, view, routine, trigger) 
// tanpa menghapus database itu sendiri, guna mempertahankan Character Set, Collation, dan Grants.
func (s *Service) SmartDropDatabaseObjects(ctx context.Context, client *database.Client, dbName string) error {
	s.log.Infof("Memulai pembersihan objek di database target: %s", dbName)

	// 1. Discovery seluruh objek yang ada di target
	objects, _ := s.DiscoverTablesAndViews(ctx, client, dbName)
	procs, funcs, _ := s.DiscoverRoutines(ctx, client, dbName)
	
	totalWork := len(objects) + len(procs) + len(funcs)
	if totalWork == 0 {
		s.log.Info("Database target sudah bersih. Melanjutkan...")
		return nil
	}

	s.log.Infof("Ditemukan %d objek untuk dibersihkan.", totalWork)

	// 2. Disable Foreign Key Checks agar drop tidak error urutan
	if _, err := client.ExecContextWithRetry(ctx, "SET FOREIGN_KEY_CHECKS = 0"); err != nil {
		return fmt.Errorf("gagal menonaktifkan foreign key checks: %w", err)
	}
	defer func() {
		_, _ = client.ExecContextWithRetry(ctx, "SET FOREIGN_KEY_CHECKS = 1")
	}()

	completed := 0

	// 3. Drop Tables & Views
	for _, obj := range objects {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		typeName := "TABLE"
		if obj.Type == TableTypeView {
			typeName = "VIEW"
		}
		
		query := fmt.Sprintf("DROP %s IF EXISTS `%s`.`%s` ", typeName, dbName, obj.Name) 
		if _, err := client.ExecContextWithRetry(ctx, query); err != nil {
			s.log.Warnf("\nGagal drop %s %s: %v", typeName, obj.Name, err)
		}

		completed++
		s.printCleanupProgress(completed, totalWork, obj.Name)
	}

	// 4. Drop Routines
	for _, p := range procs {
		query := fmt.Sprintf("DROP PROCEDURE IF EXISTS `%s`.`%s` ", dbName, p)
		_, _ = client.ExecContextWithRetry(ctx, query)
		completed++
		s.printCleanupProgress(completed, totalWork, p)
	}

	for _, f := range funcs {
		query := fmt.Sprintf("DROP FUNCTION IF EXISTS `%s`.`%s` ", dbName, f)
		_, _ = client.ExecContextWithRetry(ctx, query)
		completed++
		s.printCleanupProgress(completed, totalWork, f)
	}

	fmt.Println() // Selesai
	s.log.Info("Pembersihan database target selesai.")

	return nil
}

func (s *Service) printCleanupProgress(current, total int, currentName string) {
	percent := float64(current) * 100 / float64(total)
	fmt.Printf("\r  ⏳ Progres Cleanup: %d/%d objek (%.1f%%) [%s]         ", current, total, percent, currentName)
}
