package settings

import (
	"fmt"
	"sfdbtools/internal/app/initialize"
	"sfdbtools/internal/shared/database"
	"sfdbtools/internal/ui/prompt"

	"github.com/fatih/color"
)

func (s *Service) ResetSettings() {
	confirm, _ := prompt.Confirm("Apakah Anda yakin ingin mereset semua pengaturan ke default?", false)
	if !confirm {
		return
	}

	// Logika reset memanggil SetupDefaultSettings dari modul initialize
	svc := initialize.NewService(s.deps)
	
	// TRUNCATE TABLE agar benar-benar ter-reset
	db, _ := database.GetSQLite()
	db.Exec("DELETE FROM app_settings")

	clientCode := s.config.General.ClientCode
	if clientCode == "" {
		clientCode = "DEFAULT" // Fallback jika client code kosong di config
	}

	if err := svc.SetupDefaultSettings(clientCode); err != nil {
		fmt.Println(color.RedString("\n[ERROR] Gagal mereset pengaturan: %v", err))
		return
	}

	fmt.Println(color.GreenString("\n[SUCCESS] Semua pengaturan telah dikembalikan ke nilai default."))
	fmt.Println("Beberapa perubahan mungkin memerlukan restart aplikasi untuk diterapkan sepenuhnya.")
}
