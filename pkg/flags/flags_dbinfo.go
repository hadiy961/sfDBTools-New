package flags

import (
	"github.com/spf13/cobra"
)

func FlagsDBInfo(cmd *cobra.Command) {
	cmd.Flags().StringP("host", "H", "", "Database host")
	cmd.Flags().IntP("port", "P", 0, "Database port")
	cmd.Flags().StringP("user", "U", "", "Database username")
	cmd.Flags().StringP("password", "p", "", "Database password")
}
