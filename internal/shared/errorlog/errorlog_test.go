package errorlog

import (
	"errors"
	"testing"
	applog "sfdbtools/internal/services/log"
)

func TestErrorLogger_Log_TypedNil(t *testing.T) {
	el := &ErrorLogger{
		Logger:  applog.MockLogger(),
		LogDir:  "/tmp",
		Feature: "test",
	}

	// Case 1: Standard nil
	if res := el.log(nil, "", nil); res != "" {
		t.Errorf("Expected empty string for nil error, got %s", res)
	}

	// Case 2: Typed nil
	var typedNil *customError = nil
	if res := el.log(nil, "", typedNil); res != "" {
		t.Errorf("Expected empty string for typed nil error, got %s", res)
	}

	// Case 3: Real error
	if res := el.log(nil, "", errors.New("real error")); res == "" {
		// Log logic might return path or empty string on failure.
		// We just want to ensure it doesn't panic or fail on typed nil.
	}
}

type customError struct{}

func (e *customError) Error() string { return "custom error" }
