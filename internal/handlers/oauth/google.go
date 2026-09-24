package oauth

import (
	"IAM-server/internal/handlers"
	oauthService "IAM-server/internal/services/oauth"
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

// GoogleClaims represents the user info claims from a Google ID token.
type GoogleClaims struct {
	Sub           string `json:"sub"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	Name          string `json:"name"`
	Picture       string `json:"picture"`
}

// @Summary Initiate Google OAuth login
// @Description Redirects the user to Google's OAuth consent screen. Sets an oauth_state cookie for CSRF protection.
// @Tags OAuth
// @Produce html
// @Success 302 "Redirect to Google consent screen"
// @Failure 500 {object} map[string]string
// @Router /auth/google/login [get]
func GoogleLoginHandler(c *gin.Context) {
	// If the user already has a valid refresh token cookie, redirect to the frontend directly
	if cookie, err := c.Cookie("refresh_token"); err == nil && cookie != "" {
		frontendURL := os.Getenv("FRONTEND_URL")
		if frontendURL == "" {
			frontendURL = "/"
		}
		c.Redirect(http.StatusFound, frontendURL)
		return
	}

	state := randomString(32)

	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie("oauth_state", state, 600, "/", "", false, true)

	svc, err := oauthService.GoogleOauthService(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to initialize Google OAuth service"})
		return
	}

	c.Redirect(http.StatusFound, svc.GetAuthCodeURL(state))
}

// @Summary Handle Google OAuth callback
// @Description Handles the OAuth callback from Google. Validates the state parameter, exchanges the authorization code for tokens, verifies the ID token, finds or creates the user, and issues auth tokens.
// @Tags OAuth
// @Produce json
// @Param state query string true "OAuth state parameter for CSRF validation"
// @Param code query string true "Authorization code from Google"
// @Success 200 {object} map[string]any "Login successful with tokens and user info"
// @Failure 400 {object} map[string]string "Invalid state"
// @Failure 401 {object} map[string]string "Invalid ID token"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /auth/google/callback [get]
func GoogleCallbackHandler(c *gin.Context) {
	// Validate CSRF state
	stateCookie, err := c.Cookie("oauth_state")
	if err != nil || stateCookie != c.Query("state") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid state"})
		return
	}

	// Clear the oauth_state cookie
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie("oauth_state", "", -1, "/", "", false, true)

	// Initialize Google OAuth service
	svc, err := oauthService.GoogleOauthService(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to initialize Google OAuth service"})
		return
	}

	// Exchange authorization code for token
	token, err := svc.Exchange(c.Request.Context(), c.Query("code"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "code exchange failed"})
		return
	}

	// Extract and verify ID token
	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "no id_token in response"})
		return
	}

	idToken, err := svc.Verify(c.Request.Context(), rawIDToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid id_token"})
		return
	}

	// Parse claims
	var claims GoogleClaims
	if err := idToken.Claims(&claims); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to parse claims"})
		return
	}

	// Find or create the user
	customer, err := oauthService.FindOrCreateByGoogle(c.Request.Context(), claims.Email, claims.Name, claims.Picture)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to find or create user"})
		return
	}

	// Issue auth tokens (access, refresh, session) and set cookies
	accessToken, refreshToken, err := handlers.IssueAuthTokens(c, customer)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate tokens"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":       "Login successful",
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"user":          customer,
	})
}

func randomString(n int) string {
	b := make([]byte, n)
	rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}
