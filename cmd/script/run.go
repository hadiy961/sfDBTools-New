package scriptcmd

import (
	"fmt"
	"os"
	"path/filepath"
	"sfDBTools/internal/app/script"
	"sfDBTools/internal/cli/deps"
	"sfDBTools/internal/cli/flags"
	"sfDBTools/internal/cli/parsing"
	"sfDBTools/pkg/input"
	"sort"
	"strings"

	"github.com/spf13/cobra"
)

var CmdScriptRun = &cobra.Command{
	Use:   "run",
	Short: "Jalankan .sftools",
	Long:  "Mendekripsi .sftools ke folder temporary, lalu menjalankan entrypoint-nya dengan bash.",
	Example: `
# Run bundle
sfdbtools script run --file scripts/sftr_sf7_main_menu/main_menu.sftools --key "mypassword"

# Pakai flag pendek
sfdbtools script run --file scripts/sftr_sf7_main_menu/main_menu.sftools -k "mypassword"

# Key bisa dari env SFDB_SCRIPT_KEY
SFDB_SCRIPT_KEY="mypassword" sfdbtools script run --file scripts/sftr_sf7_main_menu/main_menu.sftools

# Pass-through args ke script (gunakan --)
sfdbtools script run -f tes -k "mypassword" -- --version
sfdbtools script run -f tes -k "mypassword" -- training ticket tes
sfdbtools script run -f tes -k "mypassword" -- --mode=1
sfdbtools script run -f tes -k "mypassword" -- --mode 1
`,
	Run: func(cmd *cobra.Command, args []string) {
		opts := parsing.ParsingScriptRunOptions(cmd)
		opts.Args = args
		if strings.TrimSpace(opts.FilePath) == "" {
			p, err := selectSFToolsFileInteractive()
			if err != nil {
				deps.Deps.Logger.Error(err.Error())
				return
			}
			opts.FilePath = p

			// Interaktif: jika user belum provide args via CLI (`-- ...`), tanyakan args opsional.
			if len(opts.Args) == 0 {
				line, err := input.PromptString("Masukkan args untuk script (opsional, pisahkan dengan spasi; kosongkan jika tidak ada)")
				if err != nil {
					deps.Deps.Logger.Error(err.Error())
					return
				}
				line = strings.TrimSpace(line)
				if line != "" {
					opts.Args = strings.Fields(line)
				}
			}
		} else if cmd.Flags().Changed("file") {
			configuredDir := ""
			if deps.Deps != nil && deps.Deps.Config != nil {
				configuredDir = strings.TrimSpace(deps.Deps.Config.Script.BundleOutputDir)
			}
			opts.FilePath = normalizeSFToolsFlagPath(opts.FilePath, configuredDir)
		}
		if err := script.RunBundle(opts); err != nil {
			deps.Deps.Logger.Error(err.Error())
		}
	},
}

func normalizeSFToolsFlagPath(fileArg string, configuredDir string) string {
	p := strings.TrimSpace(fileArg)
	if p == "" {
		return p
	}

	// If user only provides a name (no path separator), resolve against configured bundle dir.
	hasSep := strings.Contains(p, "/") || strings.Contains(p, "\\")
	if !hasSep && strings.TrimSpace(configuredDir) != "" {
		p = filepath.Join(filepath.Clean(configuredDir), p)
	}

	// Auto-append .sftools if extension is missing.
	if strings.TrimSpace(filepath.Ext(p)) == "" {
		p = p + ".sftools"
	}
	return p
}

func selectSFToolsFileInteractive() (string, error) {
	configuredDir := ""
	if deps.Deps != nil && deps.Deps.Config != nil {
		configuredDir = strings.TrimSpace(deps.Deps.Config.Script.BundleOutputDir)
	}

	// Default UX: tampilkan daftar file dari configured dir jika ada.
	if configuredDir != "" {
		items, err := listSFToolsFiles(configuredDir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: %v\n", err)
			return input.SelectFileInteractive(".", "Cari file .sftools yang akan dijalankan", []string{".sftools"})
		}
		if len(items) > 0 {
			const browseLabel = "🔎 Cari manual (browse)"
			const inputLabel = "⌨️  Input path manual"

			options := append([]string{}, items...)
			options = append(options, browseLabel, inputLabel)

			selected, err := selectSingleFromListWithDefault(options, fmt.Sprintf("Pilih file .sftools (default dari config: %s)", configuredDir), items[0])
			if err != nil {
				return "", err
			}
			switch selected {
			case browseLabel:
				return input.SelectFileInteractive(configuredDir, "Cari file .sftools yang akan dijalankan", []string{".sftools"})
			case inputLabel:
				for {
					p, err := input.PromptString("Masukkan path file .sftools")
					if err != nil {
						return "", err
					}
					p = strings.TrimSpace(p)
					if p == "" {
						continue
					}
					p = normalizeSFToolsFlagPath(p, configuredDir)
					if _, err := os.Stat(p); err != nil {
						fmt.Printf("File tidak ditemukan: %s\n", p)
						continue
					}
					return p, nil
				}
			default:
				return selected, nil
			}
		}
	}

	// Fallback: browse dari current dir (tetap bisa ketik path bebas).
	return input.SelectFileInteractive(".", "Pilih file .sftools yang akan dijalankan", []string{".sftools"})
}

func selectSingleFromListWithDefault(items []string, message string, defaultVal string) (string, error) {
	return input.SelectSingleFromListWithDefault(items, message, defaultVal)
}

func listSFToolsFiles(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("gagal membaca directory bundle_output_dir (%s): %w", dir, err)
	}

	var files []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if strings.EqualFold(filepath.Ext(name), ".sftools") {
			files = append(files, filepath.Join(dir, name))
		}
	}
	sort.Strings(files)
	return files, nil
}

func init() {
	CmdScriptMain.AddCommand(CmdScriptRun)
	flags.AddScriptRunFlags(CmdScriptRun)
}
