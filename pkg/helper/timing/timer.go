package timing

import "time"

// Timer adalah simple wrapper untuk time tracking.
type Timer struct {
	startTime time.Time
}

// NewTimer membuat timer baru yang langsung start.
func NewTimer() *Timer {
	return &Timer{startTime: time.Now()}
}

// StartTime mengembalikan waktu mulai timer.
func (t *Timer) StartTime() time.Time {
	return t.startTime
}

// Elapsed mengembalikan durasi sejak timer di-start.
func (t *Timer) Elapsed() time.Duration {
	return time.Since(t.startTime)
}
