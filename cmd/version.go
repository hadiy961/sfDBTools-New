// File : cmd/cmd_version.go
// Deskripsi : Sub-command untuk menampilkan versi aplikasi
// Author : Hadiyatna Muflihun
// Tanggal : 2024-10-03
// Last Modified : 2026-01-02
package cmd

import (
	"fmt"
	"sfDBTools/pkg/version"

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
