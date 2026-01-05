// File : pkg/ui/ui_spinner.go
// Deskripsi : Helper untuk spinner dengan elapsed time tracking
// Author : Hadiyatna Muflihun
// Tanggal : 11 November 2025
// Last Modified : 4 Januari 2026

package ui

import (
	"fmt"
	"os"
	"sfDBTools/pkg/global"
	"sfDBTools/pkg/runtimecfg"
	"sfDBTools/pkg/spinnerguard"
	"sync"
	"time"

	"github.com/briandowns/spinner"
)

// SpinnerWithElapsed adalah wrapper untuk spinner dengan elapsed time tracking
type SpinnerWithElapsed struct {
	spin      *spinner.Spinner
	done      chan bool
	message   string
	startTime time.Time
}

var (
	activeSpinner *SpinnerWithElapsed
	activeMu      sync.Mutex
)

func init() {
	// Daftarkan suspender agar komponen lain (misalnya logger) bisa men-suspend spinner
	// tanpa perlu import paket ui (menghindari import cycle).
	spinnerguard.SetSuspender(RunWithSpinnerSuspended)
}

// NewSpinnerWithElapsed membuat spinner baru dengan elapsed time tracking
func NewSpinnerWithElapsed(message string) *SpinnerWithElapsed {
	// Saat berjalan di background/daemon atau quiet, spinner bikin output tumpang tindih dengan logs.
	// Jadi dimatikan dan diganti no-op.
	if runtimecfg.IsQuiet() || runtimecfg.IsDaemon() {
		return &SpinnerWithElapsed{message: message, startTime: time.Now()}
	}

	spin := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
	spin.Suffix = fmt.Sprintf(" %s...", message)
	// Write spinner to stderr so it doesn't interleave with stdout logs
	spin.Writer = os.Stderr

	s := &SpinnerWithElapsed{
		spin:      spin,
		done:      make(chan bool, 1),
		message:   message,
		startTime: time.Now(),
	}

	return s
}

// Start memulai spinner dan elapsed time updater
func (s *SpinnerWithElapsed) Start() {
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
func (s *SpinnerWithElapsed) Stop() {
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
func (s *SpinnerWithElapsed) UpdateMessage(message string) {
	if s == nil || s.spin == nil {
		return
	}
	s.message = message
	elapsed := time.Since(s.startTime)
	s.spin.Suffix = fmt.Sprintf(" %s... (%s)", message, global.FormatDuration(elapsed))
}

// SuspendAndRun stops the spinner, runs action, then restarts the spinner preserving elapsed time
func (s *SpinnerWithElapsed) SuspendAndRun(action func()) {
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
