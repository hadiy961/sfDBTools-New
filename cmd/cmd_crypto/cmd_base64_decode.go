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
	b64dInput string
	b64dOut   string
)

// CmdBase64Decode melakukan base64 decode
var CmdBase64Decode = &cobra.Command{
	Use:     "base64decode",
	Aliases: []string{"b64d", "b64-dec"},
	Short:   "Base64-decode data dari --data atau stdin",
	Run: func(cmd *cobra.Command, args []string) {
		lg := types.Deps.Logger
		in, err := getInputString(b64dInput)
		if err != nil {
			lg.Errorf("Gagal membaca input: %v", err)
			return
		}

		b, err := base64.StdEncoding.DecodeString(strings.TrimSpace(in))
		if err != nil {
			lg.Errorf("Input bukan base64 yang valid: %v", err)
			return
		}
		ui.Headers("Base64 Decode")
		ui.PrintSubHeader("Output Text : ")
		if strings.TrimSpace(b64dOut) == "" {
			fmt.Println(string(b))
			ui.PrintDashedSeparator()
			return
		}
		if err := os.WriteFile(b64dOut, b, 0644); err != nil {
			lg.Errorf("Gagal menulis file: %v", err)
			return
		}
		lg.Infof("✓ Decode tersimpan: %s", b64dOut)
	},
}

func init() {
	CmdCryptoMain.AddCommand(CmdBase64Decode)
	CmdBase64Decode.Flags().StringVar(&b64dInput, "data", "", "Input base64 (opsional, default baca stdin jika ada)")
	CmdBase64Decode.Flags().StringVarP(&b64dOut, "out", "o", "", "File output (opsional)")
}
