package middleware

import (
	"IAM-server/internal/utils/tokens"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func ProtectRoute() gin.HandlerFunc {
	return func(c *gin.Context) {

		// Get Authorization header
		authHeader := c.GetHeader("Authorization")

		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "authorization header missing",
			})
			c.Abort()
			return
		}

		// Expected format: Bearer <token>
		parts := strings.Split(authHeader, " ")

		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "invalid authorization header format",
			})
			c.Abort()
			return
		}

		tokenString := parts[1]

		// Verify access token
		claims, err := tokens.VerifyAccessToken(tokenString)

		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "invalid or expired access token",
			})
			c.Abort()
			return
		}

		// Add claims to Gin context
		c.Set("user", claims)

		// Continue to next handler
		c.Next()
	}
}
