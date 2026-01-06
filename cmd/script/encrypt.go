package scriptcmd

import (
	"fmt"
	"os"
	"path/filepath"
	"sfdbtools/internal/app/script"
	"sfdbtools/internal/cli/deps"
	"sfdbtools/internal/cli/flags"
	"sfdbtools/internal/cli/parsing"
	"sfdbtools/internal/ui/prompt"
	"strings"

	"github.com/spf13/cobra"
)

var CmdScriptEncrypt = &cobra.Command{
	Use:   "encrypt",
	Short: "Encrypt satu folder script menjadi .sftools",
	Long:  "Membundle file entrypoint (mode single) atau seluruh isi foldernya (mode bundle) menjadi satu file .sftools terenkripsi.",
	Example: `
# Encrypt bundle dari entrypoint
sfdbtools script encrypt --file scripts/sftr_sf7_main_menu/main_menu.sh --key "mypassword"

# Pakai flag pendek
sfdbtools script encrypt --file scripts/sftr_sf7_main_menu/main_menu.sh -k "mypassword"

# Mode single (hanya 1 file script, tanpa scan folder)
sfdbtools script encrypt -f scripts/tools/hello.sh -m single -k "mypassword"

# Key bisa dari env SFDB_SCRIPT_KEY
SFDB_SCRIPT_KEY="mypassword" sfdbtools script encrypt --file scripts/sftr_sf7_main_menu/main_menu.sh
`,
	Run: func(cmd *cobra.Command, args []string) {
		opts := parsing.ParsingScriptEncryptOptions(cmd)

		// Interaktif: pilih mode jika user tidak mengisi --mode/-m.
		if !cmd.Flags().Changed("mode") {
			selectedMode, _, err := prompt.SelectOne("Pilih mode encrypt", []string{"bundle", "single"}, -1)
			if err != nil {
				deps.Deps.Logger.Error(err.Error())
				return
			}
			opts.Mode = selectedMode
		}

		if strings.TrimSpace(opts.FilePath) == "" {
			p, err := prompt.SelectFile(".", "Pilih entrypoint script (.sh) yang akan dienkripsi", []string{".sh"})
			if err != nil {
				deps.Deps.Logger.Error(err.Error())
				return
			}
			opts.FilePath = p
		}

		sourceDir := filepath.Clean(filepath.Dir(opts.FilePath))
		entryBase := filepath.Base(opts.FilePath)
		entryExt := strings.ToLower(filepath.Ext(entryBase))
		baseName := entryBase
		if entryExt != "" {
			baseName = strings.TrimSuffix(entryBase, entryExt)
		}
		defaultOutputName := baseName

		mode := strings.ToLower(strings.TrimSpace(opts.Mode))
		if mode == "" {
			mode = "bundle"
		}

		// Tentukan default output directory.
		outDir := sourceDir
		outDirFromConfig := false
		if deps.Deps != nil && deps.Deps.Config != nil {
			cfgOutDir := strings.TrimSpace(deps.Deps.Config.Script.BundleOutputDir)
			if cfgOutDir != "" {
				outDir = filepath.Clean(cfgOutDir)
				outDirFromConfig = true
			}
		}

		// Default output .sftools bisa diatur lewat YAML:
		// script:
		//   bundle_output_dir: "/path/to/output"

		// Interaktif: custom output filename (tanpa ekstensi), jika user belum set --output.
		if !cmd.Flags().Changed("output") {
			outputName, err := prompt.AskText(
				"Nama output file (tanpa ekstensi, otomatis .sftools)",
				prompt.WithDefault(defaultOutputName),
				prompt.WithValidator(func(ans interface{}) error {
					s, _ := ans.(string)
					s = strings.TrimSpace(s)
					if s == "" {
						return fmt.Errorf("nama file tidak boleh kosong")
					}
					if strings.Contains(s, string(os.PathSeparator)) || strings.Contains(s, "/") || strings.Contains(s, "\\") {
						return fmt.Errorf("nama file tidak boleh mengandung path")
					}
					return nil
				}),
			)
			if err != nil {
				deps.Deps.Logger.Error(err.Error())
				return
			}
			outputName = strings.TrimSpace(outputName)
			outputFilename := outputName
			if !strings.HasSuffix(strings.ToLower(outputFilename), ".sftools") {
				outputFilename = outputFilename + ".sftools"
			}
			opts.OutputPath = filepath.Join(outDir, outputFilename)
		}

		// Interaktif: pilihan hapus sumber, jika flag --delete-source tidak di-set.
		deleteSource := opts.DeleteSource
		if !cmd.Flags().Changed("delete-source") {
			targetLabel := sourceDir
			if mode == "single" {
				targetLabel = opts.FilePath
			}
			wantDelete, err := prompt.Confirm(
				fmt.Sprintf("Hapus sumber setelah encrypt berhasil? (%s)", targetLabel),
				false,
			)
			if err != nil {
				deps.Deps.Logger.Error(err.Error())
				return
			}
			deleteSource = wantDelete
		}

		// Jika mode bundle dan user ingin hapus sumber namun output masih berada di folder yang sama,
		// pindahkan default output ke current working dir (hanya jika output tidak berasal dari config).
		if deleteSource {
			if mode == "bundle" && !outDirFromConfig && filepath.Clean(filepath.Dir(opts.OutputPath)) == sourceDir {
				deps.Deps.Logger.Warn("Output berada di folder yang akan dihapus; output diarahkan ke folder kerja saat ini agar aman.")
				opts.OutputPath = filepath.Join(".", filepath.Base(opts.OutputPath))
			}

			confirm, err := prompt.AskText(
				"Ketik DELETE untuk konfirmasi penghapusan sumber",
				prompt.WithValidator(func(ans interface{}) error {
					s, _ := ans.(string)
					if strings.TrimSpace(s) != "DELETE" {
						return fmt.Errorf("konfirmasi tidak cocok")
					}
					return nil
				}),
			)
			if err != nil {
				deps.Deps.Logger.Error(err.Error())
				return
			}
			_ = confirm
		}

		if err := script.EncryptBundle(opts); err != nil {
			deps.Deps.Logger.Error(err.Error())
			return
		}

		if deleteSource {
			// single: hapus file entrypoint saja
			if mode == "single" {
				absEntry, err := filepath.Abs(opts.FilePath)
				if err != nil {
					deps.Deps.Logger.Error(fmt.Sprintf("gagal resolve absolute path sumber: %v", err))
					return
				}
				absEntry = filepath.Clean(absEntry)
				if absEntry == "/" || absEntry == "." || absEntry == "" {
					deps.Deps.Logger.Error("refuse delete: sumber tidak aman")
					return
				}
				if err := os.Remove(absEntry); err != nil {
					deps.Deps.Logger.Error(fmt.Sprintf("gagal menghapus sumber: %v", err))
					return
				}
				deps.Deps.Logger.Infof("✓ Sumber dihapus: %s", absEntry)
			} else {
				// bundle: hapus folder sumber
				absSource, err := filepath.Abs(sourceDir)
				if err != nil {
					deps.Deps.Logger.Error(fmt.Sprintf("gagal resolve absolute path sumber: %v", err))
					return
				}
				absSource = filepath.Clean(absSource)
				if absSource == "/" || absSource == "." || absSource == "" {
					deps.Deps.Logger.Error("refuse delete: sumber tidak aman")
					return
				}

				absOut, _ := filepath.Abs(opts.OutputPath)
				absOut = filepath.Clean(absOut)
				if absOut == absSource || strings.HasPrefix(absOut, absSource+string(os.PathSeparator)) {
					deps.Deps.Logger.Error("refuse delete: output berada di dalam folder sumber")
					return
				}

				if err := os.RemoveAll(absSource); err != nil {
					deps.Deps.Logger.Error(fmt.Sprintf("gagal menghapus sumber: %v", err))
					return
				}
				deps.Deps.Logger.Infof("✓ Sumber dihapus: %s", absSource)
			}
		}
		deps.Deps.Logger.Infof("✓ Bundle terenkripsi dibuat dari: %s", opts.FilePath)
		if strings.TrimSpace(opts.OutputPath) != "" {
			fmt.Printf("Output: %s\n", opts.OutputPath)
		} else {
			fmt.Println("(Output: otomatis .sftools di folder yang sama)")
		}
	},
}

func init() {
	CmdScriptMain.AddCommand(CmdScriptEncrypt)
	flags.AddScriptEncryptFlags(CmdScriptEncrypt)
}
