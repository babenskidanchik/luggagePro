package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
)

const sessionTokenBytes = 32

func GenerateSessionToken() (string, []byte, error) {
	randomBytes := make([]byte, sessionTokenBytes)

	if _, err := rand.Read(randomBytes); err != nil {
		return "", nil, fmt.Errorf(
			"generate session token: %w",
			err,
		)
	}

	token := base64.RawURLEncoding.EncodeToString(randomBytes)

	hash := sha256.Sum256([]byte(token))

	return token, hash[:], nil
}

func HashSessionToken(token string) []byte {
	hash := sha256.Sum256([]byte(token))

	return hash[:]
}
