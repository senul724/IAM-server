package handlers

import (
	"IAM-server/internal/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

type CheckEmailRequest struct {
	Email string `json:"email" binding:"required,email"`
}

type SendLoginOTPRequest struct {
	Email string `json:"email" binding:"required,email"`
}

type LoginWithOTPRequest struct {
	Email string `json:"email" binding:"required,email"`
	OTP   string `json:"otp" binding:"required,len=6"`
}

type RegisterWithOTPRequest struct {
	Name  string `json:"name" binding:"required"`
	Email string `json:"email" binding:"required,email"`
	OTP   string `json:"otp" binding:"required,len=6"`
}

// CheckEmailHandler handles verifying whether an account exists and has a password.
// @Summary Check email existence and password availability
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body CheckEmailRequest true "Email to check"
// @Success 200 {object} services.CheckEmailResult
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /auth/check-email [post]
func CheckEmailHandler(c *gin.Context) {
	var req CheckEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userAvailable, passwordAvailable, err := services.CheckEmail(req.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check email"})
		return
	}

	c.JSON(http.StatusOK, services.CheckEmailResult{
		UserAvailable:     userAvailable,
		PasswordAvailable: passwordAvailable,
	})
}

// SendLoginOTPHandler generates and sends an OTP to the user's email.
// @Summary Send login OTP via email
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body SendLoginOTPRequest true "Email address"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /auth/send-login-otp [post]
func SendLoginOTPHandler(c *gin.Context) {
	var req SendLoginOTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := services.SendLoginOTP(c.Request.Context(), req.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Login OTP sent successfully",
	})
}

// LoginWithOTPHandler verifies the OTP, authenticates the customer, and sets session/refresh cookies.
// @Summary Login with email and OTP
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body LoginWithOTPRequest true "Login with OTP request"
// @Success 200 {object} map[string]any
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /auth/login-otp [post]
func LoginWithOTPHandler(c *gin.Context) {
	var req LoginWithOTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	customer, err := services.LoginWithOTP(c.Request.Context(), req.Email, req.OTP)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	accessToken, refreshToken, err := IssueAuthTokens(c, customer)
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

// RegisterWithOTPHandler verifies OTP and registers a new customer.
// @Summary Register with name, email and OTP
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body RegisterWithOTPRequest true "Register with OTP request"
// @Success 201 {object} map[string]any
// @Failure 400 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /auth/register-otp [post]
func RegisterWithOTPHandler(c *gin.Context) {
	var req RegisterWithOTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	customer, err := services.RegisterWithOTP(c.Request.Context(), req.Name, req.Email, req.OTP)
	if err != nil {
		if err.Error() == "user already registered with this email" {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	accessToken, refreshToken, err := IssueAuthTokens(c, customer)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate tokens"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":       "Registration successful",
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"user":          customer,
	})
}
