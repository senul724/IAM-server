package handlers

import (
	"IAM-server/internal/services"
	"IAM-server/internal/utils/tokens"
	"net/http"

	"github.com/gin-gonic/gin"
)

// LogoutHandler clears authentication cookies, deletes the session, and revokes token records from Redis.
// @Summary Logout
// @Description Revokes the user's refresh token, clears authentication cookies, and deletes their DB session.
// @Tags Auth
// @Accept json
// @Produce json
// @Success 200 {object} map[string]string "Logout successful"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /auth/logout [post]
func LogoutHandler(c *gin.Context) {
	// Revoke refresh token, associated access token, and session from cookie if present
	if cookieToken, err := c.Cookie(services.REFRESH_COOKIE_NAME); err == nil && cookieToken != "" {
		if claims, err := tokens.VerifyRefreshToken(cookieToken); err == nil && claims.ID != "" {
			_ = services.RevokeTokenPairByRefreshJTI(c.Request.Context(), claims.ID)
			_ = services.DeleteSessionByRefreshID(c.Request.Context(), claims.ID)
		}
	}

	// Revoke token from Authorization header if present
	authHeader := c.GetHeader("Authorization")
	if authHeader != "" && len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		tokenStr := authHeader[7:]
		// Check if it's a refresh token
		if claims, err := tokens.VerifyRefreshToken(tokenStr); err == nil && claims.ID != "" {
			_ = services.RevokeTokenPairByRefreshJTI(c.Request.Context(), claims.ID)
			_ = services.DeleteSessionByRefreshID(c.Request.Context(), claims.ID)
		} else if accessClaims, err := tokens.VerifyAccessToken(tokenStr); err == nil && accessClaims.ID != "" {
			// If it's an access token, revoke access JTI
			_ = services.RevokeAccessTokenJTI(c.Request.Context(), accessClaims.ID)
		}
	}

	// Revoke token from custom X-Refresh-Token header if present (useful for mobile clients)
	if customRefresh := c.GetHeader("X-Refresh-Token"); customRefresh != "" {
		if claims, err := tokens.VerifyRefreshToken(customRefresh); err == nil && claims.ID != "" {
			_ = services.RevokeTokenPairByRefreshJTI(c.Request.Context(), claims.ID)
			_ = services.DeleteSessionByRefreshID(c.Request.Context(), claims.ID)
		}
	}

	if c.GetHeader("X-Device-Type") != "mobile" {
		// ClearAuthCookies clears the refresh and session tokens cookies.
		c.SetSameSite(http.SameSiteLaxMode)
		c.SetCookie(services.REFRESH_COOKIE_NAME, "", -1, "/", "", false, true)
		c.SetCookie(services.SESSION_COOKIE_NAME, "", -1, "/", "", false, true)
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Logged out successfully",
	})
}
