package domain

import "time"

type Hotel struct {
	ID            string
	UserID        string
	ApplicationID string

	VenueType VenueType

	DisplayName string
	CountryCode string
	City        string
	Address     string

	Latitude  *float64
	Longitude *float64

	LuggageCapacity  int
	AcceptingBookings bool

	IsActive  bool
	CreatedAt time.Time
	UpdatedAt time.Time
}
