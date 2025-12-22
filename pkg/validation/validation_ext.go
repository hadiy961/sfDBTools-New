package validation

import (
	"sfDBTools/pkg/consts"
	"strings"
)

// ProfileExt memastikan nama memiliki suffix .cnf.enc
func ProfileExt(name string) string {
	if strings.HasSuffix(name, consts.ExtCnfEnc) {
		return name
	}
	if strings.HasSuffix(name, consts.ExtCnf) {
		return name + consts.ExtEnc
	}
	return name + consts.ExtCnfEnc
}
