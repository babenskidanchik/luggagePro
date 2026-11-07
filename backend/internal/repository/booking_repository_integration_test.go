package repository

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/babenskidanchik/luggage-platform/backend/internal/domain"
	"github.com/babenskidanchik/luggage-platform/backend/internal/migrate"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestBookingRepositoryCreateIntegration(t *testing.T) {
	ctx, db := newBookingTestDB(t)

	guest1 := createBookingTestUser(t, ctx, db, "guest1@example.com", "guest")
	guest2 := createBookingTestUser(t, ctx, db, "guest2@example.com", "guest")
	guest3 := createBookingTestUser(t, ctx, db, "guest3@example.com", "guest")
	partner := createBookingTestUser(t, ctx, db, "partner@example.com", "hotel")

	hotelID := createBookingTestHotel(t, ctx, db, partner, 3)

	repo := NewBookingRepository(db)

	start := time.Now().UTC().Add(24 * time.Hour).Truncate(time.Second)

	// 1. Успешное бронирование.
	first, err := repo.Create(
		ctx,
		guest1,
		hotelID,
		2,
		start,
		start.Add(2*time.Hour),
	)
	if err != nil {
		t.Fatalf("expected first booking to succeed: %v", err)
	}

	if first.Status != domain.BookingStatusConfirmed {
		t.Fatalf("expected confirmed status, got %q", first.Status)
	}

	// 2. Пересекающееся бронирование превышает вместимость:
	// занято 2 места, запрашивается ещё 2 при capacity = 3.
	_, err = repo.Create(
		ctx,
		guest2,
		hotelID,
		2,
		start.Add(time.Hour),
		start.Add(3*time.Hour),
	)
	if !errors.Is(err, domain.ErrBookingCapacityExceeded) {
		t.Fatalf("expected capacity error, got %v", err)
	}

	// 3. Соседние интервалы не пересекаются.
	_, err = repo.Create(
		ctx,
		guest2,
		hotelID,
		3,
		start.Add(2*time.Hour),
		start.Add(4*time.Hour),
	)
	if err != nil {
		t.Fatalf("expected adjacent booking to succeed: %v", err)
	}

	// 4. Проверяем, что пересечение обоих интервалов учитывается.
	// В периоде 11:30–12:30 пересекаются обе предыдущие брони:
	// 2 + 3 места, поэтому новое бронирование невозможно.
	_, err = repo.Create(
		ctx,
		guest3,
		hotelID,
		1,
		start.Add(90*time.Minute),
		start.Add(150*time.Minute),
	)
	if !errors.Is(err, domain.ErrBookingCapacityExceeded) {
		t.Fatalf("expected capacity error for overlapping bookings, got %v", err)
	}
}

func TestBookingRepositoryConcurrentCapacityIntegration(t *testing.T) {
	ctx, db := newBookingTestDB(t)

	guest1 := createBookingTestUser(t, ctx, db, "concurrent1@example.com", "guest")
	guest2 := createBookingTestUser(t, ctx, db, "concurrent2@example.com", "guest")
	partner := createBookingTestUser(t, ctx, db, "concurrent-partner@example.com", "hotel")

	hotelID := createBookingTestHotel(t, ctx, db, partner, 1)
	repo := NewBookingRepository(db)

	start := time.Now().UTC().Add(48 * time.Hour).Truncate(time.Second)
	end := start.Add(2 * time.Hour)

	var wg sync.WaitGroup
	wg.Add(2)

	results := make(chan error, 2)

	for _, guestID := range []string{guest1, guest2} {
		go func(userID string) {
			defer wg.Done()

			_, err := repo.Create(
				ctx,
				userID,
				hotelID,
				1,
				start,
				end,
			)

			results <- err
		}(guestID)
	}

	wg.Wait()
	close(results)

	successCount := 0
	capacityErrorCount := 0

	for err := range results {
		switch {
		case err == nil:
			successCount++
		case errors.Is(err, domain.ErrBookingCapacityExceeded):
			capacityErrorCount++
		default:
			t.Fatalf("unexpected booking error: %v", err)
		}
	}

	if successCount != 1 {
		t.Fatalf("expected exactly 1 successful booking, got %d", successCount)
	}

	if capacityErrorCount != 1 {
		t.Fatalf("expected exactly 1 capacity error, got %d", capacityErrorCount)
	}

	var totalLuggage int

	err := db.QueryRow(
		ctx,
		`
			SELECT COALESCE(SUM(luggage_count), 0)
			FROM bookings
			WHERE hotel_id = $1::uuid
			  AND status IN ('pending', 'confirmed')
		`,
		hotelID,
	).Scan(&totalLuggage)
	if err != nil {
		t.Fatalf("read booked luggage count: %v", err)
	}

	if totalLuggage != 1 {
		t.Fatalf("expected total luggage count 1, got %d", totalLuggage)
	}
}

