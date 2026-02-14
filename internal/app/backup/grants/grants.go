package grants

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"sfdbtools/internal/app/backup/metadata"
	"sfdbtools/internal/app/usersgrants"
	applog "sfdbtools/internal/services/log"
	"sfdbtools/internal/shared/database"
)

func defaultSystemUsers() []string {
	// Conservatively exclude MySQL internal accounts (common across MySQL/MariaDB).
	return []string{
		"mysql.sys",
		"mysql.session",
		"mysql.infoschema",
	}
}

func uniqTrimNonEmpty(in []string) []string {
	if len(in) == 0 {
		return nil
	}
	out := make([]string, 0, len(in))
	seen := make(map[string]struct{}, len(in))
	for _, s := range in {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		key := strings.ToLower(s)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, s)
	}
	return out
}

func formatDatabasesForError(databases []string) string {
	if len(databases) == 0 {
		return ""
	}
	const max = 10
	if len(databases) <= max {
		return strings.Join(databases, ",")
	}
	head := strings.Join(databases[:max], ",")
	return fmt.Sprintf("%s,...(+%d)", head, len(databases)-max)
}

// parseFilePermissions mengkonversi string permissions (e.g., "0600") ke os.FileMode.
// Jika parsing gagal atau permissions kosong, return default 0600 (lebih restrictive).
func parseFilePermissions(permStr string, logger applog.Logger) os.FileMode {
	const defaultPerm = 0600
	if strings.TrimSpace(permStr) == "" {
		return defaultPerm
	}
	perm, err := strconv.ParseUint(strings.TrimSpace(permStr), 8, 32)
	if err != nil {
		logger.Warnf("Invalid metadata_permissions '%s', using default 0600: %v", permStr, err)
		return defaultPerm
	}
	return os.FileMode(perm)
}

