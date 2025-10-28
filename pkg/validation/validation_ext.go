package validation

import "strings"

// ProfileExt memastikan nama memiliki suffix .cnf.enc
func ProfileExt(name string) string {
	if strings.HasSuffix(name, ".cnf.enc") {
		return name
	}
	if strings.HasSuffix(name, ".cnf") {
		return name + ".enc"
	}
	return name + ".cnf.enc"
}
