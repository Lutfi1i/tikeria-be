package main

import (
	"fmt"
	"log"
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

	authService := services.NewAuthService(userRepo)
	eventService := services.NewEventService(eventRepo)

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
			c.AbortWithStatus(24)
			return
		}
		c.Next()
	})

	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"app":     "Tikeria API",
			"status":  "online",
			"swagger": "http://localhost:8080/swagger/index.html",
		})
	})

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	api := r.Group("/api/v1")
	{
		auth := api.Group("/auth")
		{
			auth.POST("/register", authController.Register)
			auth.POST("/login", authController.Login)
		}

		events := api.Group("/events")
		{
			events.GET("", eventController.GetAllEvents)
			events.GET("/:id", eventController.GetEventByID)
			events.POST("", middleware.AuthMiddleware("partner", "admin"), eventController.CreateEvent)
		}
	}

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
