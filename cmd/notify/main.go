package notifycmd

import "github.com/spf13/cobra"

var CmdNotifyMain = &cobra.Command{
	Use:   "notify",
	Short: "Manajemen dan testing notification service",
	Run:   func(cmd *cobra.Command, args []string) { cmd.Help() },
}

func init() {
	CmdNotifyMain.AddCommand(CmdNotifyTest)
	CmdNotifyMain.AddCommand(CmdNotifyGetChatID)
	CmdNotifyMain.AddCommand(CmdNotifyDeleteWebhook)
}
