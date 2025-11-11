// File : pkg/ui/ui_spinner.go
// Deskripsi : Helper untuk spinner dengan elapsed time tracking
// Author : Hadiyatna Muflihun
// Tanggal : 11 November 2025
// Last Modified : 11 November 2025

package ui

import (
	"fmt"
	"sfDBTools/pkg/global"
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

// NewSpinnerWithElapsed membuat spinner baru dengan elapsed time tracking
func NewSpinnerWithElapsed(message string) *SpinnerWithElapsed {
	spin := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
	spin.Suffix = fmt.Sprintf(" %s...", message)

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
	s.spin.Start()

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
	s.done <- true // Stop elapsed time updater
	s.spin.Stop()
}

// UpdateMessage mengubah message spinner tanpa restart
func (s *SpinnerWithElapsed) UpdateMessage(message string) {
	s.message = message
	elapsed := time.Since(s.startTime)
	s.spin.Suffix = fmt.Sprintf(" %s... (%s)", message, global.FormatDuration(elapsed))
}
