package models

type ShortLink struct {
	ID      uint64 `json:"id" gorm:"unique;not null;primaryKey"`
	LongUrl string `json:"long_url" gorm:"not null"`
}
