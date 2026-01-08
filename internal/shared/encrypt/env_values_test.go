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
	prefix := EnvEncryptedPrefixForDisplay()
	if !strings.HasPrefix(encoded, prefix) {
		t.Fatalf("expected prefix %q, got %q", prefix, encoded)
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
	_, wasEncrypted, err := DecodeEnvValue(EnvEncryptedPrefixForDisplay() + "@@@")
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

	// ID 2: 2 bytes (terlalu pendek, akan di-skip oleh validasi min 16 bytes)
	// ID 5: 16 bytes (valid, minimum untuk AES-128) = 32 hex chars
	// ID 10: 32 bytes (valid, lebih panjang) = 64 hex chars
	content := "1;AAAA\n100;BBBB\n2;0x0102\n5;0102030405060708090A0B0C0D0E0F10\n10;000102030405060708090A0B0C0D0E0F101112131415161718191A1B1C1D1E1F\n"
	if _, err := f.WriteString(content); err != nil {
		t.Fatalf("WriteString: %v", err)
	}
	if err := f.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	got := readMariaDBKeyMaterial(f.Name())
	// Expect ID 10 (highest valid ID dengan key >=16 bytes)
	if len(got) != 32 {
		t.Fatalf("expected 32 bytes, got %d", len(got))
	}
	if got[0] != 0x00 || got[1] != 0x01 {
		t.Fatalf("expected key starting with 0001, got %02x%02x", got[0], got[1])
	}
}

func TestMariaDBKeyCaching(t *testing.T) {
	// Test bahwa key material di-cache (fix #5)
	f, err := os.CreateTemp("", "sfdbtools-key-cache-*.txt")
	if err != nil {
		t.Fatalf("CreateTemp: %v", err)
	}
	defer os.Remove(f.Name())

	content := "2;0102030405060708090A0B0C0D0E0F10\n"
	if _, err := f.WriteString(content); err != nil {
		t.Fatalf("WriteString: %v", err)
	}
	if err := f.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	// First call: read dari file
	got1 := getCachedMariaDBKeyMaterial(f.Name())
	if len(got1) != 16 {
		t.Fatalf("expected 16 bytes, got %d", len(got1))
	}

	// Ubah file content
	if err := os.WriteFile(f.Name(), []byte("2;FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF\n"), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	// Second call: harus return cached value (tidak re-read file)
	got2 := getCachedMariaDBKeyMaterial(f.Name())
	if len(got2) != 16 {
		t.Fatalf("expected 16 bytes from cache, got %d", len(got2))
	}
	// Verify nilai sama dengan first call (cached)
	for i := 0; i < 16; i++ {
		if got1[i] != got2[i] {
			t.Fatalf("cache mismatch at byte %d: first=%02x, second=%02x", i, got1[i], got2[i])
		}
	}
}

func TestFailedDecodeCounter(t *testing.T) {
	// Test bahwa failed decode di-count (fix #9)
	initialCount := GetFailedDecodeCount()

	// Trigger failed decode dengan valid base64 tapi wrong ciphertext
	// Payload: version=1 (byte 0) + 12 nonce + >=16 ciphertext = minimum 29 bytes
	prefix := EnvEncryptedPrefixForDisplay()
	fakePayload := "AQECAwQFBgcICQoLDA0ODxAREhMUFRYXGBkaGxwd" // starts with 0x01 (version)
	_, _, err := DecodeEnvValue(prefix + fakePayload)
	if err == nil {
		t.Fatalf("expected error for wrong ciphertext")
	}

	newCount := GetFailedDecodeCount()
	if newCount <= initialCount {
		t.Fatalf("expected counter to increment, got initial=%d new=%d", initialCount, newCount)
	}
}
