package security

import (
	"encoding/hex"
	"testing"
)

func TestGenerateSecureKey(t *testing.T) {
	key, err := GenerateSecureKey()
	if err != nil {
		t.Fatalf("GenerateSecureKey() error = %v", err)
	}

	decoded, err := hex.DecodeString(key)
	if err != nil {
		t.Fatalf("key is not valid hex: %v", err)
	}

	if len(decoded) != 32 {
		t.Errorf("key length = %d, want 32 bytes", len(decoded))
	}

	// Verify uniqueness
	key2, err := GenerateSecureKey()
	if err != nil {
		t.Fatalf("GenerateSecureKey() error = %v", err)
	}
	if key == key2 {
		t.Error("two consecutive keys should not be equal")
	}
}

func TestGenerateSignature(t *testing.T) {
	payload := []byte(`{"event":"test"}`)
	secret := "test-secret-key"

	sig, err := GenerateSignature(payload, secret)
	if err != nil {
		t.Fatalf("GenerateSignature() error = %v", err)
	}

	if sig == "" {
		t.Error("signature should not be empty")
	}

	// Same input should produce same signature
	sig2, err := GenerateSignature(payload, secret)
	if err != nil {
		t.Fatalf("GenerateSignature() error = %v", err)
	}
	if sig != sig2 {
		t.Errorf("same input produced different signatures: %s vs %s", sig, sig2)
	}

	// Different payload should produce different signature
	sig3, err := GenerateSignature([]byte("other"), secret)
	if err != nil {
		t.Fatalf("GenerateSignature() error = %v", err)
	}
	if sig == sig3 {
		t.Error("different payloads should produce different signatures")
	}
}

func TestVerifySignature(t *testing.T) {
	payload := []byte(`{"event":"test"}`)
	secret := "test-secret-key"

	sig, err := GenerateSignature(payload, secret)
	if err != nil {
		t.Fatalf("GenerateSignature() error = %v", err)
	}

	if !VerifySignature(payload, secret, sig) {
		t.Error("VerifySignature should return true for valid signature")
	}

	if VerifySignature(payload, "wrong-secret", sig) {
		t.Error("VerifySignature should return false for wrong secret")
	}

	if VerifySignature([]byte("tampered"), secret, sig) {
		t.Error("VerifySignature should return false for tampered payload")
	}

	if VerifySignature(payload, secret, "invalid-sig") {
		t.Error("VerifySignature should return false for invalid signature")
	}
}
