package middleware

import (
	"net/http"
	"strings"

	"tikeria-be/dto"
	"tikeria-be/utils"

	"github.com/gin-gonic/gin"
)

func AuthMiddleware(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, dto.WebResponse{
				Success: false,
				Message: "Authorization header not found",
			})
			c.Abort()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, dto.WebResponse{
				Success: false,
				Message: "Invalid Authorization token format (Use: Bearer <token>)",
			})
			c.Abort()
			return
		}

		claims, err := utils.ValidateToken(parts[1])
		if err != nil {
			c.JSON(http.StatusUnauthorized, dto.WebResponse{
				Success: false,
				Message: err.Error(),
			})
			c.Abort()
			return
		}

		if len(roles) > 0 {
			roleAllowed := false
			for _, role := range roles {
				if claims.Role == role {
					roleAllowed = true
					break
				}
			}

			if !roleAllowed {
				c.JSON(http.StatusForbidden, dto.WebResponse{
					Success: false,
					Message: "You do not have access to this resource (Role not permitted)",
				})
				c.Abort()
				return
			}
		}

		c.Set("userID", claims.UserID)
		c.Set("email", claims.Email)
		c.Set("role", claims.Role)

		c.Next()
	}
}
