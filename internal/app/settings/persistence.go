package settings

import (
	"database/sql"
	"fmt"
	"github.com/fatih/color"
)

func (s *Service) saveSetting(db *sql.DB, key, value, category string) {
	_, err := db.Exec(`
		INSERT INTO app_settings (key, value, category) 
		VALUES (?, ?, ?)
		ON CONFLICT(key) DO UPDATE SET value = excluded.value, category = excluded.category
	`, key, value, category)
	if err != nil {
		fmt.Println(color.RedString("Gagal menyimpan %s: %v", key, err))
	}
}

