package pii

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
)

type PhoneProtector struct {
	aead         cipher.AEAD
	lookupSecret []byte
}

func NewPhoneProtector(
	encryptionKeyBase64 string,
	lookupSecretBase64 string,
) (*PhoneProtector, error) {
	encryptionKey, err := base64.StdEncoding.DecodeString(
		encryptionKeyBase64,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"decode phone encryption key: %w",
			err,
		)
	}

	if len(encryptionKey) != 32 {
		return nil, errors.New(
			"phone encryption key must be exactly 32 bytes",
		)
	}

	lookupSecret, err := base64.StdEncoding.DecodeString(
		lookupSecretBase64,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"decode phone lookup secret: %w",
			err,
		)
	}

	if len(lookupSecret) < 32 {
		return nil, errors.New(
			"phone lookup secret must contain at least 32 bytes",
		)
	}

	block, err := aes.NewCipher(encryptionKey)
	if err != nil {
		return nil, fmt.Errorf(
			"create AES cipher: %w",
			err,
		)
	}

	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf(
			"create AES-GCM: %w",
			err,
		)
	}

	return &PhoneProtector{
		aead:         aead,
		lookupSecret: append([]byte(nil), lookupSecret...),
	}, nil
}

func (p *PhoneProtector) Encrypt(
	phone string,
) ([]byte, error) {
	nonce := make(
		[]byte,
		p.aead.NonceSize(),
	)

	if _, err := io.ReadFull(
		rand.Reader,
		nonce,
	); err != nil {
		return nil, fmt.Errorf(
			"generate phone nonce: %w",
			err,
		)
	}

	ciphertext := p.aead.Seal(
		nil,
		nonce,
		[]byte(phone),
		nil,
	)

	result := make(
		[]byte,
		0,
		len(nonce)+len(ciphertext),
	)

	result = append(result, nonce...)
	result = append(result, ciphertext...)

	return result, nil
}

func (p *PhoneProtector) Decrypt(
	encrypted []byte,
) (string, error) {
	nonceSize := p.aead.NonceSize()

	if len(encrypted) <= nonceSize {
		return "", errors.New(
			"invalid encrypted phone",
		)
	}

	nonce := encrypted[:nonceSize]
	ciphertext := encrypted[nonceSize:]

	plaintext, err := p.aead.Open(
		nil,
		nonce,
		ciphertext,
		nil,
	)
	if err != nil {
		return "", errors.New(
			"decrypt phonee failed",
		)
	}

	return string(plaintext), nil
}

func (p *PhoneProtector) LookupHash(
	phone string,
) []byte {
	mac := hmac.New(
		sha256.New,
		p.lookupSecret,
	)

	_, _ = mac.Write(
		[]byte(phone),
	)

	return mac.Sum(nil)
}
