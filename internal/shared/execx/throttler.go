package execx

import (
	"context"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// ParseSpeed converts a string like "10MB/s" or "500KB" to bytes per second.
func ParseSpeed(s string) (int64, error) {
	s = strings.ToUpper(strings.TrimSpace(s))
	if s == "" {
		return 0, nil
	}

	// Remove "/S" if exists
	s = strings.TrimSuffix(s, "/S")

	re := regexp.MustCompile(`^(\d+)\s*(B|KB|MB|GB|TB)?$`)
	matches := re.FindStringSubmatch(s)
	if len(matches) != 3 {
		return 0, fmt.Errorf("format kecepatan tidak valid: %s", s)
	}

	val, _ := strconv.ParseInt(matches[1], 10, 64)
	unit := matches[2]

	switch unit {
	case "KB":
		val *= 1024
	case "MB":
		val *= 1024 * 1024
	case "GB":
		val *= 1024 * 1024 * 1024
	case "TB":
		val *= 1024 * 1024 * 1024 * 1024
	}

	return val, nil
}

// ThrottledReader membatasi kecepatan baca dari io.Reader.
type ThrottledReader struct {
	r     io.Reader
	ctx   context.Context
	bps   int64 // Bytes per second
	total int64 // Total bytes read
	start time.Time
}

// NewThrottledReader membuat instance baru ThrottledReader.
func NewThrottledReader(ctx context.Context, r io.Reader, bps int64) *ThrottledReader {
	return &ThrottledReader{
		r:     r,
		ctx:   ctx,
		bps:   bps,
		start: time.Now(),
	}
}

func (t *ThrottledReader) Read(p []byte) (n int, err error) {
	if t.bps <= 0 {
		return t.r.Read(p)
	}

	n, err = t.r.Read(p)
	if n <= 0 {
		return n, err
	}

	t.total += int64(n)
	
	// Kalkulasi waktu yang seharusnya dibutuhkan untuk jumlah bytes ini
	expectedDuration := time.Duration(t.total) * time.Second / time.Duration(t.bps)
	actualDuration := time.Since(t.start)

	if actualDuration < expectedDuration {
		// Kita terlalu cepat, harus tidur sejenak
		sleepTime := expectedDuration - actualDuration
		
		timer := time.NewTimer(sleepTime)
		select {
		case <-t.ctx.Done():
			timer.Stop()
			return n, t.ctx.Err()
		case <-timer.C:
		}
	}

	return n, err
}

// TotalBytes mengembalikan jumlah byte yang telah dibaca.
func (t *ThrottledReader) TotalBytes() int64 {
	return t.total
}
