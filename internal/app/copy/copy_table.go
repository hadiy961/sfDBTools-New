package copy

import (
	"context"
	"fmt"

	profileconn "sfdbtools/internal/app/profile/connection"
	"sfdbtools/internal/domain"
	"sfdbtools/internal/shared/database"
	"sfdbtools/internal/ui/progress"
	"sfdbtools/internal/ui/prompt"
)

// CopyTable melakukan penyalinan tabel spesifik.
func (s *Service) CopyTable(ctx context.Context, profile *domain.ProfileInfo, sourceDB, sourceTable, targetDB, targetTable string, schemaOnly, force, backupFirst, includeGrants, verify, nonInteractive bool) (string, string, string, error) {
	client, err := profileconn.ConnectWithProfile(s.cfg, profile, "")
	if err != nil {
		return "", "", "", fmt.Errorf("gagal koneksi ke database: %w", err)
	}
	defer client.Close()

	// Validation
	exists, err := s.validateTableExists(ctx, client, sourceDB, sourceTable)
	if err != nil || !exists {
		return "", "", "", fmt.Errorf("tabel sumber %s.%s tidak ditemukan", sourceDB, sourceTable)
	}

	if targetDB == "" {
		targetDB = sourceDB
	}

	// Create DB if not exists
	if err := client.CreateDatabaseIfNotExists(ctx, targetDB); err != nil {
		return "", "", "", err
	}

	targetExists, err := s.validateTableExists(ctx, client, targetDB, targetTable)
	if err != nil {
		return "", "", "", err
	}

	if targetExists {
		if !nonInteractive {
			s.log.Warnf("PERINGATAN: Tabel target '%s.%s' sudah ada!", targetDB, targetTable)

			if !force && !backupFirst {
				choice, _, err := prompt.SelectOne("Tabel target sudah ada. Apa yang ingin Anda lakukan?",
					[]string{"Batalkan", "Backup dulu baru timpa", "Timpa langsung (Berisiko!)"}, 0)
				if err != nil || choice == "Batalkan" {
					return "", "", "", fmt.Errorf("operasi dibatalkan")
				}
				if choice == "Backup dulu baru timpa" {
					backupFirst = true
				} else {
					force = true
				}
			}

			if backupFirst {
				confirm, err := prompt.Confirm(fmt.Sprintf("Yakin ingin membackup lalu menimpa tabel '%s.%s'?", targetDB, targetTable), true)
				if err != nil || !confirm {
					return "", "", "", fmt.Errorf("operasi dibatalkan")
				}
			} else if force {
				s.log.Warnf("!!! PERHATIAN !!!")
				confirm1, _ := prompt.Confirm(fmt.Sprintf("Yakin ingin MENIMPA tabel '%s.%s' TANPA backup?", targetDB, targetTable), false)
				if !confirm1 {
					return "", "", "", fmt.Errorf("operasi dibatalkan")
				}
				confirm2, _ := prompt.Confirm("Benar-benar yakin? Data lama akan hilang.", false)
				if !confirm2 {
					return "", "", "", fmt.Errorf("operasi dibatalkan")
				}
			}
		} else {
			if !force && !backupFirst {
				return "", "", "", fmt.Errorf("tabel target '%s.%s' sudah ada. Gunakan --force atau --backup-first", targetDB, targetTable)
			}
		}

		if backupFirst {
			s.log.Infof("Menjalankan backup pengamanan untuk tabel: %s.%s", targetDB, targetTable)
			if err := s.runSafetyTableBackup(ctx, profile, client, targetDB, targetTable); err != nil {
				return "", "", "", fmt.Errorf("gagal melakukan backup pengamanan tabel: %w", err)
			}
		}
	}

	s.log.Debugf("Memulai copy tabel: %s.%s -> %s.%s", sourceDB, sourceTable, targetDB, targetTable)

	spin := progress.NewSpinnerWithElapsed(fmt.Sprintf("Copying table %s.%s", sourceDB, sourceTable))
	spin.Start()
	defer spin.Stop()

	if force || backupFirst {
		dropSQL := fmt.Sprintf("DROP TABLE IF EXISTS `%s`.`%s` ", targetDB, targetTable)
		_, _ = client.ExecContextWithRetry(ctx, dropSQL)
	}

	createSQL := fmt.Sprintf("CREATE TABLE IF NOT EXISTS `%s`.`%s` LIKE `%s`.`%s` ", targetDB, targetTable, sourceDB, sourceTable)
	if _, err := client.ExecContextWithRetry(ctx, createSQL); err != nil {
		return "", "", "", fmt.Errorf("gagal membuat struktur tabel target: %w", err)
	}

	if !schemaOnly {
		insertSQL := fmt.Sprintf("INSERT INTO `%s`.`%s` SELECT * FROM `%s`.`%s` ", targetDB, targetTable, sourceDB, sourceTable)
		if _, err := client.ExecContextWithRetry(ctx, insertSQL); err != nil {
			return "", "", "", fmt.Errorf("gagal menyalin data tabel: %w", err)
		}
	}

	// 6. Copy Grants (if enabled)
	if includeGrants {
		// Untuk tabel tunggal, kita tetap menyalin grants database-level sebagai pendekatan aman
		if err := s.CopyGrants(ctx, profile, sourceDB, targetDB); err != nil {
			s.log.Warnf("Gagal menyalin hak akses user: %v", err)
		}
	}

	// 7. Verify Checksum
	verifyStatus := "-"
	if verify && !schemaOnly {
		ok, err := s.VerifyChecksum(ctx, client, sourceDB, sourceTable, targetDB, targetTable)
		if err != nil {
			verifyStatus = "Error Verifikasi"
			s.log.Warnf("Gagal verifikasi checksum %s: %v", sourceTable, err)
		} else if ok {
			verifyStatus = "Cocok"
		} else {
			verifyStatus = "Gagal (Mismatch)"
		}
	}

	return targetDB, targetTable, verifyStatus, nil
}

func (s *Service) validateTableExists(ctx context.Context, client *database.Client, dbName, tableName string) (bool, error) {
	query := "SELECT TABLE_NAME FROM information_schema.TABLES WHERE TABLE_SCHEMA = ? AND TABLE_NAME = ? LIMIT 1"
	rows, err := client.QueryContextWithRetry(ctx, query, dbName, tableName)
	if err != nil {
		return false, err
	}
	defer rows.Close()
	return rows.Next(), nil
}
