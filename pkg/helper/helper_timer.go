// File : pkg/helper/helper_timer.go
// Deskripsi : Helper untuk time tracking dengan automatic duration calculation
// Author : Hadiyatna Muflihun
// Tanggal : 11 November 2025
// Last Modified : 11 November 2025

package helper

import (
	"fmt"
	"time"
)

// Timer adalah simple wrapper untuk time tracking
type Timer struct {
	startTime time.Time
}

// NewTimer membuat timer baru yang langsung start
func NewTimer() *Timer {
	return &Timer{
		startTime: time.Now(),
	}
}

// Elapsed mengembalikan durasi sejak timer di-start
func (t *Timer) Elapsed() time.Duration {
	return time.Since(t.startTime)
}

// Reset me-reset timer ke waktu sekarang
func (t *Timer) Reset() {
	t.startTime = time.Now()
}

// StartTime mengembalikan waktu start timer
func (t *Timer) StartTime() time.Time {
	return t.startTime
}

// ElapsedFormatted mengembalikan durasi dalam format string yang readable
// Format: "1h 23m 45s" atau "2m 30s" atau "45s"
func (t *Timer) ElapsedFormatted() string {
	d := t.Elapsed()

	if d < time.Minute {
		return d.Round(time.Second).String()
	}

	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	s := int(d.Seconds()) % 60

	if h > 0 {
		if s > 0 {
			return fmt.Sprintf("%dh %dm %ds", h, m, s)
		}
		return fmt.Sprintf("%dh %dm", h, m)
	}

	if s > 0 {
		return fmt.Sprintf("%dm %ds", m, s)
	}
	return fmt.Sprintf("%dm", m)
}
