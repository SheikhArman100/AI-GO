package model

import (
	"time"
)

// DiskType defines the storage type enum
type DiskType string

const (
	DiskTypeLocal DiskType = "local"
	DiskTypeS3    DiskType = "s3"
)

// Image model
type Image struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	DiskType     DiskType  `gorm:"type:varchar(20)" json:"disk_type"`
	Path         string    `gorm:"type:varchar(255);not null" json:"path"`
	OriginalName string    `gorm:"type:varchar(255);not null" json:"original_name"`
	ModifiedName string    `gorm:"type:varchar(255);not null" json:"modified_name"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	UserDetailID *uint     `gorm:"unique" json:"user_detail_id"`

	// Relations
	UserDetail *UserDetail `gorm:"foreignKey:UserDetailID;constraint:OnDelete:CASCADE" json:"user_detail,omitempty"`
}

// TableName overrides the table name for Image
func (Image) TableName() string {
	return "images"
}
