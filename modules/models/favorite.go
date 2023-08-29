package models

import (
	"time"
)

type Favorite struct {
	ID        uint `gorm:"primary_key"`
	UserID    uint `gorm:"index:idx_user_video,unique"`
	VideoID   uint `gorm:"index:idx_user_video,unique"`
	CreatedAt time.Time
}
