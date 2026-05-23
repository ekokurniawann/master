package database

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Role struct {
	ID        uuid.UUID
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func SeedRoles(db *gorm.DB) error {
	roles := []string{"admin", "customer"}

	for _, roleName := range roles {
		var existingRole Role

		err := db.Where("name = ?", roleName).First(&existingRole).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			id, errGen := uuid.NewV7()
			if errGen != nil {
				return fmt.Errorf("database.seeder.SeedRoles: failed to generate uuid: %w", errGen)
			}

			newRole := Role{
				ID:        id,
				Name:      roleName,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}

			if errInsert := db.Create(&newRole).Error; errInsert != nil {
				return fmt.Errorf("database.seeder.SeedRoles: failed to insert role %s: %w", roleName, errInsert)
			}

			fmt.Printf("[SEEDER] Berhasil menambahkan role master: %s\n", roleName)
		} else if err != nil {
			return fmt.Errorf("database.seeder.SeedRoles: query error: %w", err)
		}
	}

	return nil
}
