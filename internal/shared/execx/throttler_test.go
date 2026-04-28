package execx

import (
	"context"
	"io"
	"strings"
	"testing"
	"time"
)

func TestParseSpeed(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    int64
		wantErr bool
	}{
		{"empty", "", 0, false},
		{"zero", "0", 0, false},
		{"bytes only", "1024", 1024, false},
		{"KB", "10KB", 10 * 1024, false},
		{"MB", "5MB", 5 * 1024 * 1024, false},
		{"GB", "2 GB", 2 * 1024 * 1024 * 1024, false},
		{"MB/s format", "10MB/s", 10 * 1024 * 1024, false},
		{"lowercase", "500kb", 500 * 1024, false},
		{"unit with spaces", "100 MB", 100 * 1024 * 1024, false},
		{"invalid unit", "10XY", 0, true},
		{"invalid format", "abc", 0, true},
		{"decimal (not supported by regex currently, but good to test)", "1.5MB", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseSpeed(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseSpeed(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ParseSpeed(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestThrottledReader_Timing(t *testing.T) {
	// Test data: 100KB
	size := 100 * 1024
	data := strings.Repeat("a", size)

	// Speed: 50KB/s
	// Expected duration: 100 / 50 = 2 seconds
	speed := int64(50 * 1024)

	ctx := context.Background()
	reader := strings.NewReader(data)
	throttled := NewThrottledReader(ctx, reader, speed)

	start := time.Now()

	buf := make([]byte, 8192)
	totalRead := 0
	for {
		n, err := throttled.Read(buf)
		totalRead += n
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
	}

	duration := time.Since(start)

	if totalRead != size {
		t.Errorf("Total read %d, want %d", totalRead, size)
	}

	// Kita toleransi margin error waktu (misal min 1.8 detik)
	if duration < 1800*time.Millisecond {
		t.Errorf("Throttling too fast: finished in %v, expected ~2s", duration)
	}

	if duration > 3*time.Second {
		t.Errorf("Throttling too slow: finished in %v, expected ~2s", duration)
	}
}

func TestThrottledReader_NoThrottle(t *testing.T) {
	data := "some data"
	reader := strings.NewReader(data)
	throttled := NewThrottledReader(context.Background(), reader, 0) // 0 means no limit

	buf := make([]byte, 1024)
	n, err := throttled.Read(buf)
	if n != len(data) || (err != nil && err != io.EOF) {
		t.Errorf("Read failed with no throttle: n=%d, err=%v", n, err)
	}
}
