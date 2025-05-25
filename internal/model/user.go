package model

import (
	"time"

	"gorm.io/gorm"
)

// UserRole defines the role enum
type UserRole string

// const (
// 	RoleUser  UserRole = "user"
// 	RoleAdmin UserRole = "admin"
// )

// User model
type User struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	Name        string         `gorm:"type:varchar(255)" json:"name"`
	Email       string         `gorm:"type:varchar(255);uniqueIndex;not null" json:"email"`
	PhoneNumber string         `gorm:"type:varchar(20);index" json:"phone_number"`
	Password    string         `gorm:"type:varchar(255);not null" json:"-"` // Hashed password
	IsVerified  bool           `gorm:"default:false" json:"is_verified"`
	Role        UserRole       `gorm:"type:varchar(20);default:user" json:"role"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at"`

	// Relations
	UserDetail    *UserDetail    `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"user_detail,omitempty"`
	RefreshTokens []RefreshToken `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"refresh_tokens,omitempty"`
	SocialProfiles []SocialProfile `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"social_profiles,omitempty"`
	Searches    []Search       `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"searches,omitempty"`
}

// TableName overrides the table name for User
func (User) TableName() string {
	return "users"
}
