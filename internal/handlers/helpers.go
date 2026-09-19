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

// SetAuthCookies sets the refresh and session tokens as HTTP-only cookies.
func SetAuthCookies(c *gin.Context, refreshToken, sessionToken string) {
	// Set cookies with SameSite=None to support the cookies to set in an application from another domain
	c.SetSameSite(http.SameSiteNoneMode)
	c.SetCookie(services.REFRESH_COOKIE_NAME, refreshToken, services.COOKIE_MAX_AGE, "/", "", true, true)
	c.SetCookie(services.SESSION_COOKIE_NAME, sessionToken, services.COOKIE_MAX_AGE, "/", "", true, false)
}

// ClearAuthCookies clears the refresh and session tokens cookies.
func ClearAuthCookies(c *gin.Context) {
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(services.REFRESH_COOKIE_NAME, "", -1, "/", "", false, true)
	c.SetCookie(services.SESSION_COOKIE_NAME, "", -1, "/", "", false, true)
}

// GetAuthenticatedUserID extracts the authenticated user ID (UUID string) from the Gin context.
func GetAuthenticatedUserID(c *gin.Context) (string, error) {
	userVal, exists := c.Get("user")
	if !exists {
		return "", errors.New("user not found in context")
	}

	claims, ok := userVal.(*tokens.AccessTokenClaims)
	if !ok || claims.Subject == "" {
		return "", errors.New("invalid user claims in context")
	}

	return claims.Subject, nil
}

// IssueAuthTokens generates access, refresh, and session tokens for the given customer,
// sets the cookies, and returns the access token.
func IssueAuthTokens(c *gin.Context, customer *models.Customer) (string, error) {
	userID := customer.ID.String()

	accessToken, err := tokens.GenerateAccessToken(customer.Email, userID)
	if err != nil {
		return "", err
	}

	refreshToken, err := tokens.GenerateRefreshToken(userID)
	if err != nil {
		return "", err
	}

	sessionToken, err := tokens.GenerateSessionToken(customer.Name, customer.Email, userID)
	if err != nil {
		return "", err
	}

	SetAuthCookies(c, refreshToken, sessionToken)
	return accessToken, nil
}

// RefreshTokenHandler handles refreshing the access token using the refresh_token cookie or Authorization header.
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

// LogoutHandler clears authentication cookies.
func LogoutHandler(c *gin.Context) {
	ClearAuthCookies(c)
	c.JSON(http.StatusOK, gin.H{
		"message": "Logged out successfully",
	})
}
