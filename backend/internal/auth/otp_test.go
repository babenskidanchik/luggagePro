package auth

import (
	"regexp"
	"testing"
)

func TestGenerateOTP(t *testing.T) {
	pattern := regexp.MustCompile(`^[0-9]{6}$`)

	for i := 0; i < 100; i++ {
		code, err := GenerateOTP()
		if err != nil {
			t.Fatalf(
				"GenerateOTP() error: %v",
				err,
			)
		}

		if !pattern.MatchString(code) {
			t.Fatalf(
				"invalid OTP format: %q",
				code,
			)
		}
	}
}

func TestOTPHashMatches(t *testing.T) {
	secret := []byte(
		"01234567890123456789012345678901",
	)

	phone := "+995555123456"
	code := "583291"

	hash := HashOTP(
		secret,
		phone,
		code,
	)

	if !OTPHashMatches(
		hash,
		HashOTP(secret, phone, code),
	) {
		t.Fatal(
			"expected identical OTP to match",
		)
	}

	if OTPHashMatches(
		hash,
		HashOTP(secret, phone, "111111"),
	) {
		t.Fatal(
			"expected wrong OTP not to match",
		)
	}
}

func TestOTPHashDependsOnPhone(t *testing.T) {
	secret := []byte(
		"01234567890123456789012345678901",
	)

	code := "583291"

	first := HashOTP(
		secret,
		"+995555111111",
		code,
	)

	second := HashOTP(
		secret,
		"+995555222222",
		code,
	)

	if OTPHashMatches(first, second) {
		t.Fatal(
			"expected different phones to produce different hashes",
		)
	}
}
