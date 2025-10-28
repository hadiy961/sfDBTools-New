package helper

import "strings"

// helper.TrimProfileSuffix menghapus suffix .cnf.enc dari nama jika ada.
func TrimProfileSuffix(name string) string {
	return strings.TrimSuffix(strings.TrimSuffix(name, ".enc"), ".cnf")
}
