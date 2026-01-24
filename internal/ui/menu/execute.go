// File : internal/ui/menu/execute.go
// Deskripsi : Eksekusi Cobra command dari menu interaktif
// Author : Hadiyatna Muflihun
// Tanggal : 23 Januari 2026
// Last Modified : 23 Januari 2026

package menu

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func commandArgsFromRoot(root *cobra.Command, target *cobra.Command) ([]string, error) {
	if root == nil || target == nil {
		return nil, fmt.Errorf("root/target command nil")
	}

	// Bangun args dari parent chain: root -> ... -> target
	parts := []string{}
	for c := target; c != nil && c != root; c = c.Parent() {
		parts = append([]string{c.Name()}, parts...)
	}
	if len(parts) == 0 {
		return nil, fmt.Errorf("tidak bisa membangun args untuk command: %s", target.Name())
	}
	return parts, nil
}

func executeCommand(root *cobra.Command, target *cobra.Command) error {
	args, err := commandArgsFromRoot(root, target)
	if err != nil {
		return err
	}

	// Simpan dan set os.Args supaya logging (argv) akurat.
	oldOSArgs := os.Args
	os.Args = append([]string{oldOSArgs[0]}, args...)
	defer func() { os.Args = oldOSArgs }()

	// Jalankan command terpilih.
	root.SetArgs(args)
	defer root.SetArgs([]string{})

	return root.Execute()
}
