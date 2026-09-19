package services

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"IAM-server/internal/connections"
	"IAM-server/internal/models"

	"github.com/resend/resend-go/v2"
	"gorm.io/gorm"
)

const (
	OTP_TTL             = 10 * time.Minute
	REFRESH_COOKIE_NAME = "refresh_token"
	SESSION_COOKIE_NAME = "session_token"
	COOKIE_MAX_AGE      = 15 * 24 * 60 * 60 // 15 days in seconds
)

type CheckEmailResult struct {
	UserAvailable     bool `json:"user_available"`
	PasswordAvailable bool `json:"password_available"`
}

// CheckEmail checks if a customer exists under the given email,
// and whether a password has been set for that account.
func CheckEmail(email string) (userAvailable bool, passwordAvailable bool, err error) {
	var customer models.Customer
	err = connections.DB.Where("email = ?", email).First(&customer).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, false, nil
		}
		return false, false, err
	}

	hasPassword := customer.Password != nil && *customer.Password != ""
	return true, hasPassword, nil
}

// SendLoginOTP generates a 6-digit OTP, stores it in Redis under login_{email},
// and sends a confirmation email with the OTP using Resend.
func SendLoginOTP(ctx context.Context, email string) error {
	if connections.Redis == nil {
		return errors.New("redis connection is not initialized")
	}

	otp, err := GenerateOTP(ctx, email, OTPReasonLogin)
	if err != nil {
		return fmt.Errorf("failed to generate OTP: %w", err)
	}

	err = sendOTPEmail(email, otp)
	if err != nil {
		return fmt.Errorf("failed to send OTP email: %w", err)
	}

	return nil
}

// LoginWithOTP validates the OTP from Redis, ensures the user exists,
// generates refresh and session tokens, and sends them as HTTP-only cookies.
func LoginWithOTP(ctx context.Context, email string, otp string) (*models.Customer, error) {
	if connections.Redis == nil {
		return nil, errors.New("redis connection is not initialized")
	}

	err := ValidateOTP(ctx, email, otp, OTPReasonLogin)
	if err != nil {
		return nil, err
	}

	var customer models.Customer
	err = connections.DB.Where("email = ?", email).First(&customer).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user does not exist")
		}
		return nil, err
	}
	return &customer, nil
}

// RegisterWithOTP validates the OTP from Redis, ensures the user does not exist and
// creates the customer in the database.
func RegisterWithOTP(ctx context.Context, name string, email string, otp string) (*models.Customer, error) {
	if connections.Redis == nil {
		return nil, errors.New("redis connection is not initialized")
	}

	err := ValidateOTP(ctx, email, otp, OTPReasonLogin)
	if err != nil {
		return nil, err
	}

	var existing models.Customer
	err = connections.DB.Where("email = ?", email).First(&existing).Error
	if err == nil {
		return nil, errors.New("user already registered with this email")
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	customer := models.Customer{
		Name:  name,
		Email: email,
	}

	if err := connections.DB.Create(&customer).Error; err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return &customer, nil
}

// sendOTPEmail sends an email containing the 6-digit OTP using Resend.
func sendOTPEmail(toEmail, otp string) error {
	apiKey := os.Getenv("RESEND_API_KEY")
	if apiKey == "" {
		return errors.New("RESEND_API_KEY environment variable is not set")
	}

	fromEmail := os.Getenv("RESEND_FROM_EMAIL")
	if fromEmail == "" {
		fromEmail = "onboarding@resend.dev"
	}

	client := resend.NewClient(apiKey)

	params := &resend.SendEmailRequest{
		From:    fromEmail,
		To:      []string{toEmail},
		Subject: "Your Login Verification Code",
		Html: fmt.Sprintf(`
			<div style="font-family: sans-serif; padding: 20px; line-height: 1.5; color: #111;">
				<h2>Verification Code</h2>
				<p>Your one-time authentication code is:</p>
				<h1 style="letter-spacing: 5px; color: #2563eb; font-size: 32px;">%s</h1>
				<p>This code expires in 10 minutes. If you did not request this code, you can safely ignore this email.</p>
			</div>
		`, otp),
	}

	_, err := client.Emails.Send(params)
	return err
}
