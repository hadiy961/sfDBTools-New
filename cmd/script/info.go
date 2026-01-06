package scriptcmd

import (
	"fmt"
	"sfdbtools/internal/app/script"
	"sfdbtools/internal/cli/deps"
	"sfdbtools/internal/cli/flags"
	"sfdbtools/internal/cli/parsing"
	"strings"

	"github.com/spf13/cobra"
)

var CmdScriptInfo = &cobra.Command{
	Use:   "info",
	Short: "Tampilkan info bundle .sftools",
	Long:  "Mendekripsi bundle .sftools dan membaca manifest untuk menampilkan entrypoint dan metadata.",
	Example: `
# Info bundle

sfdbtools script info -f /etc/sfDBTools/scripts/tes.sftools -k "mypassword"
`,
	Run: func(cmd *cobra.Command, args []string) {
		opts := parsing.ParsingScriptInfoOptions(cmd)
		if strings.TrimSpace(opts.FilePath) == "" {
			p, err := selectSFToolsFileInteractive()
			if err != nil {
				deps.Deps.Logger.Error(err.Error())
				return
			}
			opts.FilePath = p
		} else if cmd.Flags().Changed("file") {
			configuredDir := ""
			if deps.Deps != nil && deps.Deps.Config != nil {
				configuredDir = strings.TrimSpace(deps.Deps.Config.Script.BundleOutputDir)
			}
			opts.FilePath = normalizeSFToolsFlagPath(opts.FilePath, configuredDir)
		}

		info, err := script.GetBundleInfo(opts)
		if err != nil {
			deps.Deps.Logger.Error(err.Error())
			return
		}

		fmt.Printf("File      : %s\n", opts.FilePath)
		fmt.Printf("Version   : %d\n", info.Version)
		fmt.Printf("Mode      : %s\n", info.Mode)
		fmt.Printf("Entrypoint: %s\n", info.Entrypoint)
		if strings.TrimSpace(info.RootDir) != "" {
			fmt.Printf("Folder    : %s\n", info.RootDir)
		}
		fmt.Printf("CreatedAt : %s\n", info.CreatedAt)
		fmt.Printf("Files     : %d\n", info.FileCount)

		if info.Mode == "bundle" {
			fmt.Println("Scripts   :")
			if len(info.Scripts) == 0 {
				fmt.Println("  (tidak ada file .sh ditemukan)")
			} else {
				for _, s := range info.Scripts {
					if strings.TrimSpace(info.RootDir) != "" {
						fmt.Printf("  - %s/%s\n", info.RootDir, s)
					} else {
						fmt.Printf("  - %s\n", s)
					}
				}
			}
		}
	},
}

func init() {
	CmdScriptMain.AddCommand(CmdScriptInfo)
	flags.AddScriptInfoFlags(CmdScriptInfo)
}
