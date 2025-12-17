package input

import (
	"fmt"
	"strings"
)

// AskTicket prompts the user for a ticket number with validation.
func AskTicket(action string) (string, error) {
	prompt := fmt.Sprintf("Masukkan ticket number untuk %s request : ", action)
	return AskString(prompt, "", func(ans interface{}) error {
		str, ok := ans.(string)
		if !ok {
			return fmt.Errorf("input tidak valid")
		}
		if strings.TrimSpace(str) == "" {
			return fmt.Errorf("ticket number tidak boleh kosong")
		}
		return nil
	})
}
