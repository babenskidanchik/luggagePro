package domain

import "time"

type Hotel struct {
	ID            string
	UserID        string
	ApplicationID string
	DisplayName   string
	CountryCode   string
	City          string
	Address       string
	Latitude      *float64
	Longitude     *float64
	IsActive      bool
	CreatedAt     time.Time
	UpdatedAt     time.Time
}
