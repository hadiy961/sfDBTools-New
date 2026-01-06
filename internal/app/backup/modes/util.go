package modes

import "sfdbtools/pkg/consts"

// IsSingleModeVariant checks if mode is single/primary/secondary.
func IsSingleModeVariant(mode string) bool {
	return mode == consts.ModeSingle || mode == consts.ModePrimary || mode == consts.ModeSecondary
}
