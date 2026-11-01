package auth

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/babenskidanchik/luggage-platform/backend/internal/domain"
)

const sessionLifetime = 7 * 24 * time.Hour

type SessionRepository interface {
	Create(
		ctx context.Context,
		userID string,
		tokenHash []byte,
		expiresAt time.Time,
		userAgent *string,
		ipAddress *string,
	) (*domain.Session, error)

	GetActiveByTokenHash(
		ctx context.Context,
		tokenHash []byte,
	) (*domain.Session, error)

	RevokeByTokenHash(
		ctx context.Context,
		tokenHash []byte,
	) error
}

type UserRepository interface {
	GetByID(
		ctx context.Context,
		id string,
	) (*domain.User, error)
}

type Service struct {
	sessions SessionRepository
	users    UserRepository
}

func NewService(
	sessions SessionRepository,
	users UserRepository,
) *Service {
	return &Service{
		sessions: sessions,
		users:    users,
	}
}

func (s *Service) CreateSession(
	ctx context.Context,
	userID string,
	userAgent *string,
	ipAddress *string,
) (string, error) {
	token, tokenHash, err := GenerateSessionToken()
	if err != nil {
		return "", err
	}

	expiresAt := time.Now().UTC().Add(sessionLifetime)

	if _, err := s.sessions.Create(
		ctx,
		userID,
		tokenHash,
		expiresAt,
		userAgent,
		ipAddress,
	); err != nil {
		return "", fmt.Errorf(
			"create auth session: %w",
			err,
		)
	}

	return token, nil
}

func (s *Service) CurrentUser(
	ctx context.Context,
	token string,
) (*domain.User, error) {
	token = strings.TrimSpace(token)

	if token == "" {
		return nil, domain.ErrSessionNotFound
	}

	tokenHash := HashSessionToken(token)

	session, err := s.sessions.GetActiveByTokenHash(
		ctx,
		tokenHash,
	)
	if err != nil {
		return nil, err
	}

	user, err := s.users.GetByID(
		ctx,
		session.UserID,
	)
	if err != nil {
		return nil, err
	}

	if !user.IsActive {
		return nil, domain.ErrUserInactive
	}

	return user, nil
}

func (s *Service) Logout(
	ctx context.Context,
	token string,
) error {
	token = strings.TrimSpace(token)

	if token == "" {
		return nil
	}

	tokenHash := HashSessionToken(token)

	err := s.sessions.RevokeByTokenHash(
		ctx,
		tokenHash,
	)

	if err != nil {
		return fmt.Errorf(
			"logout: %w",
			err,
		)
	}

	return nil
}

func IsAuthenticationError(err error) bool {
	return errors.Is(err, domain.ErrSessionNotFound) ||
		errors.Is(err, domain.ErrUserNotFound) ||
		errors.Is(err, domain.ErrUserInactive)
}
