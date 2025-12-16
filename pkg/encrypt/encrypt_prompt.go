package encrypt

import (
	"os"
	"sfDBTools/pkg/input"

	"github.com/AlecAivazis/survey/v2"
)

// PromptPassword mendapatkan password dari environment variable atau user prompt.
// Parameter:
//   - envVar: nama environment variable (misal: "SFDB_ENCRYPTION_KEY")
//   - promptMsg: pesan yang ditampilkan saat meminta input user
//
// Return:
//   - password: password yang didapat
//   - source: "env" atau "prompt"
//   - error: jika gagal mendapatkan password
func PromptPassword(envVar, promptMsg string) (password, source string, err error) {
	if pw := os.Getenv(envVar); pw != "" {
		return pw, "env", nil
	}
	pw, err := input.AskPassword(promptMsg, survey.Required)
	return pw, "prompt", err
}
