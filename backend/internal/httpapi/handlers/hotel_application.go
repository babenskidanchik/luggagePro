package handlers

import (
	"context"
	"errors"
	"log"
	"net/http"

	"github.com/babenskidanchik/luggage-platform/backend/internal/domain"
	"github.com/babenskidanchik/luggage-platform/backend/internal/hotel"
	"github.com/babenskidanchik/luggage-platform/backend/internal/httpapi/middleware"
	"github.com/gin-gonic/gin"
)

type HotelApplicationService interface {
	Submit(
		ctx context.Context,
		userID string,
		input hotel.ApplicationInput,
	) (*domain.HotelApplication, error)
}

type HotelApplicationHandler struct {
	service HotelApplicationService
}

func NewHotelApplicationHandler(
	service HotelApplicationService,
) *HotelApplicationHandler {
	return &HotelApplicationHandler{
		service: service,
	}
}

type submitHotelApplicationRequest struct {
	VenueType   domain.VenueType `json:"venue_type"`
	LegalName   string           `json:"legal_name"`
	DisplayName string           `json:"display_name"`
	TaxID       string           `json:"tax_id"`
	CountryCode string           `json:"country_code"`
	City        string           `json:"city"`
	Address     string           `json:"address"`
}

func (h *HotelApplicationHandler) Submit(
	c *gin.Context,
) {
	value, exists := c.Get(
		middleware.CurrentUserContextKey,
	)

	if !exists {
		c.AbortWithStatusJSON(
			http.StatusUnauthorized,
			gin.H{
				"error": "authentication required",
			},
		)
		return
	}

	user, ok := value.(*domain.User)
	if !ok || user == nil {
		c.AbortWithStatusJSON(
			http.StatusUnauthorized,
			gin.H{
				"error": "authentication required",
			},
		)
		return
	}

	var request submitHotelApplicationRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		c.AbortWithStatusJSON(
			http.StatusBadRequest,
			gin.H{
				"error": "invalid request",
			},
		)
		return
	}

	application, err := h.service.Submit(
		c.Request.Context(),
		user.ID,
		hotel.ApplicationInput{
			VenueType:   request.VenueType,
			LegalName:   request.LegalName,
			DisplayName: request.DisplayName,
			TaxID:       request.TaxID,
			CountryCode: request.CountryCode,
			City:        request.City,
			Address:     request.Address,
		},
	)

	if err != nil {
		switch {
		case errors.Is(
			err,
			hotel.ErrInvalidApplication,
		):
			c.AbortWithStatusJSON(
				http.StatusBadRequest,
				gin.H{
					"error": "invalid application",
				},
			)

		case errors.Is(
			err,
			domain.ErrHotelApplicationAlreadyExists,
		),
			errors.Is(
				err,
				domain.ErrHotelTaxIDAlreadyInUse,
			):
			c.AbortWithStatusJSON(
				http.StatusConflict,
				gin.H{
					"error": "application conflict",
				},
			)

		default:
			log.Printf(
				"submit hotel application failed: %v",
				err,
			)

			c.AbortWithStatusJSON(
				http.StatusInternalServerError,
				gin.H{
					"error": "internal server error",
				},
			)
		}

		return
	}

	c.JSON(
		http.StatusCreated,
		gin.H{
			"venue_type":   application.VenueType,
			"id":           application.ID,
			"legal_name":   application.LegalName,
			"display_name": application.DisplayName,
			"tax_id":       application.TaxID,
			"country_code": application.CountryCode,
			"city":         application.City,
			"address":      application.Address,
			"status":       application.Status,
			"created_at":   application.CreatedAt,
		},
	)
}
