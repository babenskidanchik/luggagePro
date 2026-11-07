package handlers

import (
	"context"
	"net/http"
	"strconv"

	"github.com/babenskidanchik/luggage-platform/backend/internal/domain"
	"github.com/babenskidanchik/luggage-platform/backend/internal/httpapi/middleware"
	"github.com/gin-gonic/gin"
)

type HotelBookingsService interface {
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

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	items, err := h.service.ListForHotel(
		c.Request.Context(),
		user.ID,
		limit,
		offset,
	)
	if err != nil {
		if err == domain.ErrInvalidBooking {
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
