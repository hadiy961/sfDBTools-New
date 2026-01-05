// File : cmd/cmd_update.go
// Deskripsi : Sub-command untuk update binary sfDBTools dari GitHub Release terbaru
// Author : Hadiyatna Muflihun
// Tanggal : 2026-01-05
// Last Modified : 2026-01-05
package cmd

import (
	"context"
	"fmt"
	"os"
	"time"

	"sfDBTools/internal/app/autoupdate"
	appdeps "sfDBTools/internal/cli/deps"

	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update sfDBTools ke versi terbaru",
	Long: `Update sfDBTools dengan mengambil release terbaru dari GitHub.

Catatan:
- Saat ini hanya mendukung binary linux/amd64 (sesuai workflow release).
- Jika terpasang di /usr/bin, kemungkinan perlu menjalankan dengan sudo.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		force, _ := cmd.Flags().GetBool("force")

		ctx, cancel := context.WithTimeout(cmd.Context(), 60*time.Second)
		defer cancel()

		opts := autoupdate.OptionsFromEnv()
		opts.Force = force
		opts.ReExec = false

		// Logger optional: update harus tetap bisa berjalan walau config belum ada.
		var log autoupdate.Logger
		if appdeps.Deps != nil {
			log = appdeps.Deps.Logger
		}

		err := autoupdate.UpdateIfNeeded(ctx, log, opts)
		if err != nil {
			// Tampilkan error yang user-friendly.
			fmt.Fprintln(os.Stderr, err.Error())
			return err
		}
		return nil
	},
}

func init() {
	updateCmd.Flags().Bool("force", false, "Paksa update walau versi sama")
}
