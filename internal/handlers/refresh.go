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

// RefreshTokenHandler handles refreshing the access and refresh tokens using the refresh_token cookie or Authorization header.
// @Summary Refresh access and refresh tokens
// @Description Refresh and rotate tokens using the refresh_token cookie or Authorization Bearer header
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
		} else if customHeader := c.GetHeader("X-Refresh-Token"); customHeader != "" {
			refreshTokenString = customHeader
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

	if claims.ID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid refresh token: missing token identifier"})
		return
	}

	// Cross-check refresh token JTI with Redis
	_, err = services.ValidateRefreshTokenJTI(c.Request.Context(), claims.ID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "refresh token revoked or expired"})
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

	oldRefreshID := claims.ID

	// Invalidate/remove old token records from Redis
	_ = services.RevokeTokenPairByRefreshJTI(c.Request.Context(), oldRefreshID)

	// Generate new access and refresh tokens
	userIDString := customer.ID.String()
	newAccessToken, newAccessJTI, err := tokens.GenerateAccessToken(customer.Email, userIDString, "")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate access token"})
		return
	}

	newRefreshToken, newRefreshJTI, err := tokens.GenerateRefreshToken(userIDString, "")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate refresh token"})
		return
	}

	sessionToken, err := tokens.GenerateSessionToken(customer.Name, customer.Email, userIDString)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate session token"})
		return
	}

	// Store new token pair in Redis
	if err := services.StoreTokenPair(c.Request.Context(), newRefreshJTI, newAccessJTI, userIDString); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to store token pair in redis"})
		return
	}

	_ = services.UpdateSessionRefreshID(c.Request.Context(), oldRefreshID, newRefreshJTI)

	// Set cookies only if the device is not mobile
	if c.GetHeader("X-Device-Type") != "mobile" {
		c.SetSameSite(http.SameSiteNoneMode)
		c.SetCookie(services.REFRESH_COOKIE_NAME, newRefreshToken, services.COOKIE_MAX_AGE, "/", "", true, true)
		c.SetCookie(services.SESSION_COOKIE_NAME, sessionToken, services.COOKIE_MAX_AGE, "/", "", true, false)
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token":  newAccessToken,
		"refresh_token": newRefreshToken,
	})
}
