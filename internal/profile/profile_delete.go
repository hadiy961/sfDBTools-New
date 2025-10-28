package profile

import (
	"fmt"
	"path/filepath"
	"sfDBTools/pkg/fsops"
	"sfDBTools/pkg/input"
	"sfDBTools/pkg/ui"
	"sfDBTools/pkg/validation"
)

// PromptDeleteProfile - Menampilkan prompt konfirmasi penghapusan profil.
func (s *Service) PromptDeleteProfile() error {
	ui.Headers("Delete Database Configurations")

	configDir := s.Config.ConfigDir.DatabaseProfile

	files, err := fsops.ReadDirFiles(configDir)
	if err != nil {
		return fmt.Errorf("gagal membaca direktori konfigurasi: %w", err)
	}
	// filter hanya *.cnf.enc (gunakan helper ensureConfigExt untuk konsistensi)
	filtered := make([]string, 0, len(files))
	for _, f := range files {
		if validation.ProfileExt(f) == f { // artinya sudah ber-ekstensi .cnf.enc
			filtered = append(filtered, f)
		}
	}
	if len(filtered) == 0 {
		ui.PrintInfo("Tidak ada file konfigurasi untuk dihapus.")
		return nil
	}

	idxs, err := input.ShowMultiSelect("Pilih file konfigurasi yang akan dihapus:", filtered)
	if err != nil {
		return validation.HandleInputError(err)
	}

	selected := make([]string, 0, len(idxs))
	for _, i := range idxs {
		if i >= 1 && i <= len(filtered) {
			selected = append(selected, filepath.Join(configDir, filtered[i-1]))
		}
	}

	if len(selected) == 0 {
		ui.PrintInfo("Tidak ada file terpilih untuk dihapus.")
		return nil
	}

	ok, err := input.AskYesNo(fmt.Sprintf("Anda yakin ingin menghapus %d file?", len(selected)), false)
	if err != nil {
		return validation.HandleInputError(err)
	}
	if !ok {
		ui.PrintInfo("Penghapusan dibatalkan oleh pengguna.")
		return nil
	}

	for _, p := range selected {
		if err := fsops.RemoveFile(p); err != nil {
			s.Log.Error(fmt.Sprintf("Gagal menghapus file %s: %v", p, err))
			ui.PrintError(fmt.Sprintf("Gagal menghapus file %s: %v", p, err))
		} else {
			s.Log.Info(fmt.Sprintf("Berhasil menghapus: %s", p))
			ui.PrintSuccess(fmt.Sprintf("Berhasil menghapus: %s", p))
		}
	}

	return nil
}
