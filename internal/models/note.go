package models

import (
	"github.com/google/uuid"
)

type Note struct {
	Base
	CustomerID uuid.UUID `gorm:"type:uuid;not null;index" json:"customer_id"`
	Pinned     bool      `gorm:"default:false;not null" json:"pinned"`
	Content    string    `gorm:"type:text;not null" json:"content"`
	Title      string    `gorm:"type:text;not null" json:"title"`

	// Association
	Customer Customer `gorm:"foreignKey:CustomerID" json:"customer,omitempty"`
}
