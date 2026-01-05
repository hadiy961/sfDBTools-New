package scriptcmd

import (
	"path/filepath"
	"sfDBTools/internal/cli/deps"
	"sfDBTools/internal/cli/flags"
	"sfDBTools/internal/cli/parsing"
	"sfDBTools/internal/script"
	"sfDBTools/pkg/input"
	"strings"

	"github.com/spf13/cobra"
)

var CmdScriptExtract = &cobra.Command{
	Use:   "extract",
	Short: "Extract bundle .sftools ke folder",
	Long:  "Mendekripsi .sftools lalu mengekstrak isinya ke out-dir. Command ini butuh proteksi ganda: key + password aplikasi.",
	Example: `
# Extract bundle untuk diedit
sfdbtools script extract -f /etc/sfDBTools/scripts/tes.sftools -o ./tes_extracted -k "mypassword"
`,
	Run: func(cmd *cobra.Command, args []string) {
		opts := parsing.ParsingScriptExtractOptions(cmd)

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

		if strings.TrimSpace(opts.OutDir) == "" {
			base := strings.TrimSuffix(filepath.Base(opts.FilePath), filepath.Ext(opts.FilePath))
			defaultOut := "./" + base + "_extracted"
			out, err := input.AskString("Output directory untuk hasil extract", defaultOut, nil)
			if err != nil {
				deps.Deps.Logger.Error(err.Error())
				return
			}
			opts.OutDir = strings.TrimSpace(out)
		}

		if err := script.ExtractBundle(opts); err != nil {
			deps.Deps.Logger.Error(err.Error())
			return
		}

		absOut, _ := filepath.Abs(opts.OutDir)
		deps.Deps.Logger.Infof("✓ Extract selesai. Output: %s", absOut)
	},
}

func init() {
	CmdScriptMain.AddCommand(CmdScriptExtract)
	flags.AddScriptExtractFlags(CmdScriptExtract)
}
