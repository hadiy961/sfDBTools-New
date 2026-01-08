package restore

import (
	"context"
	"fmt"
	"sfdbtools/internal/shared/consts"
	"strings"
)

// CreateTempDatabaseIfNeeded membuat database <db>_temp (IF NOT EXISTS).
// Sesuai requirement: tidak membuat temp untuk DB yang ber-suffix _dmart.
func (s *Service) CreateTempDatabaseIfNeeded(ctx context.Context, dbName string) (string, error) {
	dbName = strings.TrimSpace(dbName)
	if dbName == "" {
		return "", fmt.Errorf("nama database kosong")
	}
	if strings.HasSuffix(dbName, consts.SuffixDmart) {
		return "", nil
	}
	if strings.HasSuffix(dbName, consts.SuffixTemp) {
		return "", nil
	}
	if s.TargetClient == nil {
		return "", fmt.Errorf("target client belum siap")
	}

	tempDB := dbName + consts.SuffixTemp
	if err := s.TargetClient.CreateDatabaseIfNotExists(ctx, tempDB); err != nil {
		return "", err
	}
	return tempDB, nil
}

func escapeMySQLLiteral(s string) string {
	// Minimal escaping for building SQL string literals safely.
	// MySQL string literal uses backslash escaping.
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "'", "\\'")
	return s
}

// CopyDatabaseGrants menyalin GRANT db-level (ON <db>.*) dari sourceDB ke targetDB.
// Operasi ini bersifat best-effort; caller biasanya akan treat error sebagai warning.
func (s *Service) CopyDatabaseGrants(ctx context.Context, sourceDB string, targetDB string) error {
	sourceDB = strings.TrimSpace(sourceDB)
	targetDB = strings.TrimSpace(targetDB)
	if sourceDB == "" || targetDB == "" {
		return fmt.Errorf("sourceDB/targetDB kosong")
	}
	if sourceDB == targetDB {
		return nil
	}
	if s.TargetClient == nil {
		return fmt.Errorf("target client belum siap")
	}

	// Ensure target DB exists (idempotent)
	if err := s.TargetClient.CreateDatabaseIfNotExists(ctx, targetDB); err != nil {
		return fmt.Errorf("gagal memastikan target DB ada: %w", err)
	}

	// Find users that have privileges on sourceDB.
	// mysql.db exists on MySQL/MariaDB; we keep scope DB-level only.
	queryUsers := "SELECT DISTINCT User, Host FROM mysql.db WHERE Db = ?"
	rows, err := s.TargetClient.QueryContextWithRetry(ctx, queryUsers, sourceDB)
	if err != nil {
		return fmt.Errorf("gagal query mysql.db: %w", err)
	}
	defer rows.Close()

	type userHost struct {
		user string
		host string
	}
	users := make([]userHost, 0, 16)
	for rows.Next() {
		var u, h string
		if scanErr := rows.Scan(&u, &h); scanErr != nil {
			return fmt.Errorf("gagal scan mysql.db: %w", scanErr)
		}
		u = strings.TrimSpace(u)
		h = strings.TrimSpace(h)
		if u == "" || h == "" {
			continue
		}
		users = append(users, userHost{user: u, host: h})
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("error iterasi mysql.db: %w", err)
	}

	if len(users) == 0 {
		return nil
	}

	fromTokenTick := fmt.Sprintf("ON `%s`.*", sourceDB)
	toTokenTick := fmt.Sprintf("ON `%s`.*", targetDB)
	fromTokenPlain := fmt.Sprintf("ON %s.*", sourceDB)
	toTokenPlain := fmt.Sprintf("ON %s.*", targetDB)

	for _, uh := range users {
		userLit := escapeMySQLLiteral(uh.user)
		hostLit := escapeMySQLLiteral(uh.host)
		show := fmt.Sprintf("SHOW GRANTS FOR '%s'@'%s'", userLit, hostLit)

		grRows, gerr := s.TargetClient.QueryContextWithRetry(ctx, show)
		if gerr != nil {
			// Skip user if cannot read grants
			continue
		}

		grantsToApply := make([]string, 0, 8)
		for grRows.Next() {
			var line string
			if err := grRows.Scan(&line); err != nil {
				continue
			}
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}

			// Only copy DB-level grants: ON <db>.*
			if strings.Contains(line, fromTokenTick) {
				grantsToApply = append(grantsToApply, strings.Replace(line, fromTokenTick, toTokenTick, 1))
				continue
			}
			if strings.Contains(line, fromTokenPlain) {
				grantsToApply = append(grantsToApply, strings.Replace(line, fromTokenPlain, toTokenPlain, 1))
				continue
			}
		}
		_ = grRows.Close()

		for _, stmt := range grantsToApply {
			if _, err := s.TargetClient.ExecContextWithRetry(ctx, stmt); err != nil {
				return fmt.Errorf("gagal apply grant (%s@%s): %w", uh.user, uh.host, err)
			}
		}
	}

	return nil
}
