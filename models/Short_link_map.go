package models

import "gorm.io/gorm"

type ShortLink struct {
	gorm.Model
	ShortKey string `json:"short_code" gorm:"unique;not null"`
	LongUrl  string `json:"short_url" gorm:"not null"`
}
