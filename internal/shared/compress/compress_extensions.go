package compress

import "sfdbtools/internal/shared/consts"

func SupportedCompressionExtensions() []string {
	return []string{consts.ExtGzip, consts.ExtZstd, consts.ExtXz, consts.ExtZlib}
}
