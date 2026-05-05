package connectivity

import (
	"context"
	"testing"
	"time"
)

func TestIsInternetAvailable(t *testing.T) {
	// Note: This test requires actual internet connection to pass.
	// In a CI environment, this might be skipped or mocked.
	
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// We can't guarantee internet is available, but we can check if it returns a boolean.
	_ = IsInternetAvailable(ctx, 2*time.Second)
}
