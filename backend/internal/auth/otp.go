package auth

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"math/big"
)

const otpDigits = 6

func GenerateOTP() (string, error) {
	const maxOTP = 1_000_000

	number, err := rand.Int(
		rand.Reader,
		big.NewInt(maxOTP),
	)
	if err != nil {
		return "", fmt.Errorf(
			"generate OTP: %w",
			err,
		)
	}

	return fmt.Sprintf(
		"%0*d",
		otpDigits,
		number.Int64(),
	), nil
}

func HashOTP(
	secret []byte,
	phone string,
	code string,
) []byte {
	mac := hmac.New(
		sha256.New,
		secret,
	)

	_, _ = mac.Write([]byte(phone))
	_, _ = mac.Write([]byte{0})
	_, _ = mac.Write([]byte(code))

	return mac.Sum(nil)
}

func OTPHashMatches(
	expected []byte,
	actual []byte,
) bool {
	return hmac.Equal(
		expected,
		actual,
	)
}
