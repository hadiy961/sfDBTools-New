package settings

import (
	"database/sql"
	"fmt"
	"github.com/fatih/color"
)

func (s *Service) saveSetting(db *sql.DB, key, value, category string) {
	_, err := db.Exec(`
		INSERT INTO app_settings (key, value, category, updated_at) 
		VALUES (?, ?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT(key) DO UPDATE SET 
			value = excluded.value, 
			category = excluded.category,
			updated_at = CURRENT_TIMESTAMP
	`, key, value, category)
	if err != nil {
		fmt.Println(color.RedString("Gagal menyimpan %s: %v", key, err))
	}
}

