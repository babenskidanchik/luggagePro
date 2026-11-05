package handlers

import (
	"context"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/babenskidanchik/luggage-platform/backend/internal/domain"
	"github.com/babenskidanchik/luggage-platform/backend/internal/hotel"
	"github.com/babenskidanchik/luggage-platform/backend/internal/httpapi/middleware"
	"github.com/gin-gonic/gin"
)

type ModerationService interface {
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

type ModerationHandler struct {
	service ModerationService
}

type moderationApplicationResponse struct {
	ID              string                        `json:"id"`
	UserID          string                        `json:"user_id"`
	VenueType       domain.VenueType              `json:"venue_type"`
	LegalName       string                        `json:"legal_name"`
	DisplayName     string                        `json:"display_name"`
	TaxID           string                        `json:"tax_id"`
	CountryCode     string                        `json:"country_code"`
	City            string                        `json:"city"`
	Address         string                        `json:"address"`
	Status          domain.HotelApplicationStatus `json:"status"`
	RejectionReason *string                       `json:"rejection_reason,omitempty"`
	ReviewedBy      *string                       `json:"reviewed_by,omitempty"`
	ReviewedAt      *time.Time                    `json:"reviewed_at,omitempty"`
	CreatedAt       time.Time                     `json:"created_at"`
	UpdatedAt       time.Time                     `json:"updated_at"`
}

func NewModerationHandler(
	service ModerationService,
) *ModerationHandler {
	return &ModerationHandler{
		service: service,
	}
}

func (h *ModerationHandler) ListPending(
	c *gin.Context,
) {
	limit := 0
	offset := 0

	if raw := c.Query("limit"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil {
			c.JSON(
				http.StatusBadRequest,
				gin.H{"error": "invalid limit"},
			)
			return
		}

		limit = parsed
	}

	if raw := c.Query("offset"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil {
			c.JSON(
				http.StatusBadRequest,
				gin.H{"error": "invalid offset"},
			)
			return
		}

		offset = parsed
	}

	applications, err := h.service.ListPending(
		c.Request.Context(),
		limit,
		offset,
	)
	if err != nil {
		log.Printf(
			"list pending hotel applications failed: %v",
			err,
		)

		c.JSON(
			http.StatusInternalServerError,
			gin.H{
				"error": "internal server error",
			},
		)
		return
	}

	response := make(
		[]moderationApplicationResponse,
		0,
		len(applications),
	)

	for _, application := range applications {
		response = append(response, moderationApplicationResponse{
			ID:              application.ID,
			UserID:          application.UserID,
			VenueType:       application.VenueType,
			LegalName:       application.LegalName,
			DisplayName:     application.DisplayName,
			TaxID:           application.TaxID,
			CountryCode:     application.CountryCode,
			City:            application.City,
			Address:         application.Address,
			Status:          application.Status,
			RejectionReason: application.RejectionReason,
			ReviewedBy:      application.ReviewedBy,
			ReviewedAt:      application.ReviewedAt,
			CreatedAt:       application.CreatedAt,
			UpdatedAt:       application.UpdatedAt,
		})
	}

	c.JSON(
		http.StatusOK,
		gin.H{
			"applications": response,
			"count":        len(response),
		},
	)
}

func (h *ModerationHandler) Approve(
	c *gin.Context,
) {
	value, exists := c.Get(
		middleware.CurrentUserContextKey,
	)

	if !exists {
		c.JSON(
			http.StatusUnauthorized,
			gin.H{"error": "authentication required"},
		)
		return
	}

	user, ok := value.(*domain.User)
	if !ok || user == nil {
		c.JSON(
			http.StatusUnauthorized,
			gin.H{"error": "authentication required"},
		)
		return
	}

	applicationID := strings.TrimSpace(
		c.Param("id"),
	)

	if applicationID == "" {
		c.JSON(
			http.StatusBadRequest,
			gin.H{"error": "invalid application id"},
		)
		return
	}

	hotelResult, err := h.service.Approve(
		c.Request.Context(),
		applicationID,
		user.ID,
	)

	if err != nil {
		switch {
		case errors.Is(
			err,
			hotel.ErrInvalidModerationRequest,
		):
			c.JSON(
				http.StatusBadRequest,
				gin.H{"error": "invalid moderation request"},
			)

		case errors.Is(
			err,
			domain.ErrHotelApplicationNotFound,
		):
			c.JSON(
				http.StatusNotFound,
				gin.H{"error": "application not found"},
			)

		case errors.Is(
			err,
			domain.ErrHotelApplicationNotPending,
		):
			c.JSON(
				http.StatusConflict,
				gin.H{"error": "application is not pending"},
			)

		default:
			log.Printf(
				"approve hotel application failed: %v",
				err,
			)

			c.JSON(
				http.StatusInternalServerError,
				gin.H{"error": "internal server error"},
			)
		}

		return
	}

	c.JSON(
		http.StatusOK,
		gin.H{
			"hotel": hotelResult,
		},
	)
}
