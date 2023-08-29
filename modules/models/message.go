package models

import (
	"gorm.io/gorm"
	"time"
)

type Message struct {
	ID         uint `gorm:"primaryKey"`
	FromUserID uint
	ToUserID   uint
	Content    string
	CreatedAt  time.Time
}

// Migrate handles the custom migrations required for the Message model.
func (m Message) Migrate(db *gorm.DB) error {
	// Add the first complex composite index
	if err := db.Exec("CREATE INDEX idx_from_to_created ON messages(from_user_id, to_user_id, created_at)").
		Error; err != nil {
		return err
	}

	// Add the second complex composite index
	return db.Exec("CREATE INDEX idx_to_from_created ON messages(to_user_id, from_user_id, created_at)").Error
}
