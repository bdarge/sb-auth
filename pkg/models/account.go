package models

import (
	"gorm.io/gorm"
	"time"
)

type Model struct {
	ID        uint32         `json:"id,string"` // https://stackoverflow.com/a/21152548
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `json:"deletedAt"`
}

// Account for each user
type Account struct {
	Model
	Email    string `gorm:"unique;not null;default:null" json:"email"`
	Password string `gorm:"not null;default:null" json:"-"`
	UserID   int    `json:"userId"`
}
