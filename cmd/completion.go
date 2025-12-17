// File : cmd/cmd_completion.go
// Deskripsi : Perintah untuk menghasilkan script shell completion (bash/zsh/fish/powershell)
// Author : Hadiyatna Muflihun
// Tanggal : 2025-11-14
// Last Modified : 2025-11-14
package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

// completionCmd menghasilkan script auto-completion untuk berbagai shell
var completionCmd = &cobra.Command{
	Use:                   "completion [bash|zsh|fish|powershell]",
	Short:                 "Generate shell completion scripts",
	Args:                  cobra.ExactValidArgs(1),
	ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
	DisableFlagsInUseLine: true,
	SilenceUsage:          true,
	SilenceErrors:         true,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Jangan menampilkan hal lain selain script completion
		// Root PersistentPreRunE akan di-skip via pengecekan di root (cmd_root.go)

		switch args[0] {
		case "bash":
			return rootCmd.GenBashCompletionV2(os.Stdout, true)
		case "zsh":
			return rootCmd.GenZshCompletion(os.Stdout)
		case "fish":
			return rootCmd.GenFishCompletion(os.Stdout, true)
		case "powershell":
			return rootCmd.GenPowerShellCompletionWithDesc(os.Stdout)
		}
		return nil
	},
}
