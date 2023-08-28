package models

import "time"

type Comment struct {
	ID        uint `gorm:"primary_key;not null"`
	UserID    uint `gorm:"not null"`
	VideoID   uint `gorm:"index:idx_video_comment_created,unique"`
	Content   string
	CreatedAt time.Time
}
