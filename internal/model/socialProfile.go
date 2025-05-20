package model

import (
	"time"

	"gorm.io/gorm"
)

// Provider represents different OAuth providers
type Provider string

const (
	Google   Provider = "google"
	Facebook Provider = "facebook"
	Twitter  Provider = "twitter"
	Github   Provider = "github"
)

type SocialProfile struct {
	ID         uint           `gorm:"primaryKey" json:"id"`
	UserID     uint           `gorm:"not null;index:idx_user_provider,unique" json:"user_id"`
	Provider   Provider       `gorm:"not null;index:idx_user_provider,unique" json:"provider"`
	ProviderID string         `gorm:"not null" json:"provider_id"`
	Name       string         `gorm:"size:255" json:"name"`
	PhotoURL   string         `gorm:"size:2048" json:"photo_url"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	User User `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"-"` // Exclude from JSON serialization
}

// TableName overrides the table name
func (SocialProfile) TableName() string {
	return "social_profiles"
}
