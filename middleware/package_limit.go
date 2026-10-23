package middleware

import (
	"net/http"

	"tikeria-be/dto"
	"tikeria-be/pkg/frappe"

	"github.com/gin-gonic/gin"
)

func PackageLimitMiddleware(frappeClient frappe.FrappeClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		organizerPackage := c.GetString("organizerPackage")
		if organizerPackage == "" {
			organizerPackage = "basic"
		}

		// If organizer is on Pro package, unlimited
		if organizerPackage == "pro" {
			c.Next()
			return
		}

		// For Basic package, check limit if adding multiple ticket categories
		categoryCount := c.GetInt("categoryCount")
		if categoryCount >= 1 {
			c.JSON(http.StatusForbidden, dto.WebResponse{
				Success: false,
				Message: "Package limit reached: Basic package allows maximum 1 ticket category. Upgrade to Pro for unlimited categories.",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
