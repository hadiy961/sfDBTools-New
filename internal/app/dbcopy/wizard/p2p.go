// File : internal/app/dbcopy/wizard/p2p.go
// Deskripsi : Wizard interaktif untuk db-copy p2p (primary -> primary)
// Author : Hadiyatna Muflihun
// Tanggal : 26 Januari 2026
// Last Modified : 27 Januari 2026

package wizard

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"sfdbtools/internal/app/backup/selection"
	"sfdbtools/internal/app/dbcopy"
	"sfdbtools/internal/app/dbcopy/model"
	"sfdbtools/internal/app/profile/helpers/loader"
	"sfdbtools/internal/domain"
	"sfdbtools/internal/shared/consts"
	"sfdbtools/internal/ui/print"
	"sfdbtools/internal/ui/prompt"
	"sfdbtools/internal/ui/table"
)

func selectAndLoadProfileInteractive(configDir string, promptText string) (*domain.ProfileInfo, error) {
	return loader.ResolveAndLoadProfile(loader.ProfileLoadOptions{
		ConfigDir:         strings.TrimSpace(configDir),
		ProfilePath:       "",
		ProfileKey:        "",
		EnvProfilePath:    "",
		EnvProfileKey:     "",
		RequireProfile:    true,
		ProfilePurpose:    "database",
		AllowInteractive:  true,
		InteractivePrompt: promptText,
	})
}

