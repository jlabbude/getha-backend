package main

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

var DATABASE *gorm.DB

type Aparelho struct {
	ID         uuid.UUID `gorm:"type:uuid;default:gen_random_uuid()"`
	Nome       string    `gorm:"type:text"`
	ImagePath  string    `gorm:"type:text"`
	VideoPath  string    `gorm:"type:text"`
	ManualPath string    `gorm:"type:text"`
}
