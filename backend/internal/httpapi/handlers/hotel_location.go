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

type HotelLocationService interface {
	UpdateCoordinates(
		ctx context.Context,
		userID string,
		latitude float64,
		longitude float64,
	) (*domain.Hotel, error)
}

type HotelLocationHandler struct {
	service HotelLocationService
}

func NewHotelLocationHandler(
	service HotelLocationService,
) *HotelLocationHandler {
	return &HotelLocationHandler{
		service: service,
	}
}

type updateHotelLocationRequest struct {
	Latitude  *float64 `json:"latitude"`
	Longitude *float64 `json:"longitude"`
}

func (h *HotelLocationHandler) Update(
	c *gin.Context,
) {
	value, exists := c.Get(
		middleware.CurrentUserContextKey,
	)
	if !exists {
		c.JSON(
			http.StatusUnauthorized,
			gin.H{"error": "unauthorized"},
		)
		return
	}

	user, ok := value.(*domain.User)
	if !ok {
		c.JSON(
			http.StatusUnauthorized,
			gin.H{"error": "unauthorized"},
		)
		return
	}

	var request updateHotelLocationRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(
			http.StatusBadRequest,
			gin.H{"error": "invalid request body"},
		)
		return
	}

	if request.Latitude == nil || request.Longitude == nil {
		c.JSON(
			http.StatusBadRequest,
			gin.H{"error": "latitude and longitude are required"},
		)
		return
	}

	result, err := h.service.UpdateCoordinates(
		c.Request.Context(),
		user.ID,
		*request.Latitude,
		*request.Longitude,
	)

	if err != nil {
		switch {
		case errors.Is(
			err,
			hotel.ErrInvalidHotelLocation,
		):
			c.JSON(
				http.StatusBadRequest,
				gin.H{"error": "invalid hotel location"},
			)

		default:
			c.JSON(
				http.StatusInternalServerError,
				gin.H{"error": "internal server error"},
			)
		}
		return
	}

	c.JSON(
		http.StatusOK,
		gin.H{"hotel": result},
	)
}
