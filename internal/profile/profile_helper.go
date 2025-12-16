package profile

import (
	"fmt"
	"os"
	"path/filepath"
	"sfDBTools/internal/types"
	"sfDBTools/pkg/helper"
	"sfDBTools/pkg/profilehelper"
	"sfDBTools/pkg/validation"
	"strings"
)

// buildFileName menormalkan input (menghapus suffix jika ada) lalu memastikan suffix .cnf.enc
// Tujuan: menghindari duplikasi suffix saat user sudah mengetikkan nama dengan ekstensi.
func buildFileName(name string) string {
	return validation.ProfileExt(helper.TrimProfileSuffix(strings.TrimSpace(name)))
}

// filePathInConfigDir membangun absolute path di dalam config dir untuk nama file konfigurasi yang diberikan.
func (s *Service) filePathInConfigDir(name string) string {
	cfgDir := s.Config.ConfigDir.DatabaseProfile
	return filepath.Join(cfgDir, buildFileName(name))
}

// loadSnapshotFromPath membaca file terenkripsi, mencoba dekripsi (jika kunci tersedia/di-prompt),
// parse nilai penting, dan mengisi s.OriginalProfileInfo beserta metadata.
func (s *Service) loadSnapshotFromPath(absPath string) error {
	info, err := profilehelper.ResolveAndLoadProfile(profilehelper.ProfileLoadOptions{
		ConfigDir:      s.Config.ConfigDir.DatabaseProfile,
		ProfilePath:    absPath,
		ProfileKey:     s.ProfileInfo.EncryptionKey,
		RequireProfile: true,
	})
	if err != nil {
		s.fillOriginalInfoFromMeta(absPath, types.ProfileInfo{})
		return err
	}
	s.fillOriginalInfoFromMeta(absPath, *info)
	return nil
}

// fillOriginalInfoFromMeta mengisi OriginalProfileInfo dengan metadata file dan nilai koneksi yang tersedia
func (s *Service) fillOriginalInfoFromMeta(absPath string, info types.ProfileInfo) {
	var fileSizeStr string
	var lastMod = info.LastModified
	if fi, err := os.Stat(absPath); err == nil {
		fileSizeStr = fmt.Sprintf("%d bytes", fi.Size())
		lastMod = fi.ModTime()
	}

	s.OriginalProfileInfo = &types.ProfileInfo{
		Path:         absPath,
		Name:         helper.TrimProfileSuffix(filepath.Base(absPath)),
		DBInfo:       info.DBInfo,
		Size:         fileSizeStr,
		LastModified: lastMod,
	}
}

// formatConfigToINI mengubah struct DBConfigInfo menjadi format string INI.
func (s *Service) formatConfigToINI() string {
	// [client] adalah header standar untuk file my.cnf
	// Ini memastikan kompatibilitas dengan banyak tools command-line MySQL/MariaDB
	content := `[client]
host=%s
port=%d
user=%s
password=%s
`
	return fmt.Sprintf(content, s.ProfileInfo.DBInfo.Host, s.ProfileInfo.DBInfo.Port, s.ProfileInfo.DBInfo.User, s.ProfileInfo.DBInfo.Password)
}
