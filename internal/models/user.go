package models

type Customer struct {
	Base
	Name     string  `gorm:"not null"   json:"name"`
	Email    string  `gorm:"uniqueIndex;not null" json:"email"`
	Password string  `gorm:"not null"             json:"-"`
	PhotoURL *string `gorm:"default:null"                   json:"photo_url,omitempty"`

	// Associations
	Notes []Note `gorm:"foreignKey:CustomerID" json:"notes,omitempty"`
}
