package handlers

import (
	"context"
	"errors"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/babenskidanchik/luggage-platform/backend/internal/domain"
	"github.com/babenskidanchik/luggage-platform/backend/internal/httpapi/middleware"
	"github.com/gin-gonic/gin"
)

type BookingService interface {
	ListForGuest(
		ctx context.Context,
		guestUserID string,
		limit int,
		offset int,
	) ([]domain.Booking, error)
	CancelForGuest(
		ctx context.Context,
		guestUserID string,
		bookingID string,
	) (*domain.Booking, error)
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

func (h *BookingHandler) List(c *gin.Context) {
	user, ok := middleware.CurrentUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	limit, err := strconv.Atoi(c.DefaultQuery("limit", "50"))
	if err != nil || limit < 1 || limit > 100 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "limit must be between 1 and 100"})
		return
	}
	offset, err := strconv.Atoi(c.DefaultQuery("offset", "0"))
	if err != nil || offset < 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "offset must be zero or greater"})
		return
	}

	items, err := h.service.ListForGuest(c.Request.Context(), user.ID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	bookings := make([]gin.H, 0, len(items))
	for _, item := range items {
		bookings = append(bookings, bookingResponse(item))
	}
	c.JSON(http.StatusOK, gin.H{
		"bookings": bookings,
		"limit": limit,
		"offset": offset,
	})
}

func (h *BookingHandler) Cancel(c *gin.Context) {
	user, ok := middleware.CurrentUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	result, err := h.service.CancelForGuest(
		c.Request.Context(),
		user.ID,
		c.Param("id"),
	)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrInvalidBooking):
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid booking"})
		case errors.Is(err, domain.ErrBookingNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "booking not found"})
		case errors.Is(err, domain.ErrBookingInvalidTransition):
			c.JSON(http.StatusConflict, gin.H{"error": "only confirmed bookings can be cancelled"})
		case errors.Is(err, domain.ErrBookingCancellationTooLate):
			c.JSON(http.StatusConflict, gin.H{"error": "booking cannot be cancelled after it starts"})
		default:
			log.Printf("cancel booking failed: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"booking": bookingResponse(*result),
		"refund_eligible": result.CancellationRefundEligible != nil && *result.CancellationRefundEligible,
	})
}

func bookingResponse(item domain.Booking) gin.H {
	response := gin.H{
		"id": item.ID,
		"hotel_id": item.HotelID,
		"luggage_count": item.LuggageCount,
		"starts_at": item.StartsAt,
		"ends_at": item.EndsAt,
		"status": item.Status,
		"created_at": item.CreatedAt,
		"updated_at": item.UpdatedAt,
	}
	if item.CancelledAt != nil {
		response["cancelled_at"] = item.CancelledAt
	}
	if item.CancellationRefundEligible != nil {
		response["refund_eligible"] = *item.CancellationRefundEligible
	}
	return response
}
