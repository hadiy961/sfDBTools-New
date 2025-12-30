package setup

import (
	"context"
	"fmt"
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
func (s *Setup) PrepareBackupSession(ctx context.Context, headerTitle string, showOptions bool, genPaths PathGenerator) (client *database.Client, dbFiltered []string, err error) {
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
	} else {
		s.Options.Profile.DBInfo.HostName = serverHostname
		s.Log.Infof("menggunakan hostname dari server: %s", serverHostname)
	}

	if genPaths == nil {
		return nil, nil, fmt.Errorf("path generator tidak tersedia")
	}

	customOutputDir := s.Options.OutputDir

	// Interactive edit loop khusus untuk backup all (jika --force tidak diberikan)
	for {
		// Untuk mode ALL interaktif, minta ticket number terlebih dahulu agar muncul di Opsi Backup.
		if s.Options.Mode == consts.ModeAll && !showOptions && s.Options.Ticket == "" {
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

		// Jika user sudah set output-dir interaktif, pertahankan override ini.
		if customOutputDir != "" {
			s.Options.OutputDir = customOutputDir
		}

		// Default behavior untuk mode selain all tetap pakai confirm biasa.
		if s.Options.Mode != consts.ModeAll || showOptions {
			break
		}

		// Mode ALL + tidak --force: tampilkan opsi + beri pilihan edit.
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
			// Edit opsi-opsi yang diambil dari config.yaml (mode all)
			var v bool
			var sVal string

			// Ticket number (wajib)
			defaultTicket := s.Options.Ticket
			if defaultTicket == "" {
				// Gunakan format yang sama seperti BackupID (lihat metadata builder): bk-<unixnano>
				defaultTicket = fmt.Sprintf("bk-%d", time.Now().UnixNano())
			}
			if sVal, err = input.AskString("Ticket number", defaultTicket, nil); err != nil {
				return nil, nil, err
			}
			s.Options.Ticket = sVal

			if v, err = input.AskYesNo("Capture GTID?", s.Options.CaptureGTID); err != nil {
				return nil, nil, err
			}
			s.Options.CaptureGTID = v

			// Export user grants == !ExcludeUser
			if v, err = input.AskYesNo("Export user grants?", !s.Options.ExcludeUser); err != nil {
				return nil, nil, err
			}
			s.Options.ExcludeUser = !v

			if v, err = input.AskYesNo("Exclude system databases?", s.Options.Filter.ExcludeSystem); err != nil {
				return nil, nil, err
			}
			s.Options.Filter.ExcludeSystem = v

			if v, err = input.AskYesNo("Exclude empty databases?", s.Options.Filter.ExcludeEmpty); err != nil {
				return nil, nil, err
			}
			s.Options.Filter.ExcludeEmpty = v

			if v, err = input.AskYesNo("Exclude data (schema only)?", s.Options.Filter.ExcludeData); err != nil {
				return nil, nil, err
			}
			s.Options.Filter.ExcludeData = v

			if v, err = input.AskYesNo("Jalankan cleanup setelah backup?", s.Options.Cleanup.Enabled); err != nil {
				return nil, nil, err
			}
			s.Options.Cleanup.Enabled = v

			// Output directory override
			if sVal, err = input.AskString("Output directory", s.Options.OutputDir, nil); err != nil {
				return nil, nil, err
			}
			if sVal != "" {
				customOutputDir = sVal
				s.Options.OutputDir = sVal
			}

			// Custom filename base (tanpa ekstensi)
			if sVal, err = input.AskString("Custom filename (tanpa ekstensi, kosongkan untuk auto)", s.Options.File.Filename, nil); err != nil {
				return nil, nil, err
			}
			s.Options.File.Filename = sVal

			// Tetap munculkan opsi encrypt on/off meskipun config encryption.enabled=true
			if v, err = input.AskYesNo("Encrypt backup file?", s.Options.Encryption.Enabled); err != nil {
				return nil, nil, err
			}
			s.Options.Encryption.Enabled = v

			// Kecuali --backup-key sudah di provide (key sudah ada), maka tidak perlu prompt key.
			if s.Options.Encryption.Enabled && s.Options.Encryption.Key == "" {
				if sVal, err = input.AskPassword("Backup Key (required)", nil); err != nil {
					return nil, nil, err
				}
				if sVal == "" {
					ui.PrintError("Backup key tidak boleh kosong saat encryption aktif.")
					continue
				}
				s.Options.Encryption.Key = sVal
			}

			// Loop lagi untuk re-filter & re-generate preview sesuai opsi terbaru
			continue
		default:
			continue
		}
	}

	// Mode selain ALL atau jika --force: tetap pakai confirm lama
	if !showOptions {
		if proceed, askErr := display.NewOptionsDisplayer(s.Options).Display(); askErr != nil {
			return nil, nil, askErr
		} else if !proceed {
			return nil, nil, validation.ErrUserCancelled
		}
	}

	success = true
	return client, dbFiltered, nil
}