// ExportUserGrantsIfNeeded exports user grants unless excluded or in dry-run.
// Policy backup (konsisten):
// - include CREATE USER + GRANT
// - exclude system users (default + config)
// - file permission mengikuti metadata_permissions
// - jika databases filter dan tidak ada user relevan, treat sebagai non-fatal dan skip
func ExportUserGrantsIfNeeded(
	ctx context.Context,
	client *database.Client,
	log applog.Logger,
	referenceBackupFile string,
	excludeUser bool,
	dryRun bool,
	databases []string,
	requireGrants bool,
	metadataPermissions string,
	configSystemUsers []string,
	excludeGrant bool,
) (string, string, error) {
	if dryRun {
		log.Info("[DRY-RUN] Skip export user grants (dry-run mode aktif)")
		return "", "", nil
	}
	if excludeUser {
		log.Debug("ExcludeUser: flag diaktifkan, skip export user grants")
		return "", "", nil
	}
	if strings.TrimSpace(referenceBackupFile) == "" {
		log.Warn("Tidak ada backup file untuk export user grants")
		return "", "", nil
	}
	if client == nil {
		log.Warn("Skip export user grants: client nil")
		return "", "", nil
	}

	// Pastikan koneksi masih valid.
	if err := client.Ping(ctx); err != nil {
		log.Warnf("Koneksi database tidak valid, mencoba reconnect: %v", err)
		if err := client.Ping(ctx); err != nil {
			log.Errorf("Gagal reconnect ke database: %v", err)
			return "", "", nil
		}
		log.Info("Reconnect ke database berhasil")
	}

	sysUsers := append(defaultSystemUsers(), configSystemUsers...)
	sysUsers = uniqTrimNonEmpty(sysUsers)

	log.Info("Export user grants ke file...")

	// Default behavior: Combined (IncludeCreateUser=true, IncludeGrants=true)
	// Tapi kalau excludeGrant=true, maka hanya IncludeCreateUser=true
	// Kalau excludeGrant=false dan excludeUser=false, maka split output (Users & Grants terpisah)

	// Case 1: Exclude Grant is TRUE -> Only Users (CREATE USER)
	if excludeGrant {
		sqlText, _, err := usersgrants.ExportSQL(ctx, client, usersgrants.ExportOptions{
			Databases:          databases,
			ExcludeSystemUsers: true,
			SystemUsers:        sysUsers,
			IncludeCreateUser:  true,
			IncludeGrants:      false,
			FlushPrivileges:    false,
		})
		if err != nil {
			return "", "", handleExportError(err, databases, requireGrants, log)
		}

		// Naming: karena isinya user definition, kita pakai _users.sql
		userDefPath := metadata.GenerateUserDefinitionFilePath(referenceBackupFile)
		perm := parseFilePermissions(metadataPermissions, log)
		if err := os.WriteFile(userDefPath, []byte(sqlText), perm); err != nil {
			log.Errorf("Gagal menulis file user definitions: %v", err)
			return "", "", nil
		}
		log.Infof("✓ User definitions berhasil disimpan ke: %s", userDefPath)
		return userDefPath, "", nil
	}

	// Case 2: Split Output (Users & Grants separate)
	// Users first
	usersSQL, _, err := usersgrants.ExportSQL(ctx, client, usersgrants.ExportOptions{
		Databases:          databases,
		ExcludeSystemUsers: true,
		SystemUsers:        sysUsers,
		IncludeCreateUser:  true,
		IncludeGrants:      false,
		FlushPrivileges:    false,
	})
	if err != nil {
		return "", "", handleExportError(err, databases, requireGrants, log)
	}

	// Grants second
	grantsSQL, _, err := usersgrants.ExportSQL(ctx, client, usersgrants.ExportOptions{
		Databases:          databases,
		ExcludeSystemUsers: true,
		SystemUsers:        sysUsers,
		IncludeCreateUser:  false,
		IncludeGrants:      true,
		FlushPrivileges:    true,
	})
	if err != nil {
		// Jika grants gagal, kita tetap simpan users?
		// Sebaiknya fail-fast atau log error. Kita return error saja.
		return "", "", handleExportError(err, databases, requireGrants, log)
	}

	perm := parseFilePermissions(metadataPermissions, log)

	// Write Users
	userDefPath := metadata.GenerateUserDefinitionFilePath(referenceBackupFile)
	if err := os.WriteFile(userDefPath, []byte(usersSQL), perm); err != nil {
		log.Errorf("Gagal menulis file user definitions: %v", err)
		return "", "", nil
	}
	log.Infof("✓ User definitions berhasil disimpan ke: %s", userDefPath)

	// Write Grants
	userGrantsPath := metadata.GenerateUserGrantsFilePath(referenceBackupFile)
	if err := os.WriteFile(userGrantsPath, []byte(grantsSQL), perm); err != nil {
		log.Errorf("Gagal menulis file user grants: %v", err)
		// Kembalikan userDefPath karena itu berhasil
		return userDefPath, "", nil
	}
	log.Infof("✓ User grants berhasil disimpan ke: %s", userGrantsPath)

	return userDefPath, userGrantsPath, nil
}

func handleExportError(err error, databases []string, requireGrants bool, log applog.Logger) error {
	if len(databases) > 0 && errors.Is(err, usersgrants.ErrNoUserWithGrants) {
		if requireGrants {
			dbStr := formatDatabasesForError(databases)
			msg := "require-grants aktif: tidak ada user grants yang relevan untuk database terpilih"
			if strings.TrimSpace(dbStr) != "" {
				msg = fmt.Sprintf("%s (databases=%s)", msg, dbStr)
			}
			msg = msg + ". Mitigasi: set --require-grants=false atau nonaktifkan export user grants."
			return fmt.Errorf("%s: %w", msg, err)
		}
		log.Infof("Tidak ada user grants yang relevan untuk database terpilih, skip export user grants")
		return nil
	}
	log.Errorf("Gagal mendapatkan user/grants: %v", err)
	return nil
}

// UpdateMetadataUserGrantsPath updates backup metadata with the actual user grants file path.
func UpdateMetadataUserGrantsPath(log applog.Logger, backupFilePath string, userDefPath string, userGrantsPath string, permissions string) {
	// Kita reuse key UserGrantsFile di metadata untuk path grants.
	// Jika userDefPath ada, kita mungkin perlu method update baru atau update manual.
	// TODO: Pastikan method di metadata.Updater support path tambahan atau kita update satu-satu.
	// Saat ini hanya stub karena engine.go yang memanggil method ini perlu diupdate juga.
}
