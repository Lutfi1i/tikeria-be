package routes

import (
	"tikeria-be/controllers"
	"tikeria-be/middleware"

	"github.com/gin-gonic/gin"
)

func RegisterEventRoutes(group *gin.RouterGroup, eventController *controllers.EventController) {
	organizer := group.Group("/organizer")
	{
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
