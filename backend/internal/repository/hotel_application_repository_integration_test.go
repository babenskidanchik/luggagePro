package repository

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/babenskidanchik/luggage-platform/backend/internal/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestHotelApplicationApproveIntegration(
	t *testing.T,
) {
	databaseURL := strings.TrimSpace(
		os.Getenv("TEST_DATABASE_URL"),
	)

	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL is not set")
	}

	// Security gate: integration test must never touch
	// the development/production database.
	if !strings.Contains(
		databaseURL,
		"luggage_test",
	) {
		t.Fatalf(
			"refusing to run against non-test database",
		)
	}

	ctx, cancel := context.WithTimeout(
		context.Background(),
		15*time.Second,
	)
	defer cancel()

	db, err := pgxpool.New(
		ctx,
		databaseURL,
	)
	if err != nil {
		t.Fatalf(
			"connect test database: %v",
			err,
		)
	}
	defer db.Close()

	if err := db.Ping(ctx); err != nil {
		t.Fatalf(
			"ping test database: %v",
			err,
		)
	}

	// Isolate this test from existing test data.
	_, err = db.Exec(
		ctx,
		`
			TRUNCATE TABLE
				hotels,
				hotel_applications,
				sessions,
				phone_verifications,
				users
			CASCADE
		`,
	)
	if err != nil {
		t.Fatalf(
			"truncate test database: %v",
			err,
		)
	}

	var guestID string

	err = db.QueryRow(
		ctx,
		`
			INSERT INTO users (
				email,
				role
			)
			VALUES (
				$1,
				'guest'
			)
			RETURNING CAST(id AS text)
		`,
		"approve-test-guest@example.com",
	).Scan(&guestID)
	if err != nil {
		t.Fatalf(
			"create guest: %v",
			err,
		)
	}

	var moderatorID string

	err = db.QueryRow(
		ctx,
		`
			INSERT INTO users (
				email,
				role
			)
			VALUES (
				$1,
				'moderator'
			)
			RETURNING CAST(id AS text)
		`,
		"approve-test-moderator@example.com",
	).Scan(&moderatorID)
	if err != nil {
		t.Fatalf(
			"create moderator: %v",
			err,
		)
	}

	var applicationID string

	err = db.QueryRow(
		ctx,
		`
			INSERT INTO hotel_applications (
				user_id,
				venue_type,
				legal_name,
				display_name,
				tax_id,
				country_code,
				city,
				address
			)
			VALUES (
				CAST($1 AS uuid),
				'luggage_shop',
				$2,
				$3,
				$4,
				$5,
				$6,
				$7
			)
			RETURNING CAST(id AS text)
		`,
		guestID,
		"Integration Storage LLC",
		"Integration Luggage Spot",
		"TEST-APPROVE-001",
		"GE",
		"Batumi",
		"1 Integration Street",
	).Scan(&applicationID)
	if err != nil {
		t.Fatalf(
			"create hotel application: %v",
			err,
		)
	}

	repository :=
		NewHotelApplicationRepository(db)

	hotel, err := repository.Approve(
		ctx,
		applicationID,
		moderatorID,
	)
	if err != nil {
		t.Fatalf(
			"approve application: %v",
			err,
		)
	}

	if hotel == nil {
		t.Fatal("expected hotel")
	}

	if hotel.UserID != guestID {
		t.Fatalf(
			"expected hotel user %q, got %q",
			guestID,
			hotel.UserID,
		)
	}

	if hotel.VenueType != domain.VenueTypeLuggageShop {
		t.Fatalf(
			"expected venue type %q, got %q",
			domain.VenueTypeLuggageShop,
			hotel.VenueType,
		)
	}

	if hotel.ApplicationID != applicationID {
		t.Fatalf(
			"expected application %q, got %q",
			applicationID,
			hotel.ApplicationID,
		)
	}

	if !hotel.IsActive {
		t.Fatal("expected hotel to be active")
	}

	var (
		status     domain.HotelApplicationStatus
		reviewedBy string
	)

	err = db.QueryRow(
		ctx,
		`
			SELECT
				status,
				CAST(reviewed_by AS text)
			FROM hotel_applications
			WHERE id = CAST($1 AS uuid)
		`,
		applicationID,
	).Scan(
		&status,
		&reviewedBy,
	)
	if err != nil {
		t.Fatalf(
			"read approved application: %v",
			err,
		)
	}

	if status != domain.HotelApplicationApproved {
		t.Fatalf(
			"expected approved status, got %q",
			status,
		)
	}

	if reviewedBy != moderatorID {
		t.Fatalf(
			"expected moderator %q, got %q",
			moderatorID,
			reviewedBy,
		)
	}

	var role domain.UserRole

	err = db.QueryRow(
		ctx,
		`
			SELECT role
			FROM users
			WHERE id = CAST($1 AS uuid)
		`,
		guestID,
	).Scan(&role)
	if err != nil {
		t.Fatalf(
			"read promoted user: %v",
			err,
		)
	}

	if role != domain.RoleHotel {
		t.Fatalf(
			"expected hotel role, got %q",
			role,
		)
	}

	var hotelCount int

	err = db.QueryRow(
		ctx,
		`
			SELECT COUNT(*)
			FROM hotels
			WHERE application_id = CAST($1 AS uuid)
		`,
		applicationID,
	).Scan(&hotelCount)
	if err != nil {
		t.Fatalf(
			"count hotels: %v",
			err,
		)
	}

	if hotelCount != 1 {
		t.Fatalf(
			"expected exactly 1 hotel, got %d",
			hotelCount,
		)
	}

	// Same application cannot be approved twice.
	_, err = repository.Approve(
		ctx,
		applicationID,
		moderatorID,
	)

	if !errors.Is(
		err,
		domain.ErrHotelApplicationNotPending,
	) {
		t.Fatalf(
			"expected ErrHotelApplicationNotPending, got %v",
			err,
		)
	}

	err = db.QueryRow(
		ctx,
		`
			SELECT COUNT(*)
			FROM hotels
			WHERE application_id = CAST($1 AS uuid)
		`,
		applicationID,
	).Scan(&hotelCount)
	if err != nil {
		t.Fatalf(
			"count hotels after second approve: %v",
			err,
		)
	}

	if hotelCount != 1 {
		t.Fatalf(
			"expected still 1 hotel, got %d",
			hotelCount,
		)
	}
}
