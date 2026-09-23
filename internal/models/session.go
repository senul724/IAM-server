package models

import (
	"time"

	"github.com/google/uuid"
)

type Session struct {
	Base
	RefreshID     string    `gorm:"type:text;not null;index" json:"refresh_id"`
	UserID        uuid.UUID `gorm:"type:uuid;not null;index" json:"user_id"`
	Location      string    `gorm:"type:text"                json:"location"`
	FirstLogin    time.Time `gorm:"not null"                 json:"first_login"`
	LastLogin     time.Time `gorm:"not null"                 json:"last_login"`
	DeviceDetails string    `gorm:"type:text"                json:"device_details"`
	IPAddress     string    `gorm:"type:text"                json:"ip_address"`

	// Association
	User Customer `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"user,omitempty"`
}
