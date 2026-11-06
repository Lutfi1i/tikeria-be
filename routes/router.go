package routes

import (
	"tikeria-be/controllers"

	"github.com/gin-gonic/gin"
)

// RegisterAPIRoutes registers all domain routes under both /api and /api/v1.
func RegisterAPIRoutes(r *gin.Engine, authController *controllers.AuthController, eventController *controllers.EventController) {
	for _, prefix := range []string{"/api", "/api/v1"} {
		group := r.Group(prefix)
		RegisterAuthRoutes(group, authController)
		RegisterEventRoutes(group, eventController)
		RegisterCheckoutRoutes(group)
		RegisterScanRoutes(group)
	}
}
