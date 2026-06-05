package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"sfdbtools/internal/shared/database"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var importCmd = &cobra.Command{
	Use:   "import [filename]",
	Short: "Import settings and profiles from JSON",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		filename := args[0]
		jsonData, err := os.ReadFile(filename)
		if err != nil {
			return err
		}

		var data map[string]interface{}
		if err := json.Unmarshal(jsonData, &data); err != nil {
			return err
		}

		db, err := database.GetSQLite()
		if err != nil {
			return err
		}

		tx, _ := db.Begin()
		defer tx.Rollback()

		// 1. Settings
		if settings, ok := data["settings"].(map[string]interface{}); ok {
			for k, v := range settings {
				if m, ok := v.(map[string]interface{}); ok {
					_, _ = tx.Exec("INSERT OR REPLACE INTO app_settings (key, value, category) VALUES (?, ?, ?)", k, m["value"], m["category"])
				}
			}
		}

		// 2. Profiles
		if profiles, ok := data["profiles"].([]interface{}); ok {
			for _, p := range profiles {
				if m, ok := p.(map[string]interface{}); ok {
					_, _ = tx.Exec("INSERT OR REPLACE INTO profiles (name, encrypted_data, account_code) VALUES (?, ?, ?)", m["name"], m["data"], m["code"])
				}
			}
		}

		if err := tx.Commit(); err != nil {
			return err
		}

		fmt.Println(color.GreenString("Successfully imported from %s", filename))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(importCmd)
}
