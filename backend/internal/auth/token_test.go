package auth

import (
	"bytes"
	"testing"
)

func TestGenerateSessionToken(t *testing.T) {
	token, hash, err := GenerateSessionToken()
	if err != nil {
		t.Fatalf("GenerateSessionToken() error: %v", err)
	}

	if token == "" {
		t.Fatalf("GenerateSessionToken() error: %v", err)
	}

	if len(hash) != 32 {
		t.Fatalf(
			"expected SHA-256 hash length 32, got %d",
			len(hash),
		)
	}

	expectedHash := HashSessionToken(token)

	if !bytes.Equal(hash, expectedHash) {
		t.Fatal("generated hash does not match token")
	}
}

func TestGenerateSessionTokenAreUnique(t *testing.T) {
	token1, _, err := GenerateSessionToken()
	if err != nil {
		t.Fatalf("first token: %v", err)
	}

	token2, _, err := GenerateSessionToken()
	if err != nil {
		t.Fatalf("second token: %v", err)
	}

	if token1 == token2 {
		t.Fatal("expected different session tokens")
	}
}
