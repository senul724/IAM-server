package handlers

import (
	"IAM-server/internal/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

type LoginWithPasswordRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type SendResetPasswordOTPRequest struct {
	Email string `json:"email" binding:"required,email"`
}

type ResetPasswordRequest struct {
	Email       string `json:"email" binding:"required,email"`
	OTP         string `json:"otp" binding:"required,len=6"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
}

// LoginWithPasswordHandler authenticates a customer using email and password.
// @Summary Login with email and password
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body LoginWithPasswordRequest true "Login credentials"
// @Success 200 {object} map[string]any
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /auth/login-password [post]
func LoginWithPasswordHandler(c *gin.Context) {
	var req LoginWithPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	customer, err := services.LoginWithPassword(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid email or password"})
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

// SendResetPasswordOTPHandler sends a password reset OTP to the user's email.
// @Summary Send password reset OTP
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body SendResetPasswordOTPRequest true "Email address"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /auth/send-reset-otp [post]
func SendResetPasswordOTPHandler(c *gin.Context) {
	var req SendResetPasswordOTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := services.SendResetPasswordOTP(c.Request.Context(), req.Email)
	if err != nil {
		if err.Error() == "user does not exist" {
			c.JSON(http.StatusNotFound, gin.H{"error": "user does not exist"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Password reset OTP sent successfully",
	})
}

// ResetPasswordHandler verifies OTP and sets a new password for the customer.
// @Summary Reset password with OTP
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body ResetPasswordRequest true "Reset password request"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /auth/reset-password [post]
func ResetPasswordHandler(c *gin.Context) {
	var req ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := services.ResetPassword(c.Request.Context(), req.Email, req.OTP, req.NewPassword)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Password reset successfully",
	})
}
