package routes

import (
	"tikeria-be/controllers"

	"github.com/gin-gonic/gin"
)

// RegisterAPIRoutes registers all domain routes under both /api and /api/v1.
func RegisterAPIRoutes(
	r *gin.Engine,
	authController *controllers.AuthController,
	eventController *controllers.EventController,
	checkoutControllers ...*controllers.CheckoutController,
) {
	var checkoutController *controllers.CheckoutController
	if len(checkoutControllers) > 0 {
		checkoutController = checkoutControllers[0]
	}

	for _, prefix := range []string{"/api", "/api/v1"} {
		group := r.Group(prefix)
		RegisterAuthRoutes(group, authController)
		RegisterEventRoutes(group, eventController)
		RegisterCheckoutRoutes(group, checkoutController)
		RegisterScanRoutes(group)
	}
}
