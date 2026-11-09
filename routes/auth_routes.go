package routes

import (
	"tikeria-be/controllers"
	"tikeria-be/middleware"

	"github.com/gin-gonic/gin"
)

func RegisterAuthRoutes(group *gin.RouterGroup, authController *controllers.AuthController) {
	auth := group.Group("/auth")
	{
		auth.POST("/register", authController.Register)
		auth.POST("/verify-otp", authController.VerifyOTP)
		auth.POST("/resend-otp", authController.ResendOTP)
		auth.POST("/login", authController.Login)
		auth.GET("/me", middleware.AuthMiddleware(), authController.GetMe)
	}

	organizer := group.Group("/organizer")
	{
		organizer.POST("/register", authController.RegisterOrganizer)
		organizer.POST("/verify-email", authController.VerifyEmail)
		organizer.GET("/me", middleware.AuthMiddleware("organizer", "partner", "admin"), authController.GetMe)
		organizer.GET("/package", middleware.AuthMiddleware("organizer", "partner", "admin"), authController.GetOrganizerPackage)
	}
}
