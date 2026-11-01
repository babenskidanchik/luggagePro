package auth

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/babenskidanchik/luggage-platform/backend/internal/domain"
)

const otpLifetime = 5 * time.Minute

var e164PhonePattern = regexp.MustCompile(
	`^\+[1-9][0-9]{7,14}$`,
)

type OTPVerificationRepository interface {
	Create(
		ctx context.Context,
		phone string,
		codeHash []byte,
		now time.Time,
		expiresAt time.Time,
	) (*domain.PhoneVerification, error)

	VerifyAndConsume(
		ctx context.Context,
		phone string,
		candidateHash []byte,
		now time.Time,
	) (*domain.PhoneVerification, error)
}

type GuestRepository interface {
	GetByPhone(
		ctx context.Context,
		phone string,
	) (*domain.User, error)

	CreateGuest(
		ctx context.Context,
		phone string,
	) (*domain.User, error)
}

type SessionIssuer interface {
	CreateSession(
		ctx context.Context,
		userID string,
		userAgent *string,
		ipAddress *string,
	) (string, error)
}

type OTPSender interface {
	SendOTP(
		ctx context.Context,
		phone string,
		code string,
	) error
}

type OTPService struct {
	verifications OTPVerificationRepository
	users         GuestRepository
	sessions      SessionIssuer
	sender        OTPSender
	secret        []byte
}

func NewOTPService(
	verifications OTPVerificationRepository,
	users GuestRepository,
	sessions SessionIssuer,
	sender OTPSender,
	secret []byte,
) *OTPService {
	return &OTPService{
		verifications: verifications,
		users:         users,
		sessions:      sessions,
		sender:        sender,
		secret:        append([]byte(nil), secret...),
	}
}

func (s *OTPService) RequestOTP(
	ctx context.Context,
	phone string,
) error {
	phone, err := normalizePhone(phone)
	if err != nil {
		return err
	}

	code, err := GenerateOTP()
	if err != nil {
		return err
	}

	codeHash := HashOTP(
		s.secret,
		phone,
		code,
	)

	now := time.Now().UTC()

	_, err = s.verifications.Create(
		ctx,
		phone,
		codeHash,
		now,
		now.Add(otpLifetime),
	)
	if err != nil {
		return err
	}

	if err := s.sender.SendOTP(
		ctx,
		phone,
		code,
	); err != nil {
		return fmt.Errorf(
			"send OTP: %w",
			err,
		)
	}

	return nil
}

func (s *OTPService) VerifyOTP(
	ctx context.Context,
	phone string,
	code string,
	userAgent *string,
	ipAddress *string,
) (string, error) {
	phone, err := normalizePhone(phone)
	if err != nil {
		return "", err
	}

	code = strings.TrimSpace(code)

	if !isValidOTPFormat(code) {
		return "", domain.ErrOTPInvalid
	}

	candidateHash := HashOTP(
		s.secret,
		phone,
		code,
	)

	_, err = s.verifications.VerifyAndConsume(
		ctx,
		phone,
		candidateHash,
		time.Now().UTC(),
	)
	if err != nil {
		return "", err
	}

	user, err := s.users.GetByPhone(
		ctx,
		phone,
	)

	if err != nil {
		if !errors.Is(err, domain.ErrUserNotFound) {
			return "", err
		}

		user, err = s.users.CreateGuest(
			ctx,
			phone,
		)

		if err != nil {
			// Защита от race:
			// другой запрос мог создать этого guest между
			// GetByPhone() и CreateGuest().
			if !errors.Is(
				err,
				domain.ErrUserAlreadyExists,
			) {
				return "", err
			}

			user, err = s.users.GetByPhone(
				ctx,
				phone,
			)
			if err != nil {
				return "", err
			}
		}
	}

	if user.Role != domain.RoleGuest {
		return "", domain.ErrPhoneInUse
	}

	if !user.IsActive {
		return "", domain.ErrUserInactive
	}

	token, err := s.sessions.CreateSession(
		ctx,
		user.ID,
		userAgent,
		ipAddress,
	)
	if err != nil {
		return "", err
	}

	return token, nil
}

func normalizePhone(
	phone string,
) (string, error) {
	phone = strings.TrimSpace(phone)

	if !e164PhonePattern.MatchString(phone) {
		return "", domain.ErrInvalidPhone
	}

	return phone, nil
}

func isValidOTPFormat(code string) bool {
	if len(code) != 6 {
		return false
	}

	for _, char := range code {
		if char < '0' || char > '9' {
			return false
		}
	}

	return true
}
