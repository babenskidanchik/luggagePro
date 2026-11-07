package domain

import "errors"

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrUserAlreadyExists = errors.New("user already exists")

	ErrSessionNotFound = errors.New("session not found")
	ErrUserInactive    = errors.New("user inactive")

	//opt errors
	ErrOTPNotFound        = errors.New("OTP verification not found")
	ErrOTPExpired         = errors.New("OTP expired")
	ErrOTPInvalid         = errors.New("invalid OTP")
	ErrOTPTooManyAttempts = errors.New("too many OTP attemps")
	ErrOTPCooldown        = errors.New("OTP cooldown active")

	ErrInvalidPhone = errors.New("invalid phone")
	ErrPhoneInUse   = errors.New("phone already used by another account type")

	ErrHotelApplicationAlreadyExists = errors.New(
		"hotel application already exists",
	)
	ErrHotelTaxIDAlreadyInUse = errors.New(
		"hotel tax id already in use",
	)
	ErrHotelApplicationNotFound = errors.New(
		"hotel application not found",
	)

	ErrHotelApplicationNotPending = errors.New(
		"hotel application is not pending",
	)

	ErrHotelNotFound = errors.New(
		"hotel not found",
	)

	ErrInvalidHotelLocation = errors.New(
		"invalid hotel location",
	)

	ErrHotelUnavailable = errors.New(
		"hotel is unavailable",
	)

	ErrInvalidBooking = errors.New(
		"invalid booking",
	)

	ErrBookingCapacityExceeded = errors.New(
		"booking luggage capacity exceeded",
	)
	ErrBookingNotFound = errors.New("booking not found")
	ErrBookingInvalidTransition = errors.New("invalid booking status transition")
	ErrBookingCancellationTooLate = errors.New("booking can no longer be cancelled")
)
