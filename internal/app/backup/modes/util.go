package modes

import "sfdbtools/internal/shared/consts"

// IsSingleModeVariant checks if mode is single.
// Catatan:
//   - Mode primary/secondary adalah mode batch (multi database) dan tidak boleh
//     mengandalkan prompt pemilihan DB, terutama pada mode background/daemon.
func IsSingleModeVariant(mode string) bool {
	return mode == consts.ModeSingle
}
