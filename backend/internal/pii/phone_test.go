package pii

import (
	"bytes"
	"encoding/base64"
	"testing"
)

func testKey(value byte) string {
	raw := bytes.Repeat(
		[]byte{value},
		32,
	)

	return base64.StdEncoding.EncodeToString(raw)
}

func TestPhoneEncryptDecrypt(t *testing.T) {
	protector, err := NewPhoneProtector(
		testKey(1),
		testKey(2),
	)
	if err != nil {
		t.Fatalf(
			"NewPhoneProtector() error: %v",
			err,
		)
	}

	phone := "+995555123456"

	encrypted, err := protector.Encrypt(phone)
	if err != nil {
		t.Fatalf(
			"Encrypt() error: %v",
			err,
		)
	}

	if bytes.Contains(
		encrypted,
		[]byte(phone),
	) {
		t.Fatal(
			"ciphertext contains plaintext phone",
		)
	}

	decrypted, err := protector.Decrypt(
		encrypted,
	)
	if err != nil {
		t.Fatalf(
			"Decrypt() error: %v",
			err,
		)
	}

	if decrypted != phone {
		t.Fatalf(
			"expected %q, got %q",
			phone,
			decrypted,
		)
	}
}

func TestPhoneEncryptionUsesRandomNonce(
	t *testing.T,
) {
	protector, err := NewPhoneProtector(
		testKey(1),
		testKey(2),
	)
	if err != nil {
		t.Fatal(err)
	}

	phone := "+995555123456"

	first, err := protector.Encrypt(phone)
	if err != nil {
		t.Fatal(err)
	}

	second, err := protector.Encrypt(phone)
	if err != nil {
		t.Fatal(err)
	}

	if bytes.Equal(first, second) {
		t.Fatal(
			"expected different ciphertexts",
		)
	}
}

func TestPhoneLookupHashStable(
	t *testing.T,
) {
	protector, err := NewPhoneProtector(
		testKey(1),
		testKey(2),
	)
	if err != nil {
		t.Fatal(err)
	}

	first := protector.LookupHash(
		"+995555123456",
	)

	second := protector.LookupHash(
		"+995555123456",
	)

	if !bytes.Equal(first, second) {
		t.Fatal(
			"expected stable lookup hash",
		)
	}
}

func TestPhoneLookupHashDiffersByPhone(
	t *testing.T,
) {
	protector, err := NewPhoneProtector(
		testKey(1),
		testKey(2),
	)
	if err != nil {
		t.Fatal(err)
	}

	first := protector.LookupHash(
		"+995555111111",
	)

	second := protector.LookupHash(
		"+995555222222",
	)

	if bytes.Equal(first, second) {
		t.Fatal(
			"expected different lookup hashes",
		)
	}
}

func TestPhoneTamperingRejected(
	t *testing.T,
) {
	protector, err := NewPhoneProtector(
		testKey(1),
		testKey(2),
	)
	if err != nil {
		t.Fatal(err)
	}

	encrypted, err := protector.Encrypt(
		"+995555123456",
	)
	if err != nil {
		t.Fatal(err)
	}

	encrypted[len(encrypted)-1] ^= 0x01

	_, err = protector.Decrypt(
		encrypted,
	)

	if err == nil {
		t.Fatal(
			"expected modified ciphertext to be rejected",
		)
	}
}
