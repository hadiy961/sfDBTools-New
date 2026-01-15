// File : internal/restore/setup_shared_profile.go
// Deskripsi : Helper profile dan koneksi database target untuk restore
// Author : Hadiyatna Muflihun
// Tanggal : 30 Desember 2025
// Last Modified : 14 Januari 2026
package restore

import (
	"context"
	"fmt"
	"path/filepath"
	profileconn "sfdbtools/internal/app/profile/connection"
	"sfdbtools/internal/app/profile/helpers/loader"
	"sfdbtools/internal/domain"
	"sfdbtools/internal/shared/consts"
)

func (s *Service) resolveTargetProfile(profileInfo *domain.ProfileInfo, allowInteractive bool) error {
	loadedProfile, err := loader.ResolveAndLoadProfile(loader.ProfileLoadOptions{
		ConfigDir:         s.Config.ConfigDir.DatabaseProfile,
		ProfilePath:       profileInfo.Path,
		ProfileKey:        profileInfo.EncryptionKey,
		EnvProfilePath:    consts.ENV_TARGET_PROFILE,
		EnvProfileKey:     consts.ENV_TARGET_PROFILE_KEY,
		RequireProfile:    true,
		ProfilePurpose:    "target",
		AllowInteractive:  allowInteractive,
		InteractivePrompt: "Pilih target profile untuk restore:",
	})
	if err != nil {
		return fmt.Errorf("gagal load profile: %w", err)
	}

	*profileInfo = *loadedProfile
	s.Profile = loadedProfile
	s.Log.Infof("Target profile: %s (%s:%d)",
		filepath.Base(profileInfo.Path),
		s.Profile.DBInfo.Host,
		s.Profile.DBInfo.Port)

	return nil
}

func (s *Service) connectToTargetDatabase(ctx context.Context) error {
	s.Log.Info("Menghubungkan ke database target...")

	client, err := profileconn.ConnectWithProfile(s.Profile, consts.DefaultInitialDatabase)
	if err != nil {
		return fmt.Errorf("koneksi database target gagal: %w", err)
	}

	s.TargetClient = client

	serverHostname, err := client.GetServerHostname(ctx)
	if err != nil {
		s.Log.Warnf("gagal mendapatkan hostname dari server: %v, menggunakan dari config", err)
		if s.Profile != nil && s.Profile.DBInfo.HostName == "" {
			s.Profile.DBInfo.HostName = s.Profile.DBInfo.Host
		}
	} else {
		if s.Profile != nil {
			s.Profile.DBInfo.HostName = serverHostname
		}
		s.Log.Infof("menggunakan hostname dari server: %s", serverHostname)
	}
	s.Log.Info("Koneksi ke database target berhasil")

	return nil
}
