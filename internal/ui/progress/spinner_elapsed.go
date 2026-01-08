// File : internal/ui/progress/spinner_elapsed.go
// Deskripsi : Helper untuk spinner dengan elapsed time tracking
// Author : Hadiyatna Muflihun
// Tanggal : 11 November 2025
// Last Modified : 5 Januari 2026

package progress

import (
	"fmt"
	"os"
	"sfdbtools/internal/shared/global"
	"sfdbtools/internal/shared/runtimecfg"
	"sfdbtools/internal/shared/spinnerguard"
	"sync"
	"time"

	"github.com/briandowns/spinner"
)

// elapsedSpinner adalah implementasi spinner dengan elapsed time tracking.
// Unexported karena API yang dipakai codebase adalah `progress.Spinner`.
type elapsedSpinner struct {
	spin      *spinner.Spinner
	done      chan bool
	message   string
	startTime time.Time
}

var (
	activeSpinner *elapsedSpinner
	activeMu      sync.Mutex
)

func init() {
	// Daftarkan suspender agar komponen lain (misalnya logger) bisa men-suspend spinner
	// tanpa perlu import paket ui (menghindari import cycle).
	spinnerguard.SetSuspender(RunWithSpinnerSuspended)
}

func newElapsedSpinner(message string) *elapsedSpinner {
	// Saat berjalan di background/daemon atau quiet, spinner bikin output tumpang tindih dengan logs.
	// Jadi dimatikan dan diganti no-op.
	if runtimecfg.IsQuiet() || runtimecfg.IsDaemon() {
		return &elapsedSpinner{message: message, startTime: time.Now()}
	}

	spin := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
	spin.Suffix = fmt.Sprintf(" %s...", message)
	// Write spinner to stderr so it doesn't interleave with stdout logs
	spin.Writer = os.Stderr

	s := &elapsedSpinner{
		spin:      spin,
		done:      make(chan bool, 1),
		message:   message,
		startTime: time.Now(),
	}

	return s
}

// Start memulai spinner dan elapsed time updater
func (s *elapsedSpinner) Start() {
	if s == nil || s.spin == nil {
		return
	}
	s.spin.Start()

	// Register as active spinner
	activeMu.Lock()
	activeSpinner = s
	activeMu.Unlock()

	// Goroutine untuk update spinner dengan elapsed time
	go func() {
		for {
			select {
			case <-s.done:
				return
			default:
				elapsed := time.Since(s.startTime)
				s.spin.Suffix = fmt.Sprintf(" %s... (%s)", s.message, global.FormatDuration(elapsed))
				time.Sleep(500 * time.Millisecond)
			}
		}
	}()
}

// Stop menghentikan spinner dan elapsed time updater
func (s *elapsedSpinner) Stop() {
	if s == nil || s.spin == nil {
		return
	}
	// Deregister active spinner if this is it
	activeMu.Lock()
	if activeSpinner == s {
		activeSpinner = nil
	}
	activeMu.Unlock()

	s.done <- true // Stop elapsed time updater
	s.spin.Stop()
}

// UpdateMessage mengubah message spinner tanpa restart
func (s *elapsedSpinner) UpdateMessage(message string) {
	if s == nil || s.spin == nil {
		return
	}
	s.message = message
	elapsed := time.Since(s.startTime)
	s.spin.Suffix = fmt.Sprintf(" %s... (%s)", message, global.FormatDuration(elapsed))
}

// SuspendAndRun stops the spinner, runs action, then restarts the spinner preserving elapsed time
func (s *elapsedSpinner) SuspendAndRun(action func()) {
	if s == nil || s.spin == nil {
		action()
		return
	}
	// Stop current spinner and goroutine
	s.done <- true
	s.spin.Stop()

	// Run the action while spinner is stopped
	action()

	// Restart spinner
	s.done = make(chan bool, 1)
	s.spin.Start()

	// Register as active spinner
	activeMu.Lock()
	activeSpinner = s
	activeMu.Unlock()

	// Restart elapsed updater
	go func() {
		for {
			select {
			case <-s.done:
				return
			default:
				elapsed := time.Since(s.startTime)
				s.spin.Suffix = fmt.Sprintf(" %s... (%s)", s.message, global.FormatDuration(elapsed))
				time.Sleep(500 * time.Millisecond)
			}
		}
	}()
}

// RunWithSpinnerSuspended runs action with any active spinner suspended during the action
func RunWithSpinnerSuspended(action func()) {
	activeMu.Lock()
	s := activeSpinner
	activeMu.Unlock()
	if s == nil {
		action()
		return
	}
	s.SuspendAndRun(action)
}
