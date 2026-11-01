package hotel

import (
	"context"
	"errors"
	"strings"

	"github.com/babenskidanchik/luggage-platform/backend/internal/domain"
)

const (
	defaultPendingLimit = 50
	maxPendingLimit     = 100
)

var ErrInvalidModerationRequest = errors.New(
	"invalid moderation request",
)

type ModerationRepository interface {
	ListPending(
		ctx context.Context,
		limit int,
		offset int,
	) ([]domain.HotelApplication, error)

	Approve(
		ctx context.Context,
		applicationID string,
		moderatorID string,
	) (*domain.Hotel, error)
}

type ModerationService struct {
	repository ModerationRepository
}

func NewModerationService(
	repository ModerationRepository,
) *ModerationService {
	return &ModerationService{
		repository: repository,
	}
}

func (s *ModerationService) ListPending(
	ctx context.Context,
	limit int,
	offset int,
) ([]domain.HotelApplication, error) {
	if limit <= 0 {
		limit = defaultPendingLimit
	}

	if limit > maxPendingLimit {
		limit = maxPendingLimit
	}

	if offset < 0 {
		offset = 0
	}

	return s.repository.ListPending(
		ctx,
		limit,
		offset,
	)
}

func (s *ModerationService) Approve(
	ctx context.Context,
	applicationID string,
	moderatorID string,
) (*domain.Hotel, error) {
	applicationID = strings.TrimSpace(
		applicationID,
	)

	moderatorID = strings.TrimSpace(
		moderatorID,
	)

	if applicationID == "" ||
		moderatorID == "" {
		return nil, ErrInvalidModerationRequest
	}

	return s.repository.Approve(
		ctx,
		applicationID,
		moderatorID,
	)

}
