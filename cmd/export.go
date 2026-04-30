package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"sfdbtools/internal/shared/database"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var exportCmd = &cobra.Command{
	Use:   "export [filename]",
	Short: "Export local settings and profiles to JSON",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		filename := "sfdbtools_export.json"
		if len(args) > 0 {
			filename = args[0]
		}

		db, err := database.GetSQLite()
		if err != nil {
			return err
		}

		data := make(map[string]interface{})

		// 1. Settings
		rows, _ := db.Query("SELECT key, value, category FROM app_settings")
		settings := make(map[string]interface{})
		for rows.Next() {
			var k, v, c string
			rows.Scan(&k, &v, &c)
			settings[k] = map[string]string{"value": v, "category": c}
		}
		rows.Close()
		data["settings"] = settings

		// 2. Profiles
		pRows, _ := db.Query("SELECT name, encrypted_data, account_code FROM profiles")
		profiles := make([]map[string]string, 0)
		for pRows.Next() {
			var n, d, c string
			pRows.Scan(&n, &d, &c)
			profiles = append(profiles, map[string]string{"name": n, "data": d, "code": c})
		}
		pRows.Close()
		data["profiles"] = profiles

		jsonData, err := json.MarshalIndent(data, "", "  ")
		if err != nil {
			return err
		}

		if err := os.WriteFile(filename, jsonData, 0600); err != nil {
			return err
		}

		fmt.Println(color.GreenString("Successfully exported to %s", filename))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(exportCmd)
}
