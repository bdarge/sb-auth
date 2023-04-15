package models

import (
	"gorm.io/gorm"
)

// Account for each user
type Account struct {
	gorm.Model
	Email    string `gorm:"unique;not null;default:null" json:"email"`
	Password string `gorm:"not null;default:null" json:"-"`
	UserID   int    `json:"userId"`
}