// RunP2PWizard melengkapi opsi P2P dengan prompt interaktif.
// Fokus wizard ini: memilih target profile (jika belum diisi) dan memilih source DB (primary).
func RunP2PWizard(ctx context.Context, configDir string, svc *dbcopy.Service, opts *model.P2POptions) (*model.P2POptions, error) {
	if svc == nil {
		return nil, fmt.Errorf("service nil")
	}
	if opts == nil {
		return nil, fmt.Errorf("options nil")
	}

	// Safety: p2p selalu pre-backup target sebelum overwrite.
	opts.PrebackupTarget = true

	print.PrintAppHeader("Database Copy - Primary to Primrary (p2p)")
	print.PrintSubHeader("Wizard db-copy p2p")

	// Ticket wajib untuk audit.
	if strings.TrimSpace(opts.Ticket) == "" {
		t, err := prompt.AskTicket("db-copy p2p")
		if err != nil {
			return nil, err
		}
		opts.Ticket = strings.TrimSpace(t)
	}

	// Opsi global (interactive UX). Kalau user sudah set via flags/env, jangan prompt.
	// Default diambil dari nilai opts (mengikuti default flag Cobra).
	if !opts.SkipConfirm {
		cont, err := prompt.Confirm("Lanjutkan jika langkah non-kritis gagal? (continue-on-error)", opts.ContinueOnError)
		if err != nil {
			return nil, err
		}
		opts.ContinueOnError = cont

		exData, err := prompt.Confirm("Backup schema saja (exclude data)?", opts.ExcludeData)
		if err != nil {
			return nil, err
		}
		opts.ExcludeData = exData

		incDmart, err := prompt.Confirm("Ikut copy companion database (_dmart) jika ada?", opts.IncludeDmart)
		if err != nil {
			return nil, err
		}
		opts.IncludeDmart = incDmart

		if strings.TrimSpace(opts.Workdir) == "" {
			wd, err := prompt.AskText("Workdir (opsional, kosong = default)")
			if err != nil {
				return nil, err
			}
			opts.Workdir = strings.TrimSpace(wd)
		}
	}

	// 0) Source profile (interactive) jika belum diisi.
	var srcProfile *domain.ProfileInfo
	if strings.TrimSpace(opts.SourceProfile) == "" {
		p, err := selectAndLoadProfileInteractive(configDir, "Pilih file konfigurasi database sumber")
		if err != nil {
			return nil, err
		}
		srcProfile = p
		opts.SourceProfile = strings.TrimSpace(p.Path)
		opts.SourceProfileKey = strings.TrimSpace(p.EncryptionKey)
	} else if strings.TrimSpace(opts.SourceProfileKey) == "" {
		k, err := prompt.AskPassword("Masukkan source profile key", nil)
		if err != nil {
			return nil, err
		}
		opts.SourceProfileKey = strings.TrimSpace(k)
	}

	// 1) Target profile wajib (tidak boleh default sama dengan source).
	if strings.TrimSpace(opts.TargetProfile) == "" {
		p, err := selectAndLoadProfileInteractive(configDir, "Pilih file konfigurasi database target")
		if err != nil {
			return nil, err
		}
		opts.TargetProfile = strings.TrimSpace(p.Path)
		// Simpan key agar executor tidak prompt ulang.
		if strings.TrimSpace(opts.TargetProfileKey) == "" {
			opts.TargetProfileKey = strings.TrimSpace(p.EncryptionKey)
		}
	} else if strings.TrimSpace(opts.TargetProfileKey) == "" {
		// Target key dibutuhkan jika target profile diisi via flag/env tapi key belum tersedia.
		k, err := prompt.AskPassword("Masukkan target profile key", nil)
		if err != nil {
			return nil, err
		}
		opts.TargetProfileKey = strings.TrimSpace(k)
	}

	// Normalisasi untuk perbandingan sederhana (tidak memaksa abs path).
	srcClean := filepath.Clean(strings.TrimSpace(opts.SourceProfile))
	tgtClean := filepath.Clean(strings.TrimSpace(opts.TargetProfile))
	if srcClean != "" && tgtClean != "" && strings.EqualFold(srcClean, tgtClean) {
		return nil, fmt.Errorf("db-copy p2p ditolak: source-profile dan target-profile tidak boleh sama")
	}

	// 2) Pilih database source (primary).
	// Jika user sudah provide source-db, gunakan itu.
	if strings.TrimSpace(opts.SourceDB) == "" {
		// Opsional: user bisa input client-code untuk memperkecil kandidat.
		cc, err := prompt.AskText("Client code (opsional, untuk filter list primary)")
		if err != nil {
			return nil, err
		}
		cc = strings.TrimSpace(cc)
		if cc != "" {
			opts.ClientCode = cc
		}

		// Reuse profile hasil seleksi jika ada, agar tidak prompt ulang.
		if srcProfile == nil {
			sp, err := svc.LoadProfile(opts.SourceProfile, opts.SourceProfileKey, consts.ENV_SOURCE_PROFILE, consts.ENV_SOURCE_PROFILE_KEY, false, "source")
			if err != nil {
				return nil, err
			}
			srcProfile = sp
		}
		srcClient, err := svc.ConnectDB(srcProfile)
		if err != nil {
			return nil, err
		}
		defer srcClient.Close()

		all, err := srcClient.GetDatabaseList(ctx)
		if err != nil {
			return nil, fmt.Errorf("gagal mengambil daftar database source: %w", err)
		}

		candidates := selection.FilterCandidatesByMode(all, consts.ModePrimary)
		candidates = selection.FilterCandidatesByClientCode(candidates, opts.ClientCode)
		if len(candidates) == 0 {
			if strings.TrimSpace(opts.ClientCode) != "" {
				return nil, fmt.Errorf("tidak ada database primary yang match client-code '%s'", opts.ClientCode)
			}
			return nil, fmt.Errorf("tidak ada database primary yang tersedia di source")
		}

		// Jika hanya satu kandidat, auto pilih.
		if len(candidates) == 1 {
			opts.SourceDB = candidates[0]
		} else {
			selected, _, err := prompt.SelectOne("Pilih database primary source:", candidates, -1)
			if err != nil {
				return nil, err
			}
			opts.SourceDB = selected
		}
	}

	// P2P: target DB selalu sama dengan source DB.
	if strings.TrimSpace(opts.TargetDB) != "" && !strings.EqualFold(strings.TrimSpace(opts.TargetDB), strings.TrimSpace(opts.SourceDB)) {
		return nil, fmt.Errorf("untuk db-copy p2p, --target-db harus sama dengan --source-db (target akan selalu = source)")
	}
	opts.TargetDB = strings.TrimSpace(opts.SourceDB)

	// Ringkasan rencana eksekusi + konfirmasi (hanya interaktif).
	if !opts.SkipConfirm {
		print.PrintSubHeader("Ringkasan")
		wd := strings.TrimSpace(opts.Workdir)
		if wd == "" {
			wd = "(default)"
		}

		rows := [][]string{
			{"Ticket", strings.TrimSpace(opts.Ticket)},
			{"Source profile", filepath.Base(strings.TrimSpace(opts.SourceProfile))},
			{"Target profile", filepath.Base(strings.TrimSpace(opts.TargetProfile))},
			{"Database", strings.TrimSpace(opts.SourceDB)},
			{"Exclude data", fmt.Sprintf("%v", opts.ExcludeData)},
			{"Include _dmart", fmt.Sprintf("%v", opts.IncludeDmart)},
			{"Continue on error", fmt.Sprintf("%v", opts.ContinueOnError)},
			{"Workdir", wd},
			{"Catatan", "p2p selalu pre-backup target (wajib)"},
		}
		table.Render([]string{"Field", "Value"}, rows)

		ok, err := prompt.Confirm("Lanjutkan eksekusi db-copy p2p?", true)
		if err != nil {
			return nil, err
		}
		if !ok {
			return nil, fmt.Errorf("dibatalkan oleh user")
		}
	}

	return opts, nil
}
