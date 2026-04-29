package notify

import (
	"strings"
	"testing"
)

func TestFormatTelegramMessage(t *testing.T) {
	msg := Message{
		Title:   "Test Title & More",
		Body:    "Test Body with <tag>",
		Level:   LevelCritical,
		Feature: "test-feature",
	}

	formatted := FormatTelegramMessage(msg)

	// Check emoji
	if !strings.Contains(formatted, "🔴") {
		t.Errorf("expected critical emoji, got %s", formatted)
	}

	// Check escaped title
	if !strings.Contains(formatted, "Test Title &amp; More") {
		t.Errorf("expected escaped title, got %s", formatted)
	}

	// Check feature
	if !strings.Contains(formatted, "test-feature") {
		t.Errorf("expected feature name, got %s", formatted)
	}

	// Check body
	if !strings.Contains(formatted, "Test Body with <tag>") {
		t.Errorf("expected body, got %s", formatted)
	}
}

func TestFormatEmailMessage(t *testing.T) {
	msg := Message{
		Title:   "Test Title",
		Body:    "Test Body",
		Level:   LevelInfo,
		Feature: "test-feature",
	}

	formatted := FormatEmailMessage(msg)

	// Check icon
	if !strings.Contains(formatted, "[INFO]") {
		t.Errorf("expected INFO icon, got %s", formatted)
	}

	// Check title
	if !strings.Contains(formatted, "Test Title") {
		t.Errorf("expected title, got %s", formatted)
	}

	// Check feature
	if !strings.Contains(formatted, "test-feature") {
		t.Errorf("expected feature name, got %s", formatted)
	}
}

func TestEscapeHTML(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Hello & World", "Hello &amp; World"},
		{"<script>", "&lt;script&gt;"},
		{"a > b", "a &gt; b"},
	}

	for _, tc := range tests {
		got := escapeHTML(tc.input)
		if got != tc.expected {
			t.Errorf("escapeHTML(%s) = %s; want %s", tc.input, got, tc.expected)
		}
	}
}
