package auth

import (
	"bytes"
	"context"
	"testing"
	"time"

	"github.com/babenskidanchik/luggage-platform/backend/internal/domain"
)

type fakeOTPRepository struct {
	phone    string
	codeHash []byte
}

func (f *fakeOTPRepository) Create(
	ctx context.Context,
	phone string,
	codeHash []byte,
	now time.Time,
	expiresAt time.Time,
) (*domain.PhoneVerification, error) {
	f.phone = phone
	f.codeHash = append(
		[]byte(nil),
		codeHash...,
	)

	return &domain.PhoneVerification{
		ID:        "verification-id",
		Phone:     phone,
		CodeHash:  codeHash,
		ExpiresAt: expiresAt,
		CreatedAt: now,
	}, nil
}

func (f *fakeOTPRepository) VerifyAndConsume(
	ctx context.Context,
	phone string,
	candidateHash []byte,
	now time.Time,
) (*domain.PhoneVerification, error) {
	if phone != f.phone {
		return nil, domain.ErrOTPNotFound
	}

	if !bytes.Equal(
		candidateHash,
		f.codeHash,
	) {
		return nil, domain.ErrOTPInvalid
	}

	return &domain.PhoneVerification{
		ID:       "verification-id",
		Phone:    phone,
		CodeHash: candidateHash,
	}, nil
}

type fakeGuestRepository struct {
	user *domain.User
}

func (f *fakeGuestRepository) GetByPhone(
	ctx context.Context,
	phone string,
) (*domain.User, error) {
	if f.user == nil {
		return nil, domain.ErrUserNotFound
	}

	if f.user.Phone == nil ||
		*f.user.Phone != phone {
		return nil, domain.ErrUserNotFound
	}

	return f.user, nil
}

func (f *fakeGuestRepository) CreateGuest(
	ctx context.Context,
	phone string,
) (*domain.User, error) {
	f.user = &domain.User{
		ID:       "guest-id",
		Phone:    &phone,
		Role:     domain.RoleGuest,
		IsActive: true,
	}

	return f.user, nil
}

type fakeSessionIssuer struct {
	userID string
}

func (f *fakeSessionIssuer) CreateSession(
	ctx context.Context,
	userID string,
	userAgent *string,
	ipAddress *string,
) (string, error) {
	f.userID = userID

	return "session-token", nil
}

type fakeOTPSender struct {
	phone string
	code  string
}

func (f *fakeOTPSender) SendOTP(
	ctx context.Context,
	phone string,
	code string,
) error {
	f.phone = phone
	f.code = code

	return nil
}

func TestOTPServiceRequestAndVerify(t *testing.T) {
	secret := []byte(
		"01234567890123456789012345678901",
	)

	verifications := &fakeOTPRepository{}
	users := &fakeGuestRepository{}
	sessions := &fakeSessionIssuer{}
	sender := &fakeOTPSender{}

	service := NewOTPService(
		verifications,
		users,
		sessions,
		sender,
		secret,
	)

	phone := "+995555123456"

	err := service.RequestOTP(
		context.Background(),
		phone,
	)
	if err != nil {
		t.Fatalf(
			"RequestOTP() error: %v",
			err,
		)
	}

	if sender.phone != phone {
		t.Fatalf(
			"expected phone %q, got %q",
			phone,
			sender.phone,
		)
	}

	if !isValidOTPFormat(sender.code) {
		t.Fatalf(
			"invalid OTP generated: %q",
			sender.code,
		)
	}

	token, err := service.VerifyOTP(
		context.Background(),
		phone,
		sender.code,
		nil,
		nil,
	)
	if err != nil {
		t.Fatalf(
			"VerifyOTP() error: %v",
			err,
		)
	}

	if token != "session-token" {
		t.Fatalf(
			"unexpected token: %q",
			token,
		)
	}

	if users.user == nil {
		t.Fatal(
			"expected guest to be created",
		)
	}

	if users.user.Role != domain.RoleGuest {
		t.Fatalf(
			"expected guest role, got %q",
			users.user.Role,
		)
	}
}

func TestOTPServiceRejectsInvalidPhone(
	t *testing.T,
) {
	service := NewOTPService(
		&fakeOTPRepository{},
		&fakeGuestRepository{},
		&fakeSessionIssuer{},
		&fakeOTPSender{},
		[]byte(
			"01234567890123456789012345678901",
		),
	)

	err := service.RequestOTP(
		context.Background(),
		"555 123 456",
	)

	if err != domain.ErrInvalidPhone {
		t.Fatalf(
			"expected ErrInvalidPhone, got %v",
			err,
		)
	}
}

func TestOTPServiceRejectsWrongCode(
	t *testing.T,
) {
	secret := []byte(
		"01234567890123456789012345678901",
	)

	verifications := &fakeOTPRepository{}
	sender := &fakeOTPSender{}

	service := NewOTPService(
		verifications,
		&fakeGuestRepository{},
		&fakeSessionIssuer{},
		sender,
		secret,
	)

	phone := "+995555123456"

	if err := service.RequestOTP(
		context.Background(),
		phone,
	); err != nil {
		t.Fatalf(
			"RequestOTP() error: %v",
			err,
		)
	}

	_, err := service.VerifyOTP(
		context.Background(),
		phone,
		"000000",
		nil,
		nil,
	)

	if err != domain.ErrOTPInvalid {
		t.Fatalf(
			"expected ErrOTPInvalid, got %v",
			err,
		)
	}
}
