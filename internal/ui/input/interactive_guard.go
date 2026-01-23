package input

import (
	"fmt"
	"os"

	"sfdbtools/internal/shared/runtimecfg"
	"sfdbtools/internal/shared/validation"

	"github.com/mattn/go-isatty"
)

func ensureInteractiveAllowed() error {
	if runtimecfg.IsQuiet() {
		return fmt.Errorf("mode non-interaktif (--quiet): input interaktif tidak tersedia: %w", validation.ErrNonInteractive)
	}
	// survey butuh TTY supaya input/output rapi.
	if !isatty.IsTerminal(os.Stdin.Fd()) || !isatty.IsTerminal(os.Stdout.Fd()) {
		return fmt.Errorf("stdin/stdout bukan TTY: input interaktif tidak tersedia: %w", validation.ErrNonInteractive)
	}
	return nil
}
