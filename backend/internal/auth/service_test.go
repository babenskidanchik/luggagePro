package auth

import (
	"bytes"
	"context"
	"testing"
	"time"

	"github.com/babenskidanchik/luggage-platform/backend/internal/domain"
)

type fakeSessionRepository struct {
	session *domain.Session
	hash    []byte
}

func (f *fakeSessionRepository) Create(
	ctx context.Context,
	userID string,
	tokenHash []byte,
	expiresAt time.Time,
	userAgent *string,
	ipAddress *string,
) (*domain.Session, error) {
	f.hash = append([]byte(nil), tokenHash...)

	f.session = &domain.Session{
		ID:        "session-id",
		UserID:    userID,
		TokenHash: tokenHash,
		ExpiresAt: expiresAt,
		CreatedAt: time.Now().UTC(),
		UserAgent: userAgent,
		IPAddress: ipAddress,
	}

	return f.session, nil
}

func (f *fakeSessionRepository) GetActiveByTokenHash(
	ctx context.Context,
	tokenHash []byte,
) (*domain.Session, error) {
	if f.session == nil {
		return nil, domain.ErrSessionNotFound
	}

	if !bytes.Equal(tokenHash, f.hash) {
		return nil, domain.ErrSessionNotFound
	}

	return f.session, nil
}

func (f *fakeSessionRepository) RevokeByTokenHash(
	ctx context.Context,
	tokenHash []byte,
) error {
	if f.session == nil {
		return nil
	}

	if bytes.Equal(tokenHash, f.hash) {
		now := time.Now().UTC()
		f.session.RevokedAt = &now
	}

	return nil
}

type fakeUserRepository struct {
	user *domain.User
}

func (f *fakeUserRepository) GetByID(
	ctx context.Context,
	id string,
) (*domain.User, error) {
	if f.user == nil || f.user.ID != id {
		return nil, domain.ErrUserNotFound
	}

	return f.user, nil
}

func TestCreateSessionAndCurrentUser(t *testing.T) {
	sessions := &fakeSessionRepository{}

	users := &fakeUserRepository{
		user: &domain.User{
			ID:       "user-id",
			Role:     domain.RoleGuest,
			IsActive: true,
		},
	}

	service := NewService(
		sessions,
		users,
	)

	token, err := service.CreateSession(
		context.Background(),
		users.user.ID,
		nil,
		nil,
	)
	if err != nil {
		t.Fatalf(
			"CreateSession() error: %v",
			err,
		)
	}

	if token == "" {
		t.Fatal("expected session token")
	}

	user, err := service.CurrentUser(
		context.Background(),
		token,
	)
	if err != nil {
		t.Fatalf(
			"CurrentUser() error: %v",
			err,
		)
	}

	if user.ID != users.user.ID {
		t.Fatalf(
			"expected user %q, got %q",
			users.user.ID,
			user.ID,
		)
	}
}

func TestInvalidTokenRejected(t *testing.T) {
	sessions := &fakeSessionRepository{}

	users := &fakeUserRepository{
		user: &domain.User{
			ID:       "user-id",
			Role:     domain.RoleGuest,
			IsActive: true,
		},
	}

	service := NewService(
		sessions,
		users,
	)

	_, err := service.CreateSession(
		context.Background(),
		users.user.ID,
		nil,
		nil,
	)
	if err != nil {
		t.Fatalf(
			"CreateSession() error: %v",
			err,
		)
	}

	_, err = service.CurrentUser(
		context.Background(),
		"fake-session-token",
	)

	if err == nil {
		t.Fatal(
			"expected invalid token to be rejected",
		)
	}
}

func TestInactiveUserRejected(t *testing.T) {
	sessions := &fakeSessionRepository{}

	users := &fakeUserRepository{
		user: &domain.User{
			ID:       "user-id",
			Role:     domain.RoleGuest,
			IsActive: false,
		},
	}

	service := NewService(
		sessions,
		users,
	)

	token, err := service.CreateSession(
		context.Background(),
		users.user.ID,
		nil,
		nil,
	)
	if err != nil {
		t.Fatalf(
			"CreateSession() error: %v",
			err,
		)
	}

	_, err = service.CurrentUser(
		context.Background(),
		token,
	)

	if err != domain.ErrUserInactive {
		t.Fatalf(
			"expected ErrUserInactive, got %v",
			err,
		)
	}
}

func TestLogoutRevokesSession(t *testing.T) {
	sessions := &fakeSessionRepository{}

	users := &fakeUserRepository{
		user: &domain.User{
			ID:       "user-id",
			Role:     domain.RoleGuest,
			IsActive: true,
		},
	}

	service := NewService(
		sessions,
		users,
	)

	token, err := service.CreateSession(
		context.Background(),
		users.user.ID,
		nil,
		nil,
	)
	if err != nil {
		t.Fatalf(
			"CreateSession() error: %v",
			err,
		)
	}

	if err := service.Logout(
		context.Background(),
		token,
	); err != nil {
		t.Fatalf(
			"Logout() error: %v",
			err,
		)
	}

	if sessions.session.RevokedAt == nil {
		t.Fatal(
			"expected session to be revoked",
		)
	}
}
