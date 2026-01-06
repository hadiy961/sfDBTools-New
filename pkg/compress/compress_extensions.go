package compress

import "sfdbtools/pkg/consts"

func SupportedCompressionExtensions() []string {
	return []string{consts.ExtGzip, consts.ExtZstd, consts.ExtXz, consts.ExtZlib}
}
