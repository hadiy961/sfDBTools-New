// File : internal/restore/setup_shared_profile.go
// Deskripsi : Helper profile dan koneksi database target untuk restore
// Author : Hadiyatna Muflihun
// Tanggal : 2025-12-30
// Last Modified : 2025-12-30

package restore

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"sfDBTools/internal/types"
	"sfDBTools/pkg/consts"
	"sfDBTools/pkg/database"
	profilehelper "sfDBTools/pkg/helper/profile"
)

func (s *Service) resolveTargetProfile(profileInfo *types.ProfileInfo, allowInteractive bool) error {
	loadedProfile, err := profilehelper.ResolveAndLoadProfile(profilehelper.ProfileLoadOptions{
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

	cfg := database.Config{
		Host:                 s.Profile.DBInfo.Host,
		Port:                 s.Profile.DBInfo.Port,
		User:                 s.Profile.DBInfo.User,
		Password:             s.Profile.DBInfo.Password,
		AllowNativePasswords: true,
		ParseTime:            true,
		Database:             "",
	}

	client, err := database.NewClient(ctx, cfg, 10*time.Second, 10, 5, 5*time.Minute)
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
