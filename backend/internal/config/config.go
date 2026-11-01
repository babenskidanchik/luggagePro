package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Port               string
	DatabaseURL        string
	OTPHMACSecret      string
	DevOTPPath         string
	PhoneEncryptionKey string
	PhoneLookupSecret  string
	AllowedOrigins     []string
	Environment        string
	SecureCookies      bool
}

func Load() (Config, error) {
	databaseURL := strings.TrimSpace(os.Getenv("DATABASE_URL"))
	if databaseURL == "" {
		return Config{}, errors.New("DATABASE_URL is required")
	}

	allowedOriginsRaw :=
		strings.TrimSpace(
			os.Getenv("ALLOWED_ORIGINS"),
		)

	if allowedOriginsRaw == "" {
		return Config{}, fmt.Errorf(
			"ALLOWED_ORIGINS is required",
		)
	}

	var allowedOrigins []string

	for _, origin := range strings.Split(
		allowedOriginsRaw,
		",",
	) {
		origin = strings.TrimSpace(origin)

		if origin != "" {
			allowedOrigins = append(
				allowedOrigins,
				origin,
			)
		}
	}

	environment := strings.TrimSpace(
		os.Getenv("APP_ENV"),
	)

	if environment == "" {
		environment = "development"
	}

	if environment != "development" &&
		environment != "test" &&
		environment != "production" {
		return Config{}, fmt.Errorf(
			"invalid APP_ENV: %q",
			environment,
		)
	}

	secureCookies := false

	if raw := strings.TrimSpace(
		os.Getenv("SECURE_COOKIES"),
	); raw != "" {
		parsed, err := strconv.ParseBool(raw)
		if err != nil {
			return Config{}, fmt.Errorf(
				"invalid SECURE_COOKIES: %w",
				err,
			)
		}

		secureCookies = parsed
	}

	otpHMACSecret := strings.TrimSpace(
		os.Getenv("OTP_HMAC_SECRET"),
	)

	if len(otpHMACSecret) < 32 {
		return Config{}, errors.New(
			"OTP_HMAC_SECRET must contain at least 32 characters",
		)
	}

	devOTPPath := strings.TrimSpace(
		os.Getenv("DEV_OTP_PATH"),
	)

	if devOTPPath == "" {
		devOTPPath = "/tmp/luggage-platform-dev-otp"
	}

	port := strings.TrimSpace(os.Getenv("PORT"))
	if port == "" {
		port = "8080"
	}

	phoneEncryptionKey := strings.TrimSpace(
		os.Getenv("PHONE_ENCRYPTION_KEY"),
	)

	if phoneEncryptionKey == "" {
		return Config{}, errors.New(
			"PHONE_ENCRYPTION_KEY is required",
		)
	}

	phoneLookupSecret := strings.TrimSpace(
		os.Getenv("PHONE_LOOKUP_SECRET"),
	)

	if phoneLookupSecret == "" {
		return Config{}, errors.New(
			"PHONE_LOOKUP_SECRET is required",
		)
	}

	cfg := Config{
		Port:               port,
		DatabaseURL:        databaseURL,
		OTPHMACSecret:      otpHMACSecret,
		DevOTPPath:         devOTPPath,
		PhoneEncryptionKey: phoneEncryptionKey,
		PhoneLookupSecret:  phoneLookupSecret,
		AllowedOrigins:     allowedOrigins,
		Environment:        environment,
		SecureCookies:      secureCookies,
	}

	if cfg.Environment == "production" {
		if !cfg.SecureCookies {
			return Config{}, fmt.Errorf(
				"SECURE_COOKIES must be true in production",
			)
		}

		for _, origin := range cfg.AllowedOrigins {
			if !strings.HasPrefix(origin, "https://") {
				return Config{}, fmt.Errorf(
					"production origin must use HTTPS: %q",
					origin,
				)
			}
		}
	}

	return cfg, nil
}
