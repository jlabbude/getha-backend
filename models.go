package main

import (
	"gorm.io/gorm"
)

var DATABASE *gorm.DB

type Aparelho struct {
	ID         uint     `gorm:"primaryKey;autoIncrement;unique"`
	Nome       string   `gorm:"type:text"`
	ImagemPath string   `gorm:"type:text"`
	Videos     []Video  `gorm:"foreignKey:AparelhoID"`
	Manuais    []Manual `gorm:"foreignKey:AparelhoID"`
}

type Manual struct {
	ID         uint   `gorm:"primaryKey;autoIncrement;unique"`
	ManualPath string `gorm:"type:text"`
	AparelhoID uint
	Aparelho   Aparelho `gorm:"foreignKey:AparelhoID;references:ID"`
}

type Video struct {
	ID         uint   `gorm:"primaryKey;autoIncrement;unique"`
	VideoPath  string `gorm:"type:text"`
	AparelhoID uint
	Aparelho   Aparelho `gorm:"foreignKey:AparelhoID;references:ID"`
}
