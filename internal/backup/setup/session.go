// File : internal/backup/setup/session.go
// Deskripsi : Setup session backup (termasuk loop interaktif untuk mode ALL)
// Author : Hadiyatna Muflihun
// Tanggal : 2025-12-30
// Last Modified : 2025-12-30

package setup

import (
	"context"
	"fmt"
	"strings"
	"time"

	"sfDBTools/internal/backup/display"
	"sfDBTools/internal/types"
	"sfDBTools/pkg/consts"
	"sfDBTools/pkg/database"
	profilehelper "sfDBTools/pkg/helper/profile"
	"sfDBTools/pkg/input"
	"sfDBTools/pkg/ui"
	"sfDBTools/pkg/validation"
)

type PathGenerator func(ctx context.Context, client *database.Client, dbFiltered []string) ([]string, error)

// PrepareBackupSession runs the whole pre-backup preparation flow.
func (s *Setup) PrepareBackupSession(ctx context.Context, headerTitle string, nonInteractive bool, genPaths PathGenerator) (client *database.Client, dbFiltered []string, err error) {
	if headerTitle != "" {
		ui.Headers(headerTitle)
	}

	if err = s.CheckAndSelectConfigFile(); err != nil {
		return nil, nil, err
	}

	client, err = profilehelper.ConnectWithProfile(&s.Options.Profile, consts.DefaultInitialDatabase)
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

	interactiveEditEnabled := s.Options.Mode == consts.ModeAll ||
		s.Options.Mode == consts.ModeSingle ||
		s.Options.Mode == consts.ModePrimary ||
		s.Options.Mode == consts.ModeSecondary ||
		s.Options.Mode == consts.ModeCombined ||
		s.Options.Mode == consts.ModeSeparated

	// Interactive edit loop khusus untuk backup all & single-variant modes (hanya jika interaktif)
	for {
		// Untuk mode ALL interaktif, minta ticket number terlebih dahulu agar muncul di Opsi Backup.
		if !nonInteractive && s.Options.Mode == consts.ModeAll && s.Options.Ticket == "" {
			defaultTicket := fmt.Sprintf("bk-%d", time.Now().UnixNano())
			ticket, ticketErr := input.AskString("Ticket number", defaultTicket, func(ans interface{}) error {
				v, ok := ans.(string)
				if !ok {
					return fmt.Errorf("input tidak valid")
				}
				if v == "" {
					return fmt.Errorf("ticket number tidak boleh kosong")
				}
				return nil
			})
			if ticketErr != nil {
				return nil, nil, ticketErr
			}
			s.Options.Ticket = ticket
		}

		var stats *types.FilterStats
		dbFiltered, stats, err = s.GetFilteredDatabases(ctx, client)
		if err != nil {
			if stats != nil {
				ui.DisplayFilterStats(stats, consts.FeatureBackup, s.Log)
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
			ui.DisplayFilterStats(stats, consts.FeatureBackup, s.Log)
			ui.PrintError("Tidak ada database yang tersedia setelah filtering!")
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

		// Untuk mode SINGLE interaktif, minta ticket number SETELAH database sudah dipilih
		// (urutan: profile -> pilih DB -> input ticket -> display opsi).
		if !nonInteractive && s.Options.Mode == consts.ModeSingle && strings.TrimSpace(s.Options.Ticket) == "" {
			defaultTicket := fmt.Sprintf("bk-%d", time.Now().UnixNano())
			ticket, ticketErr := input.AskString("Ticket number", defaultTicket, func(ans interface{}) error {
				v, ok := ans.(string)
				if !ok {
					return fmt.Errorf("input tidak valid")
				}
				if strings.TrimSpace(v) == "" {
					return fmt.Errorf("ticket number tidak boleh kosong")
				}
				return nil
			})
			if ticketErr != nil {
				return nil, nil, ticketErr
			}
			s.Options.Ticket = strings.TrimSpace(ticket)
		}

		// Untuk mode PRIMARY interaktif, minta ticket number SETELAH database sudah dipilih
		// (urutan: profile -> pilih DB (primary) -> input ticket -> display opsi).
		if !nonInteractive && s.Options.Mode == consts.ModePrimary && strings.TrimSpace(s.Options.Ticket) == "" {
			defaultTicket := fmt.Sprintf("bk-%d", time.Now().UnixNano())
			ticket, ticketErr := input.AskString("Ticket number", defaultTicket, func(ans interface{}) error {
				v, ok := ans.(string)
				if !ok {
					return fmt.Errorf("input tidak valid")
				}
				if strings.TrimSpace(v) == "" {
					return fmt.Errorf("ticket number tidak boleh kosong")
				}
				return nil
			})
			if ticketErr != nil {
				return nil, nil, ticketErr
			}
			s.Options.Ticket = strings.TrimSpace(ticket)
		}

		// Untuk mode SECONDARY interaktif, minta ticket number SETELAH database sudah dipilih
		// (urutan: profile -> pilih DB (secondary) -> input ticket -> display opsi).
		if !nonInteractive && s.Options.Mode == consts.ModeSecondary && strings.TrimSpace(s.Options.Ticket) == "" {
			defaultTicket := fmt.Sprintf("bk-%d", time.Now().UnixNano())
			ticket, ticketErr := input.AskString("Ticket number", defaultTicket, func(ans interface{}) error {
				v, ok := ans.(string)
				if !ok {
					return fmt.Errorf("input tidak valid")
				}
				if strings.TrimSpace(v) == "" {
					return fmt.Errorf("ticket number tidak boleh kosong")
				}
				return nil
			})
			if ticketErr != nil {
				return nil, nil, ticketErr
			}
			s.Options.Ticket = strings.TrimSpace(ticket)
		}

		// Untuk mode FILTER interaktif (combined/separated), minta ticket number SETELAH database sudah dipilih
		// (urutan: profile -> pilih DB (filter/multi) -> input ticket -> display opsi).
		if !nonInteractive && (s.Options.Mode == consts.ModeCombined || s.Options.Mode == consts.ModeSeparated) && strings.TrimSpace(s.Options.Ticket) == "" {
			defaultTicket := fmt.Sprintf("bk-%d", time.Now().UnixNano())
			ticket, ticketErr := input.AskString("Ticket number", defaultTicket, func(ans interface{}) error {
				v, ok := ans.(string)
				if !ok {
					return fmt.Errorf("input tidak valid")
				}
				if strings.TrimSpace(v) == "" {
					return fmt.Errorf("ticket number tidak boleh kosong")
				}
				return nil
			})
			if ticketErr != nil {
				return nil, nil, ticketErr
			}
			s.Options.Ticket = strings.TrimSpace(ticket)
		}

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

		action, selErr := input.SelectSingleFromList(
			[]string{"Lanjutkan", "Ubah opsi", "Batalkan"},
			"Pilih aksi",
		)
		if selErr != nil {
			return nil, nil, selErr
		}

		switch action {
		case "Lanjutkan":
			if s.Options.Ticket == "" {
				ui.PrintError("Ticket number belum diisi.")
				ui.PrintError("Pilih 'Ubah opsi' untuk mengisi ticket, atau isi via flag --ticket.")
				ui.WaitForEnter("Tekan Enter untuk kembali ke opsi...")
				continue
			}
			// validasi minimal: jika encryption aktif, backup key wajib tersedia
			if s.Options.Encryption.Enabled && s.Options.Encryption.Key == "" {
				ui.PrintError("Encryption diaktifkan tapi backup key belum tersedia.")
				ui.PrintError("Isi via flag --backup-key, ENV SFDB_BACKUP_ENCRYPTION_KEY, atau pilih 'Ubah opsi' untuk input interaktif.")
				ui.WaitForEnter("Tekan Enter untuk kembali ke opsi...")
				continue
			}
			success = true
			return client, dbFiltered, nil
		case "Batalkan":
			return nil, nil, validation.ErrUserCancelled
		case "Ubah opsi":
			switch s.Options.Mode {
			case consts.ModeAll:
				if err := s.editBackupAllOptionsInteractive(ctx, &client, &customOutputDir); err != nil {
					return nil, nil, err
				}
			case consts.ModeSingle:
				if err := s.editBackupSingleOptionsInteractive(ctx, &client, &customOutputDir); err != nil {
					return nil, nil, err
				}
			case consts.ModePrimary:
				if err := s.editBackupPrimaryOptionsInteractive(ctx, &client, &customOutputDir); err != nil {
					return nil, nil, err
				}
			case consts.ModeSecondary:
				if err := s.editBackupSecondaryOptionsInteractive(ctx, &client, &customOutputDir); err != nil {
					return nil, nil, err
				}
			case consts.ModeCombined:
				if err := s.editBackupCombinedOptionsInteractive(ctx, &client, &customOutputDir); err != nil {
					return nil, nil, err
				}
			case consts.ModeSeparated:
				if err := s.editBackupSeparatedOptionsInteractive(ctx, &client, &customOutputDir); err != nil {
					return nil, nil, err
				}
			default:
				// Should never happen due to interactiveEditEnabled guard
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
