package hotel

import (
	"context"
	"errors"
	"testing"

	"github.com/babenskidanchik/luggage-platform/backend/internal/domain"
)

type fakeModerationRepository struct {
	limit  int
	offset int

	approvedApplicationID string
	moderatorID           string

	approveResult *domain.Hotel
	approveErr    error
}

func (r *fakeModerationRepository) ListPending(
	_ context.Context,
	limit int,
	offset int,
) ([]domain.HotelApplication, error) {
	r.limit = limit
	r.offset = offset

	return []domain.HotelApplication{
		{
			ID:          "application-id",
			DisplayName: "Batumi Luggage Spot",
			Status:      domain.HotelApplicationPending,
		},
	}, nil
}

func TestModerationServiceListPending(
	t *testing.T,
) {
	repository := &fakeModerationRepository{}

	service := NewModerationService(
		repository,
	)

	applications, err := service.ListPending(
		context.Background(),
		25,
		10,
	)
	if err != nil {
		t.Fatalf(
			"list pending: %v",
			err,
		)
	}

	if repository.limit != 25 {
		t.Fatalf(
			"expected limit 25, got %d",
			repository.limit,
		)
	}

	if repository.offset != 10 {
		t.Fatalf(
			"expected offset 10, got %d",
			repository.offset,
		)
	}

	if len(applications) != 1 {
		t.Fatalf(
			"expected 1 application, got %d",
			len(applications),
		)
	}
}

func TestModerationServiceCapsLimit(
	t *testing.T,
) {
	repository := &fakeModerationRepository{}

	service := NewModerationService(
		repository,
	)

	_, err := service.ListPending(
		context.Background(),
		10000,
		-5,
	)
	if err != nil {
		t.Fatalf(
			"list pending: %v",
			err,
		)
	}

	if repository.limit != maxPendingLimit {
		t.Fatalf(
			"expected max limit %d, got %d",
			maxPendingLimit,
			repository.limit,
		)
	}

	if repository.offset != 0 {
		t.Fatalf(
			"expected offset 0, got %d",
			repository.offset,
		)
	}
}

func (r *fakeModerationRepository) Approve(
	_ context.Context,
	applicationID string,
	moderatorID string,
) (*domain.Hotel, error) {
	r.approvedApplicationID = applicationID
	r.moderatorID = moderatorID

	if r.approveErr != nil {
		return nil, r.approveErr
	}

	return r.approveResult, nil
}

func TestModerationServiceApprove(t *testing.T) {
	repository := &fakeModerationRepository{
		approveResult: &domain.Hotel{
			ID:          "hotel-id",
			DisplayName: "Batumi Luggage Spot",
			IsActive:    true,
		},
	}

	service := NewModerationService(
		repository,
	)

	result, err := service.Approve(
		context.Background(),
		" application-id ",
		" moderator-id ",
	)
	if err != nil {
		t.Fatalf(
			"approve application: %v",
			err,
		)
	}

	if repository.approvedApplicationID != "application-id" {
		t.Fatalf(
			"unexpected application id: %q",
			repository.approvedApplicationID,
		)
	}

	if repository.moderatorID != "moderator-id" {
		t.Fatalf(
			"unexpected moderator id: %q",
			repository.moderatorID,
		)
	}

	if result == nil ||
		result.ID != "hotel-id" {
		t.Fatalf(
			"unexpected hotel result",
		)
	}
}

func TestModerationServiceApproveRejectsMissingID(t *testing.T) {
	repository := &fakeModerationRepository{}

	service := NewModerationService(
		repository,
	)

	_, err := service.Approve(
		context.Background(),
		"",
		"moderator-id",
	)

	if !errors.Is(
		err,
		ErrInvalidModerationRequest,
	) {
		t.Fatalf(
			"expected invalid moderation request, got %v",
			err,
		)
	}
}
