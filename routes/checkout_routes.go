package routes

import "github.com/gin-gonic/gin"

// RegisterCheckoutRoutes prepares the /checkout and /payments route groups for Week 5-6 checkout & payment simulation.
func RegisterCheckoutRoutes(group *gin.RouterGroup) {
	_ = group.Group("/checkout")
	_ = group.Group("/payments")
}
