package middleware

import (
	"net/http"

	"tikeria-be/dto"
	"tikeria-be/pkg/frappe"

	"github.com/gin-gonic/gin"
)

const (
	ErrCodePackageLimitCategory    = "PACKAGE_LIMIT_TICKET_CATEGORY"
	ErrCodePackageLimitActiveEvent = "PACKAGE_LIMIT_ACTIVE_EVENT"
	ErrCodePackageLimitQuota       = "PACKAGE_LIMIT_QUOTA"
)

func RequirePackageLimit(limitType string, frappeClient frappe.FrappeClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		organizerPackage := c.GetString("organizerPackage")
		if organizerPackage == "" {
			organizerPackage = "basic"
		}

		// Pro package has unlimited active events, categories, and quota
		if organizerPackage == "pro" {
			c.Next()
			return
		}

		switch limitType {
		case "active_event":
			activeEvents := c.GetInt("activeEventCount")
			if activeEvents >= 1 {
				c.JSON(http.StatusForbidden, dto.WebResponse{
					Success:   false,
					Message:   "Package limit reached: Basic package allows 1 active event. Upgrade to Pro or wait until the current event ends.",
					ErrorCode: ErrCodePackageLimitActiveEvent,
				})
				c.Abort()
				return
			}
		case "quota":
			quota := c.GetInt("requestedQuota")
			if quota > 1000 {
				c.JSON(http.StatusForbidden, dto.WebResponse{
					Success:   false,
					Message:   "Package limit exceeded: Basic package allows max 1.000 tickets per event.",
					ErrorCode: ErrCodePackageLimitQuota,
				})
				c.Abort()
				return
			}
		default: // "category"
			categoryCount := c.GetInt("categoryCount")
			if categoryCount >= 1 {
				c.JSON(http.StatusForbidden, dto.WebResponse{
					Success:   false,
					Message:   "Package limit reached: Basic package allows maximum 1 ticket category. Upgrade to Pro for unlimited categories.",
					ErrorCode: ErrCodePackageLimitCategory,
				})
				c.Abort()
				return
			}
		}

		c.Next()
	}
}

func PackageLimitMiddleware(frappeClient frappe.FrappeClient) gin.HandlerFunc {
	return RequirePackageLimit("category", frappeClient)
}
