package services

import (
	"IAM-server/internal/connections"
	"IAM-server/internal/models"
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// CreateSession creates a new session in the database for the given user.
func CreateSession(
	ctx context.Context,
	userID uuid.UUID,
	refreshID string,
	ip string,
	deviceDetails string,
	location string,
) (*models.Session, error) {
	now := time.Now()
	session := models.Session{
		UserID:        userID,
		RefreshID:     refreshID,
		IPAddress:     ip,
		DeviceDetails: deviceDetails,
		Location:      location,
		FirstLogin:    now,
		LastLogin:     now,
	}

	if err := connections.DB.WithContext(ctx).Create(&session).Error; err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	return &session, nil
}

// UpdateSessionRefreshID updates an existing session matching oldRefreshID with the new refreshID and current login info.
func UpdateSessionRefreshID(
	ctx context.Context,
	oldRefreshID string,
	newRefreshID string,
) error {
	var session models.Session
	err := connections.DB.WithContext(ctx).Where("refresh_id = ?", oldRefreshID).First(&session).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("session not found")
		}
		return fmt.Errorf("database error fetching session: %w", err)
	}

	updates := map[string]any{
		"refresh_id": newRefreshID,
		"last_login": time.Now(),
	}

	if err := connections.DB.WithContext(ctx).Model(&session).Updates(updates).Error; err != nil {
		return fmt.Errorf("failed to update session: %w", err)
	}

	return nil
}

// DeleteSessionByRefreshID deletes the session record corresponding to the given refreshID.
func DeleteSessionByRefreshID(ctx context.Context, refreshID string) error {
	if refreshID == "" {
		return nil
	}
	result := connections.DB.WithContext(ctx).Where("refresh_id = ?", refreshID).Delete(&models.Session{})
	return result.Error
}

// DeleteSessionByID deletes a specific session by its ID, ensuring it belongs to the user,
// and invalidates its refresh and access tokens in Redis before deletion.
func DeleteSessionByID(ctx context.Context, sessionID uuid.UUID, userID uuid.UUID) error {
	var session models.Session
	err := connections.DB.WithContext(ctx).Where("id = ? AND user_id = ?", sessionID, userID).First(&session).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("session not found")
		}
		return fmt.Errorf("failed to fetch session: %w", err)
	}

	// Delete relevant refresh and access token records from Redis
	if session.RefreshID != "" {
		_ = RevokeTokenPairByRefreshJTI(ctx, session.RefreshID)
	}

	// Delete the session record from DB
	if err := connections.DB.WithContext(ctx).Delete(&session).Error; err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}

	return nil
}
