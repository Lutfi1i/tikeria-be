package routes

import (
	"tikeria-be/controllers"
	"tikeria-be/middleware"

	"github.com/gin-gonic/gin"
)

// RegisterCheckoutRoutes registers /checkout and /payments endpoints.
func RegisterCheckoutRoutes(group *gin.RouterGroup, checkoutController *controllers.CheckoutController) {
	if checkoutController == nil {
		return
	}

	checkout := group.Group("/checkout")
	{
		checkout.POST("/hold", middleware.AuthMiddleware(), checkoutController.CreateHold)
		checkout.DELETE("/hold/:id", middleware.AuthMiddleware(), checkoutController.ReleaseHold)
		checkout.POST("/participant", middleware.AuthMiddleware(), checkoutController.SaveParticipant)
		checkout.POST("/summary", checkoutController.GetSummary)
		checkout.GET("/summary/:hold_id", checkoutController.GetSummary)
		checkout.POST("/create-order", middleware.AuthMiddleware(), checkoutController.CreateOrder)
		checkout.POST("/order", middleware.AuthMiddleware(), checkoutController.CreateOrder)
		checkout.GET("/orders/:id/status", checkoutController.GetOrderStatus)
		checkout.POST("/simulate-payment", checkoutController.SimulatePaymentSuccess)
	}

	orders := group.Group("/orders")
	{
		orders.GET("/:id", checkoutController.GetOrderStatus)
		orders.GET("/:id/status", checkoutController.GetOrderStatus)
	}

	payments := group.Group("/payments")
	{
		payments.POST("/:order_id/simulate-success", checkoutController.SimulatePaymentSuccess)
		payments.POST("/notification", checkoutController.HandleNotification)
		payments.POST("/midtrans-notification", checkoutController.HandleNotification)
	}
}
