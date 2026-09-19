package services

import (
	"IAM-server/internal/connections"
	"IAM-server/internal/models"
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/resend/resend-go/v2"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func LoginWithPassword(ctx context.Context, email string, password string) (*models.Customer, error) {
	var customer models.Customer
	err := connections.DB.WithContext(ctx).Where("email = ?", email).First(&customer).Error
	if err != nil {
		return nil, err
	}

	if customer.Password == nil || *customer.Password == "" {
		return nil, errors.New("password not set for this user")
	}

	err = bcrypt.CompareHashAndPassword([]byte(*customer.Password), []byte(password))
	if err != nil {
		return nil, err
	}

	return &customer, nil
}

// SendResetPasswordOTP checks if the user exists, generates an OTP under the "reset" prefix,
// and sends an email with the verification code.
func SendResetPasswordOTP(ctx context.Context, email string) error {
	if connections.Redis == nil {
		return errors.New("redis connection is not initialized")
	}

	var customer models.Customer
	err := connections.DB.Where("email = ?", email).First(&customer).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("user does not exist")
		}
		return err
	}

	otp, err := GenerateOTP(ctx, email, OTPReasonReset)
	if err != nil {
		return fmt.Errorf("failed to generate OTP: %w", err)
	}

	err = sendPasswordResetEmail(email, otp)
	if err != nil {
		return fmt.Errorf("failed to send password reset email: %w", err)
	}

	return nil
}

// ResetPassword validates the OTP from Redis for the "reset" prefix, hashes the new password,
// and updates it in the database.
func ResetPassword(ctx context.Context, email string, otp string, newPassword string) error {
	if connections.Redis == nil {
		return errors.New("redis connection is not initialized")
	}

	if newPassword == "" {
		return errors.New("new password cannot be empty")
	}

	err := ValidateOTP(ctx, email, otp, OTPReasonReset)
	if err != nil {
		return err
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	hashStr := string(hashedPassword)
	result := connections.DB.Model(&models.Customer{}).Where("email = ?", email).Update("password", &hashStr)
	if result.Error != nil {
		return fmt.Errorf("failed to update password: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return errors.New("user does not exist")
	}

	return nil
}

// sendPasswordResetEmail sends an email containing the reset OTP using Resend.
func sendPasswordResetEmail(toEmail, otp string) error {
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
		Subject: "Password Reset Verification Code",
		Html: fmt.Sprintf(`
			<div style="font-family: sans-serif; padding: 20px; line-height: 1.5; color: #111;">
				<h2>Password Reset</h2>
				<p>Your one-time password reset code is:</p>
				<h1 style="letter-spacing: 5px; color: #2563eb; font-size: 32px;">%s</h1>
				<p>This code expires in 10 minutes. If you did not request a password reset, you can safely ignore this email.</p>
			</div>
		`, otp),
	}

	_, err := client.Emails.Send(params)
	return err
}