func TestBookingRepositoryRejectsWhenHotelPausedIntegration(t *testing.T) {
	ctx, db := newBookingTestDB(t)

	guest := createBookingTestUser(t, ctx, db, "paused-guest@example.com", "guest")
	partner := createBookingTestUser(t, ctx, db, "paused-partner@example.com", "hotel")
	hotelID := createBookingTestHotel(t, ctx, db, partner, 5)

	hotelRepo := NewHotelRepository(db)
	_, err := hotelRepo.UpdateBookingSettingsByUserID(
		ctx,
		partner,
		5,
		false,
	)
	if err != nil {
		t.Fatalf("pause hotel bookings: %v", err)
	}

	start := time.Now().UTC().Add(24 * time.Hour).Truncate(time.Second)
	_, err = NewBookingRepository(db).Create(
		ctx,
		guest,
		hotelID,
		1,
		start,
		start.Add(time.Hour),
	)
	if !errors.Is(err, domain.ErrHotelUnavailable) {
		t.Fatalf("expected hotel unavailable error, got %v", err)
	}
}

func TestBookingRepositoryListsOnlyBookingsForOwningHotelIntegration(t *testing.T) {
	ctx, db := newBookingTestDB(t)

	guest := createBookingTestUser(t, ctx, db, "hotel-list-guest@example.com", "guest")
	hotelOwner1 := createBookingTestUser(t, ctx, db, "hotel-list-owner1@example.com", "hotel")
	hotelOwner2 := createBookingTestUser(t, ctx, db, "hotel-list-owner2@example.com", "hotel")

	hotel1 := createBookingTestHotel(t, ctx, db, hotelOwner1, 10)
	hotel2 := createBookingTestHotel(t, ctx, db, hotelOwner2, 10)

	start := time.Now().UTC().Add(24 * time.Hour).Truncate(time.Second)
	repo := NewBookingRepository(db)

	created, err := repo.Create(
		ctx,
		guest,
		hotel1,
		2,
		start,
		start.Add(2*time.Hour),
	)
	if err != nil {
		t.Fatalf("create booking for first hotel: %v", err)
	}

	_, err = repo.Create(
		ctx,
		guest,
		hotel2,
		1,
		start,
		start.Add(time.Hour),
	)
	if err != nil {
		t.Fatalf("create booking for second hotel: %v", err)
	}

	firstHotelBookings, err := repo.ListByHotelUserID(ctx, hotelOwner1, 50, 0)
	if err != nil {
		t.Fatalf("list first hotel bookings: %v", err)
	}
	if len(firstHotelBookings) != 1 {
		t.Fatalf("expected 1 booking for first hotel, got %d", len(firstHotelBookings))
	}
	if firstHotelBookings[0].ID != created.ID {
		t.Fatalf("expected booking %q, got %q", created.ID, firstHotelBookings[0].ID)
	}
	if firstHotelBookings[0].HotelID != hotel1 {
		t.Fatalf("expected hotel %q, got %q", hotel1, firstHotelBookings[0].HotelID)
	}

	secondHotelBookings, err := repo.ListByHotelUserID(ctx, hotelOwner2, 50, 0)
	if err != nil {
		t.Fatalf("list second hotel bookings: %v", err)
	}
	if len(secondHotelBookings) != 1 {
		t.Fatalf("expected 1 booking for second hotel, got %d", len(secondHotelBookings))
	}
	if secondHotelBookings[0].HotelID != hotel2 {
		t.Fatalf("expected only bookings for hotel %q, got %q", hotel2, secondHotelBookings[0].HotelID)
	}
}

