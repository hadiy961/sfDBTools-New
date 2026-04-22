package copy

import (
	"context"
	"fmt"
	"strings"

	"sfdbtools/internal/app/usersgrants"
	profileconn "sfdbtools/internal/app/profile/connection"
	"sfdbtools/internal/domain"
)

// CopyGrants menyalin user dan hak akses (grants) dari database sumber ke database tujuan.
func (s *Service) CopyGrants(ctx context.Context, profile *domain.ProfileInfo, sourceDB, targetDB string) error {
	client, err := profileconn.ConnectWithProfile(s.cfg, profile, "")
	if err != nil {
		return fmt.Errorf("gagal koneksi ke database untuk copy grants: %w", err)
	}
	defer client.Close()

	s.log.Infof("Mengekstrak hak akses user untuk database: %s", sourceDB)

	// 1. Ekstrak SQL User & Grants untuk database sumber
	opts := usersgrants.ExportOptions{
		Databases:          []string{sourceDB},
		IncludeCreateUser:  true,
		IncludeGrants:      true,
		FlushPrivileges:    true,
		ExcludeSystemUsers: true,
	}

	sql, _, err := usersgrants.ExportSQL(ctx, client, opts)
	if err != nil {
		if strings.Contains(err.Error(), "tidak ditemukan user dengan grants") {
			s.log.Infof("Tidak ada user khusus yang memiliki akses ke database '%s'. Skip copy grants.", sourceDB)
			return nil
		}
		return fmt.Errorf("gagal mengekstrak grants: %w", err)
	}

	if sql == "" {
		return nil
	}

	// 2. Transformasi SQL: Ganti nama database sumber menjadi target
	// Kita harus hati-hati dengan quoting.
	transformedSQL := s.transformGrantsSQL(sql, sourceDB, targetDB)

	s.log.Infof("Menerapkan hak akses ke database baru: %s", targetDB)

	// 3. Eksekusi SQL hasil transformasi
	// Pisahkan per statement karena driver Go biasanya tidak support multiple statements dalam satu Exec.
	statements := strings.Split(transformedSQL, ";")
	for _, stmt := range statements {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" || strings.HasPrefix(stmt, "--") {
			continue
		}
		if _, err := client.ExecContextWithRetry(ctx, stmt); err != nil {
			// Jika error karena user sudah ada, kita bisa ignore atau log
			if strings.Contains(strings.ToLower(err.Error()), "already exists") {
				continue
			}
			return fmt.Errorf("gagal menerapkan grant: %w (stmt: %s)", err, stmt)
		}
	}

	return nil
}

func (s *Service) transformGrantsSQL(sql, sourceDB, targetDB string) string {
	// Pola umum: ON `db`.* atau ON db.*
	// Kita ganti dengan case-insensitive jika perlu, tapi biasanya SHOW GRANTS mengikuti case DB.
	
	// 1. Ganti dengan backticks
	sourceQuoted := "`" + sourceDB + "`"
	targetQuoted := "`" + targetDB + "`"
	res := strings.ReplaceAll(sql, " ON "+sourceQuoted+".", " ON "+targetQuoted+".")
	
	// 2. Ganti tanpa backticks (untuk safety)
	res = strings.ReplaceAll(res, " ON "+sourceDB+".", " ON "+targetDB+".")
	
	return res
}
