package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"tikeria-be/config"
	"tikeria-be/controllers"
	_ "tikeria-be/docs"
	"tikeria-be/repositories"
	"tikeria-be/routes"
	"tikeria-be/services"
	"tikeria-be/worker"

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

	cfg := config.LoadConfig()
	db := config.InitDatabase()

	// Start background Ticket Hold Expiry Worker (runs every 1 minute)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	worker.StartHoldExpiryWorker(ctx, db, 1*time.Minute)

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
			"app":                 "Tikeria API",
			"status":              "online",
			"frappe_sync_enabled": cfg.FrappeSyncEnabled,
			"swagger":             fmt.Sprintf("http://localhost:%s/swagger/index.html", cfg.Port),
		})
	})

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Register modular domain routes under both /api and /api/v1
	routes.RegisterAPIRoutes(r, authController, eventController)

	fmt.Printf("TikeriaGO is running at : http://localhost:%s\n", cfg.Port)
	fmt.Printf("DOCS endpoint for applying : http://localhost:%s/swagger/index.html\n\n", cfg.Port)

	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatalf("Failed to running the server: %v", err)
	}
}
