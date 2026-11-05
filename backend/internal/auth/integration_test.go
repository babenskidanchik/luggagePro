package auth_test

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/babenskidanchik/luggage-platform/backend/internal/auth"
	"github.com/babenskidanchik/luggage-platform/backend/internal/database"
	"github.com/babenskidanchik/luggage-platform/backend/internal/domain"
	"github.com/babenskidanchik/luggage-platform/backend/internal/migrate"
	"github.com/babenskidanchik/luggage-platform/backend/internal/pii"
	"github.com/babenskidanchik/luggage-platform/backend/internal/repository"
	"github.com/jackc/pgx/v5/pgxpool"
)

func testPhoneProtector(
	t *testing.T,
) *pii.PhoneProtector {
	t.Helper()

	encryptionKey := base64.StdEncoding.EncodeToString(
		bytes.Repeat([]byte{1}, 32),
	)

	lookupSecret := base64.StdEncoding.EncodeToString(
		bytes.Repeat([]byte{2}, 32),
	)

	protector, err := pii.NewPhoneProtector(
		encryptionKey,
		lookupSecret,
	)
	if err != nil {
		t.Fatalf(
			"create test phone protector: %v",
			err,
		)
	}

	return protector
}

func setupTestDB(t *testing.T) *pgxpool.Pool {
	t.Helper()

	databaseURL := strings.TrimSpace(
		os.Getenv("TEST_DATABASE_URL"),
	)

	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL is not set")
	}

	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		t.Fatalf("parse test database URL: %v", err)
	}
	if config.ConnConfig.Database != "luggage_test" {
		t.Fatalf("refusing to run against database %q", config.ConnConfig.Database)
	}

	ctx, cancel := context.WithTimeout(
		context.Background(),
		10*time.Second,
	)
	defer cancel()

	db, err := database.NewPostgresPool(
		ctx,
		databaseURL,
	)
	if err != nil {
		t.Fatalf("connect test database: %v", err)
	}

	if err := migrate.Run(ctx, db); err != nil {
		db.Close()
		t.Fatalf("run test migrations: %v", err)
	}

	if _, err := db.Exec(
		ctx,
		"TRUNCATE TABLE users CASCADE",
	); err != nil {
		db.Close()
		t.Fatalf("clean test database: %v", err)
	}

	t.Cleanup(func() {
		db.Close()
	})

	return db
}

func TestSessionIntergation(t *testing.T) {
	db := setupTestDB(t)

	ctx := context.Background()

	users := repository.NewUserRepository(db, testPhoneProtector(t))
	sessions := repository.NewSessionRepository(db)

	service := auth.NewService(
		sessions,
		users,
	)

	phone := "+995555100001"

	user, err := users.CreateGuest(
		ctx,
		phone,
	)
	if err != nil {
		t.Fatalf("CreateGuest() error: %v", err)
	}

	if user.Role != domain.RoleGuest {
		t.Fatalf(
			"expected guest role, got %q",
			user.Role,
		)
	}

	token, err := service.CreateSession(
		ctx,
		user.ID,
		nil,
		nil,
	)
	if err != nil {
		t.Fatalf("CreateSession() error: %v", err)
	}

	if token == "" {
		t.Fatal("expected non-empty session token")
	}

	currentUser, err := service.CurrentUser(
		ctx,
		token,
	)
	if err != nil {
		t.Fatalf("CurrentUser() error: %v", err)
	}

	if currentUser.ID != user.ID {
		t.Fatalf(
			"expected user %q, got %q",
			user.ID,
			currentUser.ID,
		)
	}

	_, err = service.CurrentUser(
		ctx,
		"random-invalid-token",
	)

	if !errors.Is(
		err,
		domain.ErrSessionNotFound,
	) {
		t.Fatalf(
			"expected ErrSessionNotFound, got %v",
			err,
		)
	}

	if err := service.Logout(
		ctx,
		token,
	); err != nil {
		t.Fatalf("Logout() error: %v", err)
	}

	_, err = service.CurrentUser(
		ctx,
		token,
	)

	if !errors.Is(
		err,
		domain.ErrSessionNotFound,
	) {
		t.Fatalf(
			"expected revoked session to fail, got %v",
			err,
		)
	}
}

