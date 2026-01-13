// File : internal/cli/runner/runner.go
// Deskripsi : Runner untuk standarisasi eksekusi command + handling deps nil
// Author : Hadiyatna Muflihun
// Tanggal : 14 Januari 2026
// Last Modified : 14 Januari 2026

package runner

import (
	"fmt"
	"os"

	appdeps "sfdbtools/internal/cli/deps"

	"github.com/spf13/cobra"
)

const depsMissingMsg = "✗ Dependencies tidak tersedia. Pastikan aplikasi diinisialisasi dengan benar."

// Run mengeksekusi action dengan guard dependency dan logging error terpusat.
// Catatan: fungsi ini sengaja TIDAK panic dan TIDAK melakukan os.Exit.
func Run(cmd *cobra.Command, action func() error) {
	if appdeps.Deps == nil {
		printErr(cmd, depsMissingMsg)
		return
	}
	if action == nil {
		return
	}
	if err := action(); err != nil {
		logOrPrintErr(cmd, err)
	}
}

// RunResolved menjalankan resolve -> action(T) dengan handling error terpusat.
func RunResolved[T any](cmd *cobra.Command, resolve func() (T, error), action func(T) error) {
	if appdeps.Deps == nil {
		printErr(cmd, depsMissingMsg)
		return
	}
	if resolve == nil || action == nil {
		return
	}
	v, err := resolve()
	if err != nil {
		logOrPrintErr(cmd, err)
		return
	}
	if err := action(v); err != nil {
		logOrPrintErr(cmd, err)
	}
}

func logOrPrintErr(cmd *cobra.Command, err error) {
	if err == nil {
		return
	}
	if appdeps.Deps != nil && appdeps.Deps.Logger != nil {
		appdeps.Deps.Logger.Error(err.Error())
		return
	}
	printErr(cmd, err.Error())
}

func printErr(cmd *cobra.Command, msg string) {
	if cmd != nil {
		cmd.PrintErrln(msg)
		return
	}
	fmt.Fprintln(os.Stderr, msg)
}
