package models

import (
	"gorm.io/gorm"
	"time"
)

type Comment struct {
	ID        uint `gorm:"primary_key;not null"`
	UserID    uint `gorm:"not null"`
	User      User `gorm:"foreignKey:UserID"`
	VideoID   uint `gorm:"index:idx_video_comment_created"`
	Content   string
	CreatedAt time.Time `gorm:"index:idx_video_comment_created"`
}

func (f *Comment) AfterCreate(tx *gorm.DB) (err error) {
	tx.Model(&Video{}).Where("id = ?", f.VideoID).
		UpdateColumn("favorite_count", gorm.Expr("comment_count + ?", 1))
	//fmt.Println("increment favorite_count by 1")
	return
}

func (f *Comment) AfterDelete(tx *gorm.DB) (err error) {
	tx.Model(&Video{}).Where("id = ?", f.VideoID).
		UpdateColumn("favorite_count", gorm.Expr("comment_count - ?", 1))
	//fmt.Println("decrement favorite_count by 1")
	return
}
