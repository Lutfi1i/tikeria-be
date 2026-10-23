package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"tikeria-be/config"
	"tikeria-be/controllers"
	_ "tikeria-be/docs"
	"tikeria-be/middleware"
	"tikeria-be/repositories"
	"tikeria-be/services"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func registerAPIRoutes(group *gin.RouterGroup, authController *controllers.AuthController, eventController *controllers.EventController) {
	auth := group.Group("/auth")
	{
		auth.POST("/register", authController.Register)
		auth.POST("/login", authController.Login)
		auth.GET("/me", middleware.AuthMiddleware(), authController.GetMe)
	}

	organizer := group.Group("/organizer")
	{
		organizer.POST("/register", authController.RegisterOrganizer)
		organizer.POST("/verify-email", authController.VerifyEmail)
		organizer.GET("/me", middleware.AuthMiddleware("organizer", "partner", "admin"), authController.GetMe)
		organizer.GET("/package", middleware.AuthMiddleware("organizer", "partner", "admin"), authController.GetOrganizerPackage)
		organizer.GET("/events", middleware.AuthMiddleware("organizer", "partner", "admin"), eventController.GetOrganizerEvents)
		organizer.POST("/events", middleware.AuthMiddleware("organizer", "partner", "admin"), eventController.CreateEvent)
		organizer.PUT("/events/:id", middleware.AuthMiddleware("organizer", "partner", "admin"), eventController.UpdateEvent)
		organizer.POST("/events/:id/categories", middleware.AuthMiddleware("organizer", "partner", "admin"), eventController.AddTicketCategory)
		organizer.PUT("/events/:id/categories/:category_id", middleware.AuthMiddleware("organizer", "partner", "admin"), eventController.UpdateTicketCategory)
		organizer.DELETE("/events/:id/categories/:category_id", middleware.AuthMiddleware("organizer", "partner", "admin"), eventController.DeleteTicketCategory)
		organizer.POST("/events/:id/submit-approval", middleware.AuthMiddleware("organizer", "partner", "admin"), eventController.SubmitEventApproval)
	}

	admin := group.Group("/admin")
	{
		admin.POST("/events/:id/approve", middleware.AuthMiddleware("admin"), eventController.ApproveEvent)
		admin.POST("/events/:id/reject", middleware.AuthMiddleware("admin"), eventController.RejectEvent)
	}

	events := group.Group("/events")
	{
		events.GET("", eventController.GetAllEvents)
		events.GET("/:id", eventController.GetEventByID)
		events.GET("/:id/categories", eventController.GetEventCategories)
		events.POST("", middleware.AuthMiddleware("organizer", "partner", "admin"), eventController.CreateEvent)
	}
}

// @title           Tikeria Ticketing Platform API
// @version         1.0
// @description     Backend API RESTful for ticket platform events
// @termsOfService  http://swagger.io/terms/

// @license.name  MIT
// @license.url   https://opensource.org/licenses/MIT

// @host      localhost:8080
// @BasePath  /

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Insert your JWT Token: Bearer <jwt_token>
func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("env file is not found")
	}

	db := config.InitDatabase()

	userRepo := repositories.NewUserRepository(db)
	eventRepo := repositories.NewEventRepository(db)

	authService := services.NewAuthService(userRepo, eventRepo)
	eventService := services.NewEventService(eventRepo, userRepo)

	authController := controllers.NewAuthController(authService)
	eventController := controllers.NewEventController(eventService)

	r := gin.Default()

	// Setup CORS Middleware agar Frontend React / Flutter bisa akses Backend
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	})

	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"app":     "Tikeria API",
			"status":  "online",
			"swagger": "http://localhost:8080/swagger/index.html",
		})
	})

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Register routes under both /api and /api/v1 for full compatibility
	registerAPIRoutes(r.Group("/api"), authController, eventController)
	registerAPIRoutes(r.Group("/api/v1"), authController, eventController)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("TikeriaGO is running at : http://localhost:%s\n", port)
	fmt.Printf("DOCS endpoint for applying : http://localhost:%s/swagger/index.html\n\n", port)

	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to running the server: %v", err)
	}
}
