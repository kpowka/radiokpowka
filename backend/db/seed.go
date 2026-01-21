// Purpose: Seed admin user on startup (idempotent) + bcrypt helpers.

package db

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"radiokpowka/backend/config"
)

func SeedAdmin(database *gorm.DB, cfg config.Config) error {
	if cfg.AdminUsername == "" || cfg.AdminPassword == "" {
		return errors.New("ADMIN_USERNAME/ADMIN_PASSWORD must be set (or defaults used)")
	}

	var cnt int64
	if err := database.Model(&User{}).Where("username = ?", cfg.AdminUsername).Count(&cnt).Error; err != nil {
		return err
	}
	if cnt > 0 {
		return nil
	}

	hash, err := HashPassword(cfg.AdminPassword)
	if err != nil {
		return err
	}

	u := User{
		ID:           uuid.New(),
		Username:     cfg.AdminUsername,
		PasswordHash: hash,
		Role:         "owner",
		CreatedAt:    time.Now().UTC(),
	}
	return database.Create(&u).Error
}

func HashPassword(plain string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.DefaultCost)
	return string(b), err
}

func CheckPassword(hash, plain string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain)) == nil
}
