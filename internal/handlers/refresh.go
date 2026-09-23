package handlers

import (
	"IAM-server/internal/connections"
	"IAM-server/internal/models"
	"IAM-server/internal/services"
	"IAM-server/internal/utils/tokens"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// RefreshTokenHandler handles refreshing the access token using the refresh_token cookie or Authorization header.
// @Summary Refresh access token
// @Description Refresh access token using the refresh_token cookie or Authorization Bearer header
// @Tags Auth
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /auth/refresh [post]
func RefreshTokenHandler(c *gin.Context) {
	var refreshTokenString string

	cookieToken, err := c.Cookie(services.REFRESH_COOKIE_NAME)
	if err == nil && cookieToken != "" {
		refreshTokenString = cookieToken
	} else {
		// Fallback to Authorization header if cookie is absent
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" && len(authHeader) > 7 && authHeader[:7] == "Bearer " {
			refreshTokenString = authHeader[7:]
		}
	}

	if refreshTokenString == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "refresh token missing"})
		return
	}

	claims, err := tokens.VerifyRefreshToken(refreshTokenString)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired refresh token"})
		return
	}

	userID := claims.Subject
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user ID in token"})
		return
	}

	var customer models.Customer
	if err := connections.DB.First(&customer, "id = ?", userUUID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}

	newAccessToken, err := tokens.GenerateAccessToken(customer.Email, customer.ID.String())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate access token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token": newAccessToken,
	})
}
