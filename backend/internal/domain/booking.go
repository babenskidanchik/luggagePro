package domain

import "time"

type BookingStatus string

const (
	BookingStatusPending BookingStatus = "pending"

	BookingStatusConfirmed BookingStatus = "confirmed"

	BookingStatusCancelled BookingStatus = "cancelled"

	BookingStatusCompleted BookingStatus = "completed"
)

type Booking struct {
	ID string

	GuestUserID string
	HotelID     string

	LuggageCount int

	StartsAt time.Time
	EndsAt   time.Time

	Status BookingStatus

	CreatedAt time.Time
	UpdatedAt time.Time
}
