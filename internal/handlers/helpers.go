package handlers

import (
	"IAM-server/internal/models"
	"IAM-server/internal/services"
	"IAM-server/internal/utils/tokens"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ResolvedUser struct {
	ID    string `json:"id"`
	Email string `json:"email"`
}

// Resolve user form the context update by middleware in protected routes.
func resolveUser(c *gin.Context) (*ResolvedUser, error) {
	value, exists := c.Get("user")
	if !exists {
		return nil, errors.New("user not found in context")
	}

	claims, ok := value.(*tokens.AccessTokenClaims)
	if !ok || claims == nil {
		return nil, errors.New("invalid user claims in context")
	}

	return &ResolvedUser{
		ID:    claims.Subject,
		Email: claims.Email,
	}, nil
}

// IssueAuthTokens generates access, refresh, and session tokens for the given customer,
// stores the token pair in Redis, creates a DB session, sets the cookies, and returns the access and refresh tokens.
func IssueAuthTokens(c *gin.Context, customer *models.Customer) (string, string, error) {
	userID := customer.ID.String()

	accessToken, accessJTI, err := tokens.GenerateAccessToken(customer.Email, userID, "")
	if err != nil {
		return "", "", err
	}

	refreshToken, refreshJTI, err := tokens.GenerateRefreshToken(userID, "")
	if err != nil {
		return "", "", err
	}

	sessionToken, err := tokens.GenerateSessionToken(customer.Name, customer.Email, userID)
	if err != nil {
		return "", "", err
	}

	// Store token pair in Redis (refresh_jti -> access_jti, access_jti -> userID)
	if err := services.StoreTokenPair(c.Request.Context(), refreshJTI, accessJTI, userID); err != nil {
		return "", "", err
	}

	// Create a new session in DB
	ip := c.ClientIP()
	deviceDetails := c.GetHeader("User-Agent")
	location := c.GetHeader("X-Location")
	_, _ = services.CreateSession(c.Request.Context(), customer.ID, refreshJTI, ip, deviceDetails, location)

	// Set cookies only if the device is not mobile
	if c.GetHeader("X-Device-Type") != "mobile" {
		// Set cookies with SameSite=None to support the cookies to set in an application from another domain
		c.SetSameSite(http.SameSiteNoneMode)
		c.SetCookie(services.REFRESH_COOKIE_NAME, refreshToken, services.COOKIE_MAX_AGE, "/", "", true, true)
		c.SetCookie(services.SESSION_COOKIE_NAME, sessionToken, services.COOKIE_MAX_AGE, "/", "", true, false)
	}

	return accessToken, refreshToken, nil
}

// LogoutHandler clears authentication cookies, deletes the session, and revokes token records from Redis.
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
