package entity

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AuthLog struct {
	ID           uuid.UUID
	UserID       *uuid.UUID
	Event        string
	Status       string
	IPAddress    string
	UserAgent    string
	ErrorMessage *string
	CreatedAt    time.Time

	User *User `gorm:"foreignKey:UserID;references:ID"`
}

func (l *AuthLog) BeforeCreate(tx *gorm.DB) error {
	id, err := uuid.NewV7()
	if err != nil {
		return err
	}
	l.ID = id
	return nil
}