func TestExpiredSessionRejected(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	users := repository.NewUserRepository(db, testPhoneProtector(t))
	sessions := repository.NewSessionRepository(db)

	user, err := users.CreateGuest(
		ctx,
		"+995555100002",
	)
	if err != nil {
		t.Fatalf("CreateGuest() error: %v", err)
	}

	token, tokenHash, err := auth.GenerateSessionToken()
	if err != nil {
		t.Fatalf("GenerateSessionToken() error: %v", err)
	}

	expiresAt := time.Now().UTC().Add(-time.Hour)

	_, err = sessions.Create(
		ctx,
		user.ID,
		tokenHash,
		expiresAt,
		nil,
		nil,
	)
	if err != nil {
		t.Fatalf("Create() error: %v", err)
	}

	service := auth.NewService(
		sessions,
		users,
	)

	_, err = service.CurrentUser(
		ctx,
		token,
	)

	if !errors.Is(
		err,
		domain.ErrSessionNotFound,
	) {
		t.Fatalf(
			"expected expired session to be rejected, got %v",
			err,
		)
	}
}

func TestInactiveUserRejectedIntegration(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	users := repository.NewUserRepository(db, testPhoneProtector(t))
	sessions := repository.NewSessionRepository(db)

	user, err := users.CreateGuest(
		ctx,
		"+995555100003",
	)
	if err != nil {
		t.Fatalf("CreateGuest() error: %v", err)
	}

	service := auth.NewService(
		sessions,
		users,
	)

	token, err := service.CreateSession(
		ctx,
		user.ID,
		nil,
		nil,
	)
	if err != nil {
		t.Fatalf("CreateSession() error: %v", err)
	}

	_, err = db.Exec(
		ctx,
		`
			UPDATE users
			SET is_active = FALSE
			WHERE id = $1
		`,
		user.ID,
	)
	if err != nil {
		t.Fatalf("disable user: %v", err)
	}

	_, err = service.CurrentUser(
		ctx,
		token,
	)

	if !errors.Is(
		err,
		domain.ErrUserInactive,
	) {
		t.Fatalf(
			"expected ErrUserInactive, got %v",
			err,
		)
	}
}

// test SQL injection
func TestCreateGuestTreatsSQLInjectionAsData(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	users := repository.NewUserRepository(db, testPhoneProtector(t))

	maliciousPhone := `'+1' OR 1=1;--`

	user, err := users.CreateGuest(
		ctx,
		maliciousPhone,
	)
	if err != nil {
		t.Fatalf(
			"CreateGuest() error: %v",
			err,
		)
	}

	if user.Phone == nil {
		t.Fatal("expected phone")
	}

	if *user.Phone != maliciousPhone {
		t.Fatalf(
			"expected phone to be strong as data",
		)
	}

	var usersTableExists bool

	err = db.QueryRow(
		ctx,
		`
			SELECT EXISTS (
				SELECT 1
				FROM information_schema.tables
				WHERE table_schema = 'public'
					AND table_name = 'users'
			)
		`,
	).Scan(&usersTableExists)

	if err != nil {
		t.Fatalf(
			"check users table: %v",
			err,
		)
	}

	if !usersTableExists {
		t.Fatal(
			"users table was unexpectedly removed",
		)
	}
}

func TestOTPRepositoryIntegration(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	repo :=
		repository.NewPhoneVerificationRepository(db, testPhoneProtector(t))

	secret := []byte(
		"01234567890123456789012345678901",
	)

	phone := "+995555200001"
	code := "583291"

	now := time.Now().UTC()

	codeHash := auth.HashOTP(
		secret,
		phone,
		code,
	)

	verification, err := repo.Create(
		ctx,
		phone,
		codeHash,
		now,
		now.Add(5*time.Minute),
	)
	if err != nil {
		t.Fatalf(
			"Create() error: %v",
			err,
		)
	}

	if verification.Phone != phone {
		t.Fatalf(
			"expected phone %q, got %q",
			phone,
			verification.Phone,
		)
	}

	candidateHash := auth.HashOTP(
		secret,
		phone,
		code,
	)

	verified, err := repo.VerifyAndConsume(
		ctx,
		phone,
		candidateHash,
		now.Add(time.Second),
	)
	if err != nil {
		t.Fatalf(
			"VerifyAndConsume() error: %v",
			err,
		)
	}

	if verified.ConsumedAt == nil {
		t.Fatal(
			"expected OTP to be consumed",
		)
	}

	_, err = repo.VerifyAndConsume(
		ctx,
		phone,
		candidateHash,
		now.Add(2*time.Second),
	)

	if !errors.Is(
		err,
		domain.ErrOTPNotFound,
	) {
		t.Fatalf(
			"expected consumed OTP to be rejected, got %v",
			err,
		)
	}
}