func TestBookingRepositoryHotelBaggageLifecycleIntegration(t *testing.T) {
	ctx, db := newBookingTestDB(t)

	guest := createBookingTestUser(t, ctx, db, "baggage-lifecycle-guest@example.com", "guest")
	hotelOwner := createBookingTestUser(t, ctx, db, "baggage-lifecycle-owner@example.com", "hotel")
	otherHotelOwner := createBookingTestUser(t, ctx, db, "baggage-lifecycle-other@example.com", "hotel")
	hotelID := createBookingTestHotel(t, ctx, db, hotelOwner, 1)
	createBookingTestHotel(t, ctx, db, otherHotelOwner, 5)

	start := time.Now().UTC().Add(24 * time.Hour).Truncate(time.Second)
	repo := NewBookingRepository(db)
	created, err := repo.Create(ctx, guest, hotelID, 1, start, start.Add(time.Hour))
	if err != nil {
		t.Fatalf("create booking: %v", err)
	}

	_, err = repo.TransitionStatusForHotelUserID(
		ctx,
		otherHotelOwner,
		created.ID,
		domain.BookingStatusConfirmed,
		domain.BookingStatusBaggageReceived,
	)
	if !errors.Is(err, domain.ErrBookingNotFound) {
		t.Fatalf("expected other hotel to be unable to access booking, got %v", err)
	}

	_, err = repo.TransitionStatusForHotelUserID(
		ctx,
		hotelOwner,
		created.ID,
		domain.BookingStatusBaggageReceived,
		domain.BookingStatusCompleted,
	)
	if !errors.Is(err, domain.ErrBookingInvalidTransition) {
		t.Fatalf("expected completion before receipt to fail, got %v", err)
	}

	received, err := repo.TransitionStatusForHotelUserID(
		ctx,
		hotelOwner,
		created.ID,
		domain.BookingStatusConfirmed,
		domain.BookingStatusBaggageReceived,
	)
	if err != nil {
		t.Fatalf("confirm baggage receipt: %v", err)
	}
	if received.Status != domain.BookingStatusBaggageReceived {
		t.Fatalf("expected baggage_received status, got %q", received.Status)
	}

	_, err = repo.Create(
		ctx,
		guest,
		hotelID,
		1,
		start,
		start.Add(time.Hour),
	)
	if !errors.Is(err, domain.ErrBookingCapacityExceeded) {
		t.Fatalf("expected received baggage to occupy capacity, got %v", err)
	}

	_, err = repo.TransitionStatusForHotelUserID(
		ctx,
		hotelOwner,
		created.ID,
		domain.BookingStatusConfirmed,
		domain.BookingStatusBaggageReceived,
	)
	if !errors.Is(err, domain.ErrBookingInvalidTransition) {
		t.Fatalf("expected duplicate receipt confirmation to fail, got %v", err)
	}

	completed, err := repo.TransitionStatusForHotelUserID(
		ctx,
		hotelOwner,
		created.ID,
		domain.BookingStatusBaggageReceived,
		domain.BookingStatusCompleted,
	)
	if err != nil {
		t.Fatalf("confirm baggage release: %v", err)
	}
	if completed.Status != domain.BookingStatusCompleted {
		t.Fatalf("expected completed status, got %q", completed.Status)
	}

	_, err = repo.Create(
		ctx,
		guest,
		hotelID,
		1,
		start,
		start.Add(time.Hour),
	)
	if err != nil {
		t.Fatalf("expected completed booking to release capacity: %v", err)
	}
}

func newBookingTestDB(t *testing.T) (context.Context, *pgxpool.Pool) {
	t.Helper()

	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL is not set")
	}

	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		t.Fatalf("parse test database URL: %v", err)
	}

	if config.ConnConfig.Database != "luggage_test" {
		t.Fatalf("refusing to run against database %q", config.ConnConfig.Database)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	t.Cleanup(cancel)

	db, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		t.Fatalf("connect test database: %v", err)
	}

	t.Cleanup(db.Close)

	if err := db.Ping(ctx); err != nil {
		t.Fatalf("ping test database: %v", err)
	}

	if err := migrate.Run(ctx, db); err != nil {
		t.Fatalf("run migrations: %v", err)
	}

	_, err = db.Exec(ctx, `
		TRUNCATE TABLE
			bookings,
			hotels,
			hotel_applications,
			sessions,
			phone_verifications,
			users
		CASCADE
	`)
	if err != nil {
		t.Fatalf("truncate test database: %v", err)
	}

	return ctx, db
}

func createBookingTestUser(
	t *testing.T,
	ctx context.Context,
	db *pgxpool.Pool,
	email string,
	role string,
) string {
	t.Helper()

	var userID string

	err := db.QueryRow(
		ctx,
		`
			INSERT INTO users (email, role)
			VALUES ($1, $2)
			RETURNING id::text
		`,
		email,
		role,
	).Scan(&userID)
	if err != nil {
		t.Fatalf("create test user: %v", err)
	}

	return userID
}

func createBookingTestHotel(
	t *testing.T,
	ctx context.Context,
	db *pgxpool.Pool,
	partnerID string,
	capacity int,
) string {
	t.Helper()

	var applicationID string

	err := db.QueryRow(
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
				$1::uuid,
				'luggage_shop',
				$2,
				$3,
				$4,
				'GE',
				'Batumi',
				'Integration test address'
			)
			RETURNING id::text
		`,
		partnerID,
		"Test Company "+partnerID,
		"Test Storage "+partnerID,
		fmt.Sprintf("TEST-%s", partnerID),
	).Scan(&applicationID)
	if err != nil {
		t.Fatalf("create test application: %v", err)
	}

	var hotelID string

	err = db.QueryRow(
		ctx,
		`
			INSERT INTO hotels (
				user_id,
				application_id,
				venue_type,
				display_name,
				country_code,
				city,
				address,
				latitude,
				longitude,
				is_active,
				luggage_capacity
			)
			VALUES (
				$1::uuid,
				$2::uuid,
				'luggage_shop',
				'Integration Test Storage',
				'GE',
				'Batumi',
				'Integration test address',
				41.6168,
				41.6367,
				TRUE,
				$3
			)
			RETURNING id::text
		`,
		partnerID,
		applicationID,
		capacity,
	).Scan(&hotelID)
	if err != nil {
		t.Fatalf("create test hotel: %v", err)
	}

	return hotelID
}
