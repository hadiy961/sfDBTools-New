// File : internal/profile/profile_entry.go
// Deskripsi : Entry point untuk profile command execution
// Author : Hadiyatna Muflihun
// Tanggal : 2025-12-16
// Last Modified : 2025-12-16

package profile

import (
	"sfDBTools/internal/types"
)

// ExecuteProfileCommand adalah entry point untuk profile command execution
func (s *Service) ExecuteProfileCommand(config types.ProfileEntryConfig) error {
	// Log prefix untuk tracking
	if config.LogPrefix != "" {
		s.Log.Infof("[%s] Memulai profile operation dengan mode: %s", config.LogPrefix, config.Mode)
	}

	// Tampilkan options jika diminta
	if config.ShowOptions {
		s.displayProfileOptions()
	}

	// Jalankan profile operation berdasarkan mode
	var err error
	switch config.Mode {
	case "create":
		err = s.CreateProfile()
	case "show":
		err = s.ShowProfile()
	case "edit":
		err = s.EditProfile()
	case "delete":
		err = s.PromptDeleteProfile()
	default:
		return ErrInvalidProfileMode
	}

	return err
}

// displayProfileOptions menampilkan konfigurasi profile yang akan dijalankan (jika diperlukan)
func (s *Service) displayProfileOptions() {
	// This method can be implemented later if needed for showing options before execution
	// For now, profile operations show their own UI in their respective methods
	s.Log.Debug("displayProfileOptions called (not implemented)")
}
