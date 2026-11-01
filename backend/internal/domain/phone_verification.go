package domain

import "time"

type PhoneVerification struct {
	ID         string
	Phone      string
	CodeHash   []byte
	Attempts   int16
	ExpiresAt  time.Time
	ConsumedAt *time.Time
	CreatedAt  time.Time
}
