package encrypt

import (
	"os"
	"sfDBTools/pkg/input"
	"sfDBTools/pkg/ui"

	"github.com/AlecAivazis/survey/v2"
)

// EncryptionPrompt mendapatkan password enkripsi dari env atau prompt.
// Mengembalikan password, sumbernya ("env" atau "prompt"), dan error jika ada.
func EncryptionPrompt(promptMessage string, env string) (string, string, error) {
	// Cek environment variable SFDB_ENCRYPTION_KEY
	// Jika ada, gunakan itu
	// Jika tidak, minta user memasukkan password
	if password := os.Getenv(env); password != "" {
		return password, "env", nil
	} else {
		ui.PrintSubHeader("Authentication Required")
		ui.PrintWarning("Environment variable " + env + " tidak ditemukan atau kosong. Silakan atur " + env + " atau ketik password.")
	}
	// Minta user memasukkan password
	// Validator: tidak boleh kosong
	// return input.AskPassword(promptMessage, survey.Required)
	EncryptionPassword, err := input.AskPassword("Encryption Password", survey.Required)

	return EncryptionPassword, "prompt", err

}
