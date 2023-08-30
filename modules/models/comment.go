package models

import (
	"fmt"
	"gorm.io/gorm"
	"time"
)

type Comment struct {
	ID        uint `gorm:"primary_key;not null"`
	UserID    uint `gorm:"not null"`
	User      User `gorm:"foreignKey:UserID"`
	VideoID   uint `gorm:"index:idx_video_comment_created;not null"`
	Content   string
	CreatedAt time.Time `gorm:"index:idx_video_comment_created"`
}

func (f *Comment) AfterCreate(tx *gorm.DB) (err error) {
	fmt.Println("Video ID: ", f.VideoID)
	tx.Model(&Video{}).Where("id = ?", f.VideoID).
		UpdateColumn("comment_count", gorm.Expr("comment_count + 1"))
	return
}

func (f *Comment) AfterDelete(tx *gorm.DB) (err error) {
	fmt.Println("Video ID: ", f.VideoID)
	tx.Model(&Video{}).Where("id = ?", f.VideoID).
		UpdateColumn("comment_count", gorm.Expr(
			"CASE WHEN comment_count > 0 THEN comment_count - 1 ELSE 0 END"))
	return
}
