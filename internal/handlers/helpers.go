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
// sets the cookies, and returns the access token.
func IssueAuthTokens(c *gin.Context, customer *models.Customer) (string, string, error) {
	userID := customer.ID.String()

	accessToken, err := tokens.GenerateAccessToken(customer.Email, userID)
	if err != nil {
		return "", "", err
	}

	refreshToken, err := tokens.GenerateRefreshToken(userID)
	if err != nil {
		return "", "", err
	}

	sessionToken, err := tokens.GenerateSessionToken(customer.Name, customer.Email, userID)
	if err != nil {
		return "", "", err
	}

	// Set cookies with SameSite=None to support the cookies to set in an application from another domain
	c.SetSameSite(http.SameSiteNoneMode)
	c.SetCookie(services.REFRESH_COOKIE_NAME, refreshToken, services.COOKIE_MAX_AGE, "/", "", true, true)
	c.SetCookie(services.SESSION_COOKIE_NAME, sessionToken, services.COOKIE_MAX_AGE, "/", "", true, false)

	return accessToken, refreshToken, nil
}

// LogoutHandler clears authentication cookies.
func LogoutHandler(c *gin.Context) {
	// ClearAuthCookies clears the refresh and session tokens cookies.
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(services.REFRESH_COOKIE_NAME, "", -1, "/", "", false, true)
	c.SetCookie(services.SESSION_COOKIE_NAME, "", -1, "/", "", false, true)

	c.JSON(http.StatusOK, gin.H{
		"message": "Logged out successfully",
	})
}
