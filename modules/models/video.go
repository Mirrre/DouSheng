package models

import (
	"gorm.io/gorm"
	"time"
)

type Video struct {
	gorm.Model
	UserID        uint      `gorm:"index:idx_user_created" json:"user_id"`
	User          User      `gorm:"foreignKey:UserID"`
	Title         string    `json:"title"`
	PlayUrl       string    `json:"play_url"`
	CoverUrl      string    `json:"cover_url"`
	FavoriteCount uint      `gorm:"default:0;not null" json:"favorite_count"`
	CommentCount  uint      `gorm:"default:0;not null" json:"comment_count"`
	PublishTime   time.Time `gorm:"index:idx_publish_time;index:idx_user_created" json:"published_at"`
}
