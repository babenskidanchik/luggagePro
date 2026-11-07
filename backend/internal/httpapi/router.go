package httpapi

import (
	"github.com/babenskidanchik/luggage-platform/backend/internal/auth"
	"github.com/babenskidanchik/luggage-platform/backend/internal/booking"
	"github.com/babenskidanchik/luggage-platform/backend/internal/domain"
	"github.com/babenskidanchik/luggage-platform/backend/internal/hotel"
	"github.com/babenskidanchik/luggage-platform/backend/internal/httpapi/handlers"
	"github.com/babenskidanchik/luggage-platform/backend/internal/httpapi/middleware"
	"github.com/babenskidanchik/luggage-platform/backend/internal/repository"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"time"
)

func NewRouter(
	db *pgxpool.Pool,
	otpHandler *handlers.OTPHandler,
	phoneProtector repository.PhoneProtector,
	allowedOrigins []string,
	secureCookies bool,
	environment string,
) *gin.Engine {

	router := gin.New()

	router.Use(
		middleware.SecurityHeaders(
			environment == "production",
		),
	)

	router.Use(
		middleware.BodyLimit(1 << 20),
	)

	corsMiddleware :=
		middleware.NewCORS(
			allowedOrigins,
		)

	if err := router.SetTrustedProxies(nil); err != nil {
		panic(err)
	}

	csrfMiddleware := middleware.NewCSRF(secureCookies)

	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(
		corsMiddleware.Middleware(),
	)
	router.Use(
		csrfMiddleware.Middleware(),
	)

	userRepository :=
		repository.NewUserRepository(db, phoneProtector)

	sessionRepository :=
		repository.NewSessionRepository(db)

	authService := auth.NewService(
		sessionRepository,
		userRepository,
	)

	otpRequestLimiter :=
		middleware.NewIPRateLimiter(
			10,
			10*time.Minute,
		)

	otpVerifyLimiter :=
		middleware.NewIPRateLimiter(
			30,
			10*time.Minute,
		)

	hotelApplicationRepository :=
		repository.NewHotelApplicationRepository(db)

	hotelApplicationService :=
		hotel.NewApplicationService(
			hotelApplicationRepository,
		)
	hotelApplicationHandler :=
		handlers.NewHotelApplicationHandler(
			hotelApplicationService,
		)

	hotelRepository :=
		repository.NewHotelRepository(db)

	hotelLocationService :=
		hotel.NewLocationService(
			hotelRepository,
		)

	hotelLocationHandler :=
		handlers.NewHotelLocationHandler(
			hotelLocationService,
		)

	moderationService :=
		hotel.NewModerationService(
			hotelApplicationRepository,
		)

	moderationHandler :=
		handlers.NewModerationHandler(
			moderationService,
		)

	publicHotelRoutes := hotel.NewPublicService(
		hotelRepository,
	)

	publicHotelHandler := handlers.NewPublicHotelHandler(
		publicHotelRoutes,
	)

	bookingRepository :=
		repository.NewBookingRepository(db)

	bookingService :=
		booking.NewService(bookingRepository)

	bookingHandler :=
		handlers.NewBookingHandler(bookingService)

	requireAuth :=
		middleware.RequireAuth(authService)

	api := router.Group("/api/v1")

	bookingRoutes := api.Group(
		"/bookings",
		requireAuth,
		middleware.RequireRole(domain.RoleGuest),
	)

	bookingRoutes.POST(
		"",
		bookingHandler.Create,
	)

	hotelRoutes := api.Group(
		"/hotel",
	)

	hotelRoutes.Use(
		requireAuth,
		middleware.RequireRole(
			domain.RoleHotel,
		),
	)

	hotelRoutes.PATCH(
		"/location",
		hotelLocationHandler.Update,
	)

	api.POST(
		"/hotel/applications",
		requireAuth,
		middleware.RequireRole(
			domain.RoleGuest,
		),
		hotelApplicationHandler.Submit,
	)

	api.GET(
		"/health",
		handlers.Health(db),
	)

	api.GET(
		"/hotels",
		publicHotelHandler.ListByCity,
	)

	authRoutes := api.Group("/auth")

	authRoutes.GET(
		"/current",
		requireAuth,
		handlers.CurrentUser,
	)

	authRoutes.POST(
		"/logout",
		requireAuth,
		handlers.Logout(authService),
	)

	if otpHandler != nil {
		authRoutes.POST(
			"/otp/request",
			otpRequestLimiter.Middleware(),
			otpHandler.Request,
		)

		authRoutes.POST(
			"/otp/verify",
			otpVerifyLimiter.Middleware(),
			otpHandler.Verify,
		)
	}

	moderatorRoutes := api.Group(
		"/moderator",
	)

	moderatorRoutes.GET(
		"/hotel-applications",
		requireAuth,
		middleware.RequireRole(
			domain.RoleModerator,
		),
		moderationHandler.ListPending,
	)

	moderatorRoutes.POST(
		"/hotel-applications/:id/approve",
		requireAuth,
		middleware.RequireRole(
			domain.RoleModerator,
		),
		moderationHandler.Approve,
	)

	return router
}
