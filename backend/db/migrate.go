// Purpose: Dev/local migrations. Creates tables via GORM AutoMigrate.
// В проде лучше использовать SQL-миграции, но для старта проекта это самый простой путь.

package db

import "gorm.io/gorm"

func AutoMigrate(database *gorm.DB) error {
	return database.AutoMigrate(
		&User{},
		&Track{},
		&QueueEntry{},
		&Donation{},
		&Integration{},
	)
}
