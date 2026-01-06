// File : internal/ui/progress/spinner.go
// Deskripsi : Spinner/progress untuk UI facade
// Author : Hadiyatna Muflihun
// Tanggal : 5 Januari 2026
// Last Modified : 5 Januari 2026

package progress

import (
	"fmt"
	"os"
	"sfdbtools/pkg/runtimecfg"
	"time"

	"github.com/briandowns/spinner"
)

// Spinner adalah wrapper spinner yang expose API sederhana.
type Spinner struct {
	noElapsed *spinner.Spinner
	elapsed   *elapsedSpinner
}

// NewSpinner membuat spinner tanpa elapsed time.
func NewSpinner(label string) *Spinner {
	if runtimecfg.IsQuiet() || runtimecfg.IsDaemon() {
		return &Spinner{}
	}

	sp := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
	sp.Suffix = fmt.Sprintf(" %s...", label)
	sp.Writer = os.Stderr

	return &Spinner{noElapsed: sp}
}

// NewSpinnerWithElapsed membuat spinner dengan elapsed time tracking.
func NewSpinnerWithElapsed(label string) *Spinner {
	return &Spinner{elapsed: newElapsedSpinner(label)}
}

func (s *Spinner) Start() {
	if s == nil {
		return
	}
	if s.elapsed != nil {
		s.elapsed.Start()
		return
	}
	if s.noElapsed != nil {
		s.noElapsed.Start()
	}
}

func (s *Spinner) Stop() {
	if s == nil {
		return
	}
	if s.elapsed != nil {
		s.elapsed.Stop()
		return
	}
	if s.noElapsed != nil {
		s.noElapsed.Stop()
	}
}

func (s *Spinner) Update(label string) {
	if s == nil {
		return
	}
	if s.elapsed != nil {
		s.elapsed.UpdateMessage(label)
		return
	}
	if s.noElapsed != nil {
		s.noElapsed.Suffix = fmt.Sprintf(" %s...", label)
	}
}
