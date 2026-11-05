package domain

import "time"

type HotelApplicationStatus string

const (
	HotelApplicationPending  HotelApplicationStatus = "pending"
	HotelApplicationApproved HotelApplicationStatus = "approved"
	HotelApplicationRejected HotelApplicationStatus = "rejected"
)

type HotelApplication struct {
	ID     string
	UserID string

	VenueType VenueType

	LegalName   string
	DisplayName string
	TaxID       string

	CountryCode string
	City        string
	Address     string

	Status          HotelApplicationStatus
	RejectionReason *string
	ReviewedBy      *string
	ReviewedAt      *time.Time

	CreatedAt time.Time
	UpdatedAt time.Time
}
