package handlers

import (
	"context"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/babenskidanchik/luggage-platform/backend/internal/domain"
	"github.com/babenskidanchik/luggage-platform/backend/internal/httpapi/middleware"
	"github.com/gin-gonic/gin"
)

type BookingService interface {
	Create(
		ctx context.Context,
		guestUserID string,
		hotelID string,
		luggageCount int,
		startsAt time.Time,
		endsAt time.Time,
	) (*domain.Booking, error)
}

type BookingHandler struct {
	service BookingService
}

func NewBookingHandler(
	service BookingService,
) *BookingHandler {
	return &BookingHandler{
		service: service,
	}
}

type createBookingRequest struct {
	HotelID      string    `json:"hotel_id"`
	LuggageCount int       `json:"luggage_count"`
	StartsAt     time.Time `json:"starts_at"`
	EndsAt       time.Time `json:"ends_at"`
}

func (h *BookingHandler) Create(c *gin.Context) {
	user, ok := middleware.CurrentUser(c)
	if !ok {
		c.JSON(
			http.StatusUnauthorized,
			gin.H{"error": "authentication required"},
		)
		return
	}

	var request createBookingRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(
			http.StatusBadRequest,
			gin.H{"error": "invalid request body"},
		)
		return
	}

	result, err := h.service.Create(
		c.Request.Context(),
		user.ID,
		request.HotelID,
		request.LuggageCount,
		request.StartsAt,
		request.EndsAt,
	)

	if err != nil {
		switch {
		case errors.Is(err, domain.ErrInvalidBooking):
			c.JSON(
				http.StatusBadRequest,
				gin.H{"error": "invalid booking"},
			)

		case errors.Is(err, domain.ErrBookingCapacityExceeded):
			c.JSON(
				http.StatusConflict,
				gin.H{"error": "insufficient luggage capacity"},
			)

		case errors.Is(err, domain.ErrHotelUnavailable):
			c.JSON(
				http.StatusConflict,
				gin.H{"error": "storage location unavailable"},
			)

		default:
			log.Printf("create booking failed: %v", err)

			c.JSON(
				http.StatusInternalServerError,
				gin.H{"error": "internal server error"},
			)
		}

		return
	}

	c.JSON(
		http.StatusCreated,
		gin.H{
			"booking": gin.H{
				"id":            result.ID,
				"hotel_id":      result.HotelID,
				"luggage_count": result.LuggageCount,
				"starts_at":     result.StartsAt,
				"ends_at":       result.EndsAt,
				"status":        result.Status,
				"created_at":    result.CreatedAt,
			},
		},
	)
}
