package copy

import (
	"context"
	"fmt"
	"strings"
	"time"

	copyexec "sfdbtools/internal/app/copy/execution"
	profileconn "sfdbtools/internal/app/profile/connection"
	"sfdbtools/internal/domain"
	"sfdbtools/internal/ui/prompt"
)

// CopyDatabase melakukan penyalinan satu database utuh.
func (s *Service) CopyDatabase(ctx context.Context, profile *domain.ProfileInfo, sourceDB, targetDB string, schemaOnly, useDisk, useConcurrent bool, workers int, limitSpeed int64, force, backupFirst, includeGrants, verify, skipRoutines, skipEvents, skipTriggers, nonInteractive bool) (string, error) {
	client, err := profileconn.ConnectWithProfile(s.cfg, profile, "")
	if err != nil {
		return "", fmt.Errorf("gagal koneksi ke database: %w", err)
	}
	defer client.Close()

	// Pre-flight checks
	exists, err := client.CheckDatabaseExists(ctx, sourceDB)
	if err != nil {
		return "", err
	}
	if !exists {
		return "", fmt.Errorf("database sumber '%s' tidak ditemukan", sourceDB)
	}

	if targetDB == "" {
		if nonInteractive {
			targetDB = fmt.Sprintf("%s_copy_%s", sourceDB, time.Now().Format("20060102"))
			s.log.Infof("Nama database target otomatis: %s", targetDB)
		} else {
			targetDB, err = prompt.AskText("Masukkan nama database target:", prompt.WithDefault(fmt.Sprintf("%s_copy_%s", sourceDB, time.Now().Format("20060102"))))
			if err != nil {
				return "", err
			}
		}
	}

	if strings.EqualFold(sourceDB, targetDB) {
		return "", fmt.Errorf("database target tidak boleh sama dengan database sumber")
	}

	targetExists, err := client.CheckDatabaseExists(ctx, targetDB)
	if err != nil {
		return "", err
	}

	if targetExists {
		if !nonInteractive {
			// Mode Interaktif: Konfirmasi berlapis
			s.log.Warnf("PERINGATAN: Database target '%s' sudah ada!", targetDB)

			// Jika belum ada flag yang menspesifikasi tindakan, tanya user
			if !force && !backupFirst {
				choice, _, err := prompt.SelectOne("Database target sudah ada. Apa yang ingin Anda lakukan?",
					[]string{"Batalkan", "Backup dulu baru timpa (Sangat Disarankan)", "Timpa langsung (Data lama hilang!)"}, 0)
				if err != nil || choice == "Batalkan" {
					return "", fmt.Errorf("operasi dibatalkan")
				}
				if choice == "Backup dulu baru timpa (Sangat Disarankan)" {
					backupFirst = true
				} else {
					force = true
				}
			}

			// Konfirmasi Akhir
			if backupFirst {
				confirm, err := prompt.Confirm(fmt.Sprintf("Yakin ingin membackup lalu menimpa database '%s'?", targetDB), true)
				if err != nil || !confirm {
					return "", fmt.Errorf("operasi dibatalkan")
				}
			} else if force {
				// Konfirmasi Berlapis untuk Overwrite tanpa Backup
				s.log.Warnf("!!! PERHATIAN !!!")
				s.log.Warnf("Anda memilih untuk MENIMPA database '%s' TANPA backup.", targetDB)
				confirm1, _ := prompt.Confirm("Apakah Anda sadar data lama di target akan hilang permanen?", false)
				if !confirm1 {
					return "", fmt.Errorf("operasi dibatalkan")
				}
				confirm2, _ := prompt.Confirm("Konfirmasi terakhir: Benar-benar ingin menimpa tanpa backup?", false)
				if !confirm2 {
					return "", fmt.Errorf("operasi dibatalkan")
				}
			}
		} else {
			// Mode Non-Interaktif: Cek flag
			if !force && !backupFirst {
				return "", fmt.Errorf("database target '%s' sudah ada. Gunakan --force atau --backup-first", targetDB)
			}
		}

		// Jalankan backup target jika diminta
		if backupFirst {
			s.log.Infof("Menjalankan backup pengamanan untuk target: %s", targetDB)
			if err := s.runSafetyBackup(ctx, profile, client, targetDB); err != nil {
				return "", fmt.Errorf("gagal melakukan backup pengamanan: %w", err)
			}
		}
	}

	// Route to Concurrent Engine if requested
	if useConcurrent && !useDisk && !schemaOnly {
		return s.CopyDatabaseConcurrent(ctx, profile, sourceDB, targetDB, workers, limitSpeed, force, backupFirst, includeGrants, verify, skipRoutines, skipEvents, skipTriggers, nonInteractive)
	}

	if err := client.CreateDatabaseIfNotExists(ctx, targetDB); err != nil {
		return "", err
	}

	// Smart Overwrite (Clean up objects if target already exists)
	if targetExists && (force || backupFirst) {
		if err := s.SmartDropDatabaseObjects(ctx, client, targetDB); err != nil {
			s.log.Warnf("Gagal membersihkan database target secara bersih: %v", err)
		}
	}

	methodName := "Piping"
	if useDisk {
		methodName = "Disk-based"
	}
	s.log.Debugf("Memulai copy database: %s -> %s [Metode: %s]", sourceDB, targetDB, methodName)

	// Execution
	if useDisk {
		if err := s.executeDiskCopy(ctx, profile, client, sourceDB, targetDB, schemaOnly); err != nil {
			return "", err
		}
	} else {
		extraDumpArgs := ""
		if !skipRoutines {
			extraDumpArgs += " --routines"
		}
		if !skipEvents {
			extraDumpArgs += " --events"
		}
		if !skipTriggers {
			extraDumpArgs += " --triggers"
		}

		if err := copyexec.ExecutePiping(ctx, s.log, copyexec.PipingOptions{
			Profile:      profile,
			SourceDB:     sourceDB,
			TargetDB:     targetDB,
			SchemaOnly:   schemaOnly,
			BaseDumpArgs: s.cfg.Backup.MysqlDumpArgs + extraDumpArgs,
			LimitSpeed:   limitSpeed,
		}); err != nil {
			return "", err
		}
	}

	// 6. Copy Grants (if enabled)
	if includeGrants {
		if err := s.CopyGrants(ctx, profile, sourceDB, targetDB); err != nil {
			s.log.Warnf("Gagal menyalin hak akses user: %v", err)
		}
	}

	// 7. Verify Checksum (for non-concurrent)
	if verify && !schemaOnly {
		s.log.Info("Memulai verifikasi database...")
		objects, _ := s.DiscoverTablesAndViews(ctx, client, sourceDB)
		for _, obj := range objects {
			if obj.Type == TableTypeBaseTable {
				ok, _ := s.VerifyChecksum(ctx, client, sourceDB, obj.Name, targetDB, obj.Name)
				if !ok {
					s.log.Warnf("Checksum mismatch pada tabel %s", obj.Name)
				}
			}
		}
	}

	return targetDB, nil
}
