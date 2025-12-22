package backup

import "sfDBTools/pkg/consts"

func isProfileSelectionInteractiveMode(mode string) bool {
	switch mode {
	case consts.ModeSingle, consts.ModePrimary, consts.ModeSecondary, consts.ModeCombined, consts.ModeAll, consts.ModeSeparated:
		return true
	default:
		return false
	}
}
