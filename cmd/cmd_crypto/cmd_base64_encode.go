package cmdcrypto

import (
	"encoding/base64"
	"fmt"
	"os"
	"strings"

	"sfDBTools/internal/types"
	"sfDBTools/pkg/ui"

	"github.com/spf13/cobra"
)

var (
	b64eInput string
	b64eOut   string
)

// CmdBase64Encode melakukan base64 encode
var CmdBase64Encode = &cobra.Command{
	Use:     "base64encode",
	Aliases: []string{"b64e", "b64-enc"},
	Short:   "Base64-encode data dari --text atau stdin",
	Run: func(cmd *cobra.Command, args []string) {
		lg := types.Deps.Logger
		b, err := getInputBytes(b64eInput)
		if err != nil {
			lg.Errorf("Gagal membaca input: %v", err)
			return
		}
		enc := base64.StdEncoding.EncodeToString(b)
		ui.Headers("Base64 Encode")
		if strings.TrimSpace(b64eOut) == "" {
			ui.PrintSubHeader("Output text")
			fmt.Println(enc)
			ui.PrintDashedSeparator()
			return
		}
		if err := os.WriteFile(b64eOut, []byte(enc), 0644); err != nil {
			lg.Errorf("Gagal menulis file: %v", err)
			return
		}
		lg.Infof("✓ Base64 tersimpan: %s", b64eOut)
	},
}

func init() {
	CmdCryptoMain.AddCommand(CmdBase64Encode)
	CmdBase64Encode.Flags().StringVar(&b64eInput, "text", "", "Teks input (opsional, default baca stdin jika ada)")
	CmdBase64Encode.Flags().StringVarP(&b64eOut, "out", "o", "", "File output (opsional)")
}
