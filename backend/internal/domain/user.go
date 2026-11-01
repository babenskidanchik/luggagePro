package domain

import "time"

type UserRole string

const (
	RoleGuest     UserRole = "guest"
	RoleHotel     UserRole = "hotel"
	RoleModerator UserRole = "moderator"
)

type User struct {
	ID        string
	Phone     *string
	Email     *string
	Role      UserRole
	IsActive  bool
	CreatedAt time.Time
	UpdatedAt time.Time
}
