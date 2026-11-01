package main

import (
	"context"
	"log"
	"time"

	"github.com/babenskidanchik/luggage-platform/backend/internal/auth"
	"github.com/babenskidanchik/luggage-platform/backend/internal/config"
	"github.com/babenskidanchik/luggage-platform/backend/internal/database"
	"github.com/babenskidanchik/luggage-platform/backend/internal/httpapi"
	"github.com/babenskidanchik/luggage-platform/backend/internal/httpapi/handlers"
	"github.com/babenskidanchik/luggage-platform/backend/internal/migrate"
	"github.com/babenskidanchik/luggage-platform/backend/internal/otpdelivery"
	"github.com/babenskidanchik/luggage-platform/backend/internal/pii"
	"github.com/babenskidanchik/luggage-platform/backend/internal/repository"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	ctx, cancel := context.WithTimeout(
		context.Background(),
		10*time.Second,
	)
	defer cancel()

	db, err := database.NewPostgresPool(
		ctx,
		cfg.DatabaseURL,
	)
	if err != nil {
		log.Fatalf("connect database: %v", err)
	}
	defer db.Close()

	phoneProtector, err := pii.NewPhoneProtector(
		cfg.PhoneEncryptionKey,
		cfg.PhoneLookupSecret,
	)
	if err != nil {
		log.Fatalf(
			"create phone protector: %v",
			err,
		)
	}

	migrationCtx, migrationCancel := context.WithTimeout(
		context.Background(),
		30*time.Second,
	)
	defer migrationCancel()

	if err := migrate.Run(migrationCtx, db); err != nil {
		log.Fatalf("run migrations: %v", err)
	}

	userRepository :=
		repository.NewUserRepository(db, phoneProtector)

	sessionRepository :=
		repository.NewSessionRepository(db)

	verificationRepostiory :=
		repository.NewPhoneVerificationRepository(db, phoneProtector)

	authService := auth.NewService(
		sessionRepository,
		userRepository,
	)

	otpSender := otpdelivery.NewDevSender(
		cfg.DevOTPPath,
	)

	otpService := auth.NewOTPService(
		verificationRepostiory,
		userRepository,
		authService,
		otpSender,
		[]byte(cfg.OTPHMACSecret),
	)

	otpHandler := handlers.NewOTPHandler(
		otpService,
		cfg.SecureCookies,
	)

	router := httpapi.NewRouter(
		db,
		otpHandler,
		phoneProtector,
		cfg.AllowedOrigins,
		cfg.SecureCookies,
		cfg.Environment,
	)

	log.Printf("API listening on port %s", cfg.Port)

	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatalf("run server: %v:", err)
	}
}
