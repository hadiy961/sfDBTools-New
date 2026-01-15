// File : internal/app/backup/setup/session.go
// Deskripsi : Setup session backup (termasuk loop interaktif untuk mode ALL)
// Author : Hadiyatna Muflihun
// Tanggal : 2025-12-30
// Last Modified : 14 Januari 2026

package setup

import (
	"context"
	"fmt"
	"strings"

	"sfdbtools/internal/app/backup/display"
	profileconn "sfdbtools/internal/app/profile/connection"
	"sfdbtools/internal/domain"
	"sfdbtools/internal/shared/consts"
	"sfdbtools/internal/shared/database"
	"sfdbtools/internal/shared/validation"
	"sfdbtools/internal/ui/print"
	"sfdbtools/internal/ui/prompt"
)

type PathGenerator func(ctx context.Context, client *database.Client, dbFiltered []string) ([]string, error)

// PrepareBackupSession runs the whole pre-backup preparation flow.
func (s *Setup) PrepareBackupSession(ctx context.Context, headerTitle string, nonInteractive bool, genPaths PathGenerator) (client *database.Client, dbFiltered []string, err error) {
	if headerTitle != "" {
		print.PrintAppHeader(headerTitle)
	}

	// Normalisasi input awal (misalnya dari flag/ENV)
	s.Options.Ticket = strings.TrimSpace(s.Options.Ticket)

	if err = s.CheckAndSelectConfigFile(); err != nil {
		return nil, nil, err
	}

	// Ticket selalu diminta sebelum koneksi DB dibuat (sesuai keputusan Persiapan 0).
	// Hanya untuk mode interaktif dan hanya jika ticket masih kosong.
	ticketRequired := isInteractiveMode(s.Options.Mode)
	if !nonInteractive && ticketRequired && s.Options.Ticket == "" {
		if ticketErr := s.changeBackupTicketInteractive(); ticketErr != nil {
			return nil, nil, ticketErr
		}
	}

	client, err = profileconn.ConnectWithProfile(&s.Options.Profile, consts.DefaultInitialDatabase)
	if err != nil {
		return nil, nil, err
	}

	var success bool
	defer func() {
		if !success && client != nil {
			client.Close()
		}
	}()

	serverHostname, err := client.GetServerHostname(ctx)
	if err != nil {
		s.Log.Warnf("gagal mendapatkan hostname dari server: %v, menggunakan dari config", err)
		serverHostname = s.Options.Profile.DBInfo.Host
		s.Options.Profile.DBInfo.HostName = serverHostname
		s.Log.Infof("menggunakan hostname fallback dari config: %s", serverHostname)
	} else {
		s.Options.Profile.DBInfo.HostName = serverHostname
		s.Log.Infof("menggunakan hostname dari server: %s", serverHostname)
	}

	if genPaths == nil {
		return nil, nil, fmt.Errorf("path generator tidak tersedia")
	}

	customOutputDir := s.Options.OutputDir

	interactiveEditEnabled := isInteractiveMode(s.Options.Mode)

	// Interactive edit loop khusus untuk backup all & single-variant modes (hanya jika interaktif)
	for {
		var stats *domain.FilterStats
		dbFiltered, stats, err = s.GetFilteredDatabases(ctx, client)
		if err != nil {
			if stats != nil {
				print.PrintFilterStats(stats, consts.FeatureBackup, s.Log)
			}
			return nil, nil, fmt.Errorf("gagal mendapatkan daftar database: %w", err)
		}

		if s.Options.Mode == consts.ModeAll && stats != nil && s.ExcludedDatabases != nil {
			*s.ExcludedDatabases = stats.ExcludedDatabases
			s.Log.Infof("Menyimpan %d excluded databases untuk metadata", len(*s.ExcludedDatabases))
			if len(*s.ExcludedDatabases) > 0 {
				s.Log.Debugf("Excluded databases: %v", *s.ExcludedDatabases)
			}
		}

		if len(dbFiltered) == 0 {
			print.PrintFilterStats(stats, consts.FeatureBackup, s.Log)
			print.PrintError("Tidak ada database yang tersedia setelah filtering!")
			if stats != nil {
				s.DisplayFilterWarnings(stats)
			}
			return nil, nil, fmt.Errorf("tidak ada database tersedia untuk backup setelah filtering")
		}

		// Generate output directory and filename preview (and expand dbFiltered for single/primary/secondary).
		dbFiltered, err = genPaths(ctx, client, dbFiltered)
		if err != nil {
			return nil, nil, err
		}
		if len(dbFiltered) == 0 {
			return nil, nil, fmt.Errorf("path generation menghasilkan daftar database kosong")
		}

		// Log daftar database yang akan di-backup untuk mode ALL (penting untuk mode background).
		if s.Options.Mode == consts.ModeAll {
			s.Log.Infof("Database yang akan di-backup (total=%d)", len(dbFiltered))
			if len(dbFiltered) <= consts.MaxDisplayDatabases {
				s.Log.Infof("Daftar database: %s", strings.Join(dbFiltered, ", "))
			} else {
				s.Log.Infof("Daftar database (first %d): %s", consts.MaxDisplayDatabases, strings.Join(dbFiltered[:consts.MaxDisplayDatabases], ", "))
				// s.Log.Debugf("Daftar database lengkap: %v", dbFiltered)
			}
		}

		// Normalisasi (misalnya jika ticket diubah di menu edit pada langkah selanjutnya).
		s.Options.Ticket = strings.TrimSpace(s.Options.Ticket)

		// Jika user sudah set backup-dir interaktif, pertahankan override ini.
		if customOutputDir != "" {
			s.Options.OutputDir = customOutputDir
		}

		// Jika mode non-interaktif: jangan tampilkan menu/edit/konfirmasi apapun.
		if nonInteractive {
			success = true
			return client, dbFiltered, nil
		}

		// Default behavior untuk mode selain all/single/primary/secondary tetap pakai confirm biasa.
		if !interactiveEditEnabled {
			break
		}

		// Mode ALL/SINGLE/PRIMARY/SECONDARY + interaktif: tampilkan opsi + beri pilihan edit.
		displayer := display.NewOptionsDisplayer(s.Options)
		displayer.Render()

		action, _, selErr := prompt.SelectOne(
			"Pilih aksi",
			[]string{"Lanjutkan", "Ubah opsi", "Batalkan"},
			-1,
		)
		if selErr != nil {
			return nil, nil, selErr
		}

		switch action {
		case "Lanjutkan":
			if s.Options.Ticket == "" {
				print.PrintError("Ticket number belum diisi.")
				print.PrintError("Pilih 'Ubah opsi' untuk mengisi ticket, atau isi via flag --ticket.")
				prompt.WaitForEnter("Tekan Enter untuk kembali ke opsi...")
				continue
			}
			// validasi minimal: jika encryption aktif, backup key wajib tersedia
			if s.Options.Encryption.Enabled && s.Options.Encryption.Key == "" {
				print.PrintError("Encryption diaktifkan tapi backup key belum tersedia.")
				print.PrintError("Isi via flag --backup-key, ENV SFDB_BACKUP_ENCRYPTION_KEY, atau pilih 'Ubah opsi' untuk input interaktif.")
				prompt.WaitForEnter("Tekan Enter untuk kembali ke opsi...")
				continue
			}
			success = true
			return client, dbFiltered, nil
		case "Batalkan":
			return nil, nil, validation.ErrUserCancelled
		case "Ubah opsi":
			if err := s.editBackupOptionsInteractive(ctx, &client, &customOutputDir, s.Options.Mode); err != nil {
				return nil, nil, err
			}
			// Loop lagi untuk re-filter & re-generate preview sesuai opsi terbaru
			continue
		default:
			continue
		}
	}

	// Mode selain ALL (interaktif): konfirmasi sebelum lanjut
	if !nonInteractive {
		if proceed, askErr := display.NewOptionsDisplayer(s.Options).Display(); askErr != nil {
			return nil, nil, askErr
		} else if !proceed {
			return nil, nil, validation.ErrUserCancelled
		}
	}

	success = true
	return client, dbFiltered, nil
}
