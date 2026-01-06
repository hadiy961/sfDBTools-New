// File : cmd/cmd_version.go
// Deskripsi : Sub-command untuk menampilkan versi aplikasi
// Author : Hadiyatna Muflihun
// Tanggal : 3 Oktober 2024
// Last Modified : 2 Januari 2026
package cmd

import (
	"fmt"
	"sfdbtools/pkg/version"

	"github.com/spf13/cobra"
)

// versionCmd adalah sub-command untuk menampilkan versi aplikasi
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Menampilkan versi aplikasi",
	Run: func(cmd *cobra.Command, args []string) {
		// Cetak ke stdout agar mudah dipakai di pipeline/script.
		fmt.Println(version.String())
	},
}
