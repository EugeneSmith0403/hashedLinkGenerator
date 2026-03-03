package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Link struct {
	gorm.Model
	UserID uint   `json:"userId"`
	Url    string `json:"url"`
	Hash   string `json:"hash" gorm:"uniqueIndex"`
}

func NewLink(url string, userID uint) *Link {
	return &Link{
		UserID: userID,
		Url:    url,
		Hash:   uuid.NewString(),
	}
}
