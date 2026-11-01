package handlers

import (
	"context"
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/babenskidanchik/luggage-platform/backend/internal/domain"
	"github.com/babenskidanchik/luggage-platform/backend/internal/hotel"
	"github.com/gin-gonic/gin"
)

type PublicHotelsService interface {
	ListByCity(
		ctx context.Context,
		city string,
		limit int,
		offset int,
	) ([]domain.Hotel, error)
}

type PublicHotelHandler struct {
	service PublicHotelsService
}

func NewPublicHotelHandler(
	service PublicHotelsService,
) *PublicHotelHandler {
	return &PublicHotelHandler{
		service: service,
	}
}

type publicHotelResponse struct {
	ID          string   `json:"id"`
	DisplayName string   `json:"display_name"`
	CountryCode string   `json:"country_code"`
	City        string   `json:"city"`
	Address     string   `json:"address"`
	Latitude    *float64 `json:"latitude,omitempty"`
	Longitude   *float64 `json:"longitude,omitempty"`
}

func (h *PublicHotelHandler) ListByCity(
	c *gin.Context,
) {
	city := c.Query("city")

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

	hotels, err := h.service.ListByCity(
		c.Request.Context(),
		city,

		limit,
		offset,
	)
	if err != nil {
		if errors.Is(
			err,
			hotel.ErrInvalidPublicHotelQuery,
		) {
			c.JSON(
				http.StatusBadRequest,
				gin.H{"error": "invalid query parameters"},
			)
			return
		}

		log.Printf(
			"list public hotels by city failed: %v",
			err,
		)

		c.JSON(
			http.StatusInternalServerError,
			gin.H{"error": "internal server error"},
		)
		return
	}

	response := make(
		[]publicHotelResponse,
		0,
		len(hotels),
	)

	for _, item := range hotels {
		if item.Latitude == nil || item.Longitude == nil {
			continue
		}

		response = append(
			response,
			publicHotelResponse{
				ID:          item.ID,
				DisplayName: item.DisplayName,
				CountryCode: item.CountryCode,
				City:        item.City,
				Address:     item.Address,
				Latitude:    item.Latitude,
				Longitude:   item.Longitude,
			},
		)
	}

	c.JSON(
		http.StatusOK,
		gin.H{
			"hotels": response,
			"count":  len(response),
		},
	)
}
