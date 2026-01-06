package encrypt

import (
	"os"
	"strings"
	"testing"
)

func TestDecodeEnvValue_NoPrefix(t *testing.T) {
	got, wasEncrypted, err := DecodeEnvValue("hello")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if wasEncrypted {
		t.Fatalf("expected wasEncrypted=false")
	}
	if got != "hello" {
		t.Fatalf("expected 'hello', got %q", got)
	}
}

func TestEncodeDecodeEnvValue_RoundTrip(t *testing.T) {
	encoded, err := EncodeEnvValue("  my-secret  ")
	if err != nil {
		t.Fatalf("EncodeEnvValue error: %v", err)
	}
	if !strings.HasPrefix(encoded, EnvEncryptedPrefix) {
		t.Fatalf("expected prefix %q, got %q", EnvEncryptedPrefix, encoded)
	}

	decoded, wasEncrypted, err := DecodeEnvValue(encoded)
	if err != nil {
		t.Fatalf("DecodeEnvValue error: %v", err)
	}
	if !wasEncrypted {
		t.Fatalf("expected wasEncrypted=true")
	}
	if decoded != "my-secret" {
		t.Fatalf("expected 'my-secret', got %q", decoded)
	}
}

func TestDecodeEnvValue_InvalidPayload(t *testing.T) {
	_, wasEncrypted, err := DecodeEnvValue(EnvEncryptedPrefix + "@@@")
	if err == nil {
		t.Fatalf("expected error")
	}
	if !wasEncrypted {
		t.Fatalf("expected wasEncrypted=true")
	}
}

func TestResolveEnvSecret_Plaintext(t *testing.T) {
	t.Setenv("SFDB_TEST_SECRET", "plain")
	got, err := ResolveEnvSecret("SFDB_TEST_SECRET")
	if err != nil {
		t.Fatalf("ResolveEnvSecret error: %v", err)
	}
	if got != "plain" {
		t.Fatalf("expected 'plain', got %q", got)
	}
}

func TestReadMariaDBKeyMaterial_SelectHighestID(t *testing.T) {
	f, err := os.CreateTemp("", "sfdbtools-key-maria-*.txt")
	if err != nil {
		t.Fatalf("CreateTemp: %v", err)
	}
	defer os.Remove(f.Name())

	content := "1;AAAA\n100;BBBB\n2;0x0102\n5;0A0B\n"
	if _, err := f.WriteString(content); err != nil {
		t.Fatalf("WriteString: %v", err)
	}
	if err := f.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	got := readMariaDBKeyMaterial(f.Name())
	if len(got) != 2 || got[0] != 0x0a || got[1] != 0x0b {
		t.Fatalf("expected key 0A0B, got %x", got)
	}
}
