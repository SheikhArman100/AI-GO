package model

import (
	"gorm.io/gorm"
	"time"
)

type Search struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	Title     string         `gorm:"type:varchar(255);not null" json:"title" validate:"required,max=255"`
	Ip        string         `gorm:"type:varchar(255);not null" json:"ip" validate:"required,ip,max=255"`
	UserID    uint           `gorm:"index;not null" json:"user_id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`

	Responses []Response `gorm:"foreignKey:SearchID;constraint:OnDelete:CASCADE" json:"responses,omitempty"`
}

type Response struct {
	ID                uint                     `gorm:"primaryKey" json:"id"`
	SearchID          uint                     `gorm:"index;not null" json:"search_id"`
	Question          string                   `gorm:"type:varchar(255);not null" json:"question" validate:"required,max=255"`
	Details           string                   `gorm:"type:text;not null" json:"details" validate:"required"`
	RelatedQuestions  []string                 `gorm:"type:json" json:"related_questions"`
	Images            []string                 `gorm:"type:json" json:"images"`
	Charts            []map[string]interface{} `gorm:"type:json" json:"charts"`
	IsRelatedQuestion bool                     `gorm:"default:false" json:"isRelatedQuestion"`
	CreatedAt         time.Time                `json:"created_at"`
	UpdatedAt         time.Time                `json:"updated_at"`
}
