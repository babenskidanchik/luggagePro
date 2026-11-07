package handlers

import (
	"context"
	"errors"
	"net/http"

	"github.com/babenskidanchik/luggage-platform/backend/internal/domain"
	"github.com/babenskidanchik/luggage-platform/backend/internal/hotel"
	"github.com/babenskidanchik/luggage-platform/backend/internal/httpapi/middleware"
	"github.com/gin-gonic/gin"
)

type HotelBookingSettingsService interface {
	Update(
		ctx context.Context,
		userID string,
		luggageCapacity int,
		acceptingBookings bool,
	) (*domain.Hotel, error)
}

type HotelBookingSettingsHandler struct {
	service HotelBookingSettingsService
}

func NewHotelBookingSettingsHandler(service HotelBookingSettingsService) *HotelBookingSettingsHandler {
	return &HotelBookingSettingsHandler{service: service}
}

type updateHotelBookingSettingsRequest struct {
	LuggageCapacity  *int  `json:"luggage_capacity"`
	AcceptingBookings *bool `json:"accepting_bookings"`
}

func (h *HotelBookingSettingsHandler) Update(c *gin.Context) {
	user, ok := middleware.CurrentUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	var request updateHotelBookingSettingsRequest
	if err := c.ShouldBindJSON(&request); err != nil ||
		request.LuggageCapacity == nil ||
		request.AcceptingBookings == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "luggage_capacity and accepting_bookings are required"})
		return
	}

	result, err := h.service.Update(
		c.Request.Context(),
		user.ID,
		*request.LuggageCapacity,
		*request.AcceptingBookings,
	)
	if err != nil {
		switch {
		case errors.Is(err, hotel.ErrInvalidBookingSettings):
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid booking settings"})
		case errors.Is(err, domain.ErrHotelNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "hotel not found"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"booking_settings": gin.H{
			"luggage_capacity": result.LuggageCapacity,
			"accepting_bookings": result.AcceptingBookings,
		},
	})
}
