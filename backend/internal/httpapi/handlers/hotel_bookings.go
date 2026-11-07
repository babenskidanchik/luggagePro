package handlers

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/babenskidanchik/luggage-platform/backend/internal/domain"
	"github.com/babenskidanchik/luggage-platform/backend/internal/httpapi/middleware"
	"github.com/gin-gonic/gin"
)

type HotelBookingsService interface {
	ConfirmBaggageReceived(
		ctx context.Context,
		hotelUserID string,
		bookingID string,
	) (*domain.Booking, error)
	Complete(
		ctx context.Context,
		hotelUserID string,
		bookingID string,
	) (*domain.Booking, error)
	ListForHotel(
		ctx context.Context,
		hotelUserID string,
		limit int,
		offset int,
	) ([]domain.Booking, error)
}

type HotelBookingsHandler struct {
	service HotelBookingsService
}

func NewHotelBookingsHandler(service HotelBookingsService) *HotelBookingsHandler {
	return &HotelBookingsHandler{service: service}
}

func (h *HotelBookingsHandler) List(c *gin.Context) {
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

	items, err := h.service.ListForHotel(
		c.Request.Context(),
		user.ID,
		limit,
		offset,
	)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidBooking) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid booking list parameters"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	bookings := make([]gin.H, 0, len(items))
	for _, item := range items {
		bookings = append(bookings, gin.H{
			"id": item.ID,
			"hotel_id": item.HotelID,
			"luggage_count": item.LuggageCount,
			"starts_at": item.StartsAt,
			"ends_at": item.EndsAt,
			"status": item.Status,
			"created_at": item.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"bookings": bookings,
		"limit": limit,
		"offset": offset,
	})
}

func (h *HotelBookingsHandler) Receive(c *gin.Context) {
	user, ok := middleware.CurrentUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	result, err := h.service.ConfirmBaggageReceived(
		c.Request.Context(),
		user.ID,
		c.Param("id"),
	)
	h.respondToTransition(c, result, err)
}

func (h *HotelBookingsHandler) Complete(c *gin.Context) {
	user, ok := middleware.CurrentUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	result, err := h.service.Complete(
		c.Request.Context(),
		user.ID,
		c.Param("id"),
	)
	h.respondToTransition(c, result, err)
}

func (h *HotelBookingsHandler) respondToTransition(
	c *gin.Context,
	result *domain.Booking,
	err error,
) {
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrInvalidBooking):
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid booking"})
		case errors.Is(err, domain.ErrBookingNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "booking not found"})
		case errors.Is(err, domain.ErrBookingInvalidTransition):
			c.JSON(http.StatusConflict, gin.H{"error": "booking cannot move to requested status"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"booking": gin.H{
			"id": result.ID,
			"status": result.Status,
			"updated_at": result.UpdatedAt,
		},
	})
}
