package routes

import "github.com/gin-gonic/gin"

// RegisterScanRoutes prepares the /scan route group for Week 7 QR ticket check-in validation.
func RegisterScanRoutes(group *gin.RouterGroup) {
	_ = group.Group("/scan")
}
