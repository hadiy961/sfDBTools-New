package compress

import "sfDBTools/pkg/consts"

func SupportedCompressionExtensions() []string {
	return []string{consts.ExtGzip, consts.ExtZstd, consts.ExtXz, consts.ExtZlib}
}
