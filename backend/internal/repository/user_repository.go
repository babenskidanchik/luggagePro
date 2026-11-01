package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/babenskidanchik/luggage-platform/backend/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository struct {
	db     *pgxpool.Pool
	phones PhoneProtector
}

func NewUserRepository(
	db *pgxpool.Pool,
	phones PhoneProtector,
) *UserRepository {
	return &UserRepository{
		db:     db,
		phones: phones,
	}
}

func (r *UserRepository) CreateGuest(
	ctx context.Context,
	phone string,
) (*domain.User, error) {
	encryptedPhone, err := r.phones.Encrypt(phone)
	if err != nil {
		return nil, fmt.Errorf(
			"encrypt guest phone: %w",
			err,
		)
	}

	lookupHash := r.phones.LookupHash(phone)

	const query = `
		INSERT INTO users (
			phone_encrypted,
			phone_lookup_hash,
			role
		)
		VALUES (
			$1,
			$2,
			'guest'
		)
		RETURNING
			CAST(id AS text),
			phone_encrypted,
			email,
			role,
			is_active,
			created_at,
			updated_at
	`

	var user domain.User
	var storedEncryptedPhone []byte

	err = r.db.QueryRow(
		ctx,
		query,
		encryptedPhone,
		lookupHash,
	).Scan(
		&user.ID,
		&storedEncryptedPhone,
		&user.Email,
		&user.Role,
		&user.IsActive,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) &&
			pgErr.Code == "23505" {
			return nil, domain.ErrUserAlreadyExists
		}

		return nil, fmt.Errorf(
			"create guest: %w",
			err,
		)
	}

	if err := r.setDecryptedPhone(
		&user,
		storedEncryptedPhone,
	); err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *UserRepository) GetByID(
	ctx context.Context,
	id string,
) (*domain.User, error) {
	const query = `
		SELECT
			CAST(id AS text),
			phone_encrypted,
			email,
			role,
			is_active,
			created_at,
			updated_at
		FROM users
		WHERE id = $1
	`

	var user domain.User
	var encryptedPhone []byte

	err := r.db.QueryRow(
		ctx,
		query,
		id,
	).Scan(
		&user.ID,
		&encryptedPhone,
		&user.Email,
		&user.Role,
		&user.IsActive,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}

		return nil, fmt.Errorf(
			"get user by id: %w",
			err,
		)
	}

	if err := r.setDecryptedPhone(
		&user,
		encryptedPhone,
	); err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *UserRepository) GetByPhone(
	ctx context.Context,
	phone string,
) (*domain.User, error) {
	lookupHash := r.phones.LookupHash(phone)

	const query = `
		SELECT
			CAST(id AS text),
			phone_encrypted,
			email,
			role,
			is_active,
			created_at,
			updated_at
		FROM users
		WHERE phone_lookup_hash = $1
	`

	var user domain.User
	var encryptedPhone []byte

	err := r.db.QueryRow(
		ctx,
		query,
		lookupHash,
	).Scan(
		&user.ID,
		&encryptedPhone,
		&user.Email,
		&user.Role,
		&user.IsActive,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}

		return nil, fmt.Errorf(
			"get user by phone hash: %w",
			err,
		)
	}

	if err := r.setDecryptedPhone(
		&user,
		encryptedPhone,
	); err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *UserRepository) setDecryptedPhone(
	user *domain.User,
	encrypted []byte,
) error {
	if len(encrypted) == 0 {
		user.Phone = nil
		return nil
	}

	phone, err := r.phones.Decrypt(encrypted)
	if err != nil {
		return fmt.Errorf(
			"decrypt user phone: %w",
			err,
		)
	}

	user.Phone = &phone

	return nil
}
