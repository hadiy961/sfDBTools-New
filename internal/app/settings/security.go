package settings

import (
	"database/sql"
	"fmt"
	"sfdbtools/internal/crypto"
	"sfdbtools/internal/shared/database"
	"sfdbtools/internal/ui/prompt"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/fatih/color"
)

func (s *Service) SecurityMenu(db *sql.DB) {
	for {
		options := []string{
			"1. Rotate Master Encryption Key",
			"0. Back",
		}
		
		sel, _, err := prompt.SelectOne("Security Settings:", options, 0)
		if err != nil || strings.Contains(sel, "Back") {
			return
		}

		switch sel[0:1] {
		case "1":
			s.RotateMasterKey(db)
		}
	}
}

func (s *Service) RotateMasterKey(db *sql.DB) {
	fmt.Println(color.CyanString("\n--- Master Key Rotation ---"))
	fmt.Println("This will re-encrypt all stored database profiles with a new key.")
	
	var oldKey, newKey string
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().Title("Current Master Key").EchoMode(huh.EchoModePassword).Value(&oldKey),
			huh.NewInput().Title("New Master Key").EchoMode(huh.EchoModePassword).Value(&newKey),
		),
	)

	if err := form.Run(); err != nil {
		return
	}

	if oldKey == "" || newKey == "" {
		fmt.Println(color.RedString("Keys cannot be empty!"))
		return
	}

	// 1. Load all profiles
	rows, err := db.Query("SELECT name, encrypted_data FROM profiles")
	if err != nil {
		fmt.Println(color.RedString("Failed to load profiles: %v", err))
		return
	}
	defer rows.Close()

	type profile struct {
		name string
		data string
	}
	var profiles []profile
	for rows.Next() {
		var p profile
		if err := rows.Scan(&p.name, &p.data); err != nil {
			fmt.Println(color.RedString("Error scanning profile: %v", err))
			return
		}
		profiles = append(profiles, p)
	}

	// 2. Re-encrypt
	fmt.Println("Re-encrypting profiles...")
	tx, err := db.Begin()
	if err != nil {
		fmt.Println(color.RedString("Failed to start transaction: %v", err))
		return
	}
	defer tx.Rollback()

	for _, p := range profiles {
		// Decrypt with old key
		decrypted, err := crypto.DecryptData([]byte(p.data), []byte(oldKey))
		if err != nil {
			fmt.Println(color.RedString("Failed to decrypt profile '%s'. Wrong old key?", p.name))
			return
		}

		// Encrypt with new key
		encrypted, err := crypto.EncryptData(decrypted, []byte(newKey))
		if err != nil {
			fmt.Println(color.RedString("Failed to encrypt profile '%s': %v", p.name, err))
			return
		}

		_, err = tx.Exec("UPDATE profiles SET encrypted_data = ? WHERE name = ?", string(encrypted), p.name)
		if err != nil {
			fmt.Println(color.RedString("Failed to update profile '%s': %v", p.name, err))
			return
		}
	}

	if err := tx.Commit(); err != nil {
		fmt.Println(color.RedString("Failed to commit changes: %v", err))
		return
	}

	fmt.Println(color.GreenString("\n[SUCCESS] Master key rotated successfully!"))
	fmt.Println("Make sure to update your SFDB_ENCRYPTION_KEY or key file accordingly.")
}
