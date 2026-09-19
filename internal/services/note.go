package services

import (
	"IAM-server/internal/connections"
	"IAM-server/internal/models"
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// GetCustomerNotes retrieves all notes for a specific customer ID.
func GetCustomerNotes(ctx context.Context, customerID string) ([]models.Note, error) {
	custUUID, err := uuid.Parse(customerID)
	if err != nil {
		return nil, fmt.Errorf("invalid customer ID: %w", err)
	}

	var notes []models.Note
	err = connections.DB.WithContext(ctx).Where("customer_id = ?", custUUID).Order("created_at desc").Find(&notes).Error
	if err != nil {
		return nil, err
	}
	return notes, nil
}

// GetCustomerNote retrieves a single note for a specific customer and note ID.
func GetCustomerNote(ctx context.Context, customerID string, noteID string) (*models.Note, error) {
	custUUID, err := uuid.Parse(customerID)
	if err != nil {
		return nil, fmt.Errorf("invalid customer ID: %w", err)
	}
	nUUID, err := uuid.Parse(noteID)
	if err != nil {
		return nil, fmt.Errorf("invalid note ID: %w", err)
	}

	var note models.Note
	err = connections.DB.WithContext(ctx).Where("customer_id = ? AND id = ?", custUUID, nUUID).First(&note).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("note not found")
		}
		return nil, err
	}
	return &note, nil
}

// CreateCustomerNote creates a new note for a specific customer.
func CreateCustomerNote(ctx context.Context, customerID string, title string, content string) (*models.Note, error) {
	custUUID, err := uuid.Parse(customerID)
	if err != nil {
		return nil, fmt.Errorf("invalid customer ID: %w", err)
	}

	note := models.Note{
		CustomerID: custUUID,
		Title:      title,
		Content:    content,
	}

	err = connections.DB.WithContext(ctx).Create(&note).Error
	if err != nil {
		return nil, err
	}
	return &note, nil
}

// UpdateCustomerNote updates an existing note for a specific customer.
func UpdateCustomerNote(ctx context.Context, customerID string, noteID string, title *string, content *string) (*models.Note, error) {
	custUUID, err := uuid.Parse(customerID)
	if err != nil {
		return nil, fmt.Errorf("invalid customer ID: %w", err)
	}
	nUUID, err := uuid.Parse(noteID)
	if err != nil {
		return nil, fmt.Errorf("invalid note ID: %w", err)
	}

	var note models.Note
	err = connections.DB.WithContext(ctx).Where("customer_id = ? AND id = ?", custUUID, nUUID).First(&note).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("note not found")
		}
		return nil, err
	}

	if title != nil {
		note.Title = *title
	}
	if content != nil {
		note.Content = *content
	}

	err = connections.DB.WithContext(ctx).Save(&note).Error
	if err != nil {
		return nil, err
	}
	return &note, nil
}

// DeleteCustomerNote deletes a note for a specific customer and note ID.
func DeleteCustomerNote(ctx context.Context, customerID string, noteID string) error {
	custUUID, err := uuid.Parse(customerID)
	if err != nil {
		return fmt.Errorf("invalid customer ID: %w", err)
	}
	nUUID, err := uuid.Parse(noteID)
	if err != nil {
		return fmt.Errorf("invalid note ID: %w", err)
	}

	result := connections.DB.WithContext(ctx).Where("customer_id = ? AND id = ?", custUUID, nUUID).Delete(&models.Note{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("note not found")
	}
	return nil
}

