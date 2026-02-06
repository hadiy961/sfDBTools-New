package dbusercmd

import "github.com/spf13/cobra"

// CmdDBUserMain adalah perintah induk (parent command) untuk operasi user/grants.
var CmdDBUserMain = &cobra.Command{
	Use:     "db-user",
	Aliases: []string{"dbuser", "user", "users"},
	Short:   "Suite tools untuk export/apply user dan grants",
	Long: `Perintah 'db-user' digunakan untuk mengelola user accounts dan grants.

Saat ini tersedia:
  - export: export user+grants menjadi file SQL
  - apply : apply (restore) file SQL user+grants ke target server

Ke depan dapat diperluas untuk manajemen user (create/drop/rotate password, dll).`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func init() {
	CmdDBUserMain.AddCommand(CmdDBUserExport)
	CmdDBUserMain.AddCommand(CmdDBUserApply)
}
