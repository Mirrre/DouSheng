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

// AfterCreate hook for the Video model.
func (video *Video) AfterCreate(tx *gorm.DB) (err error) {
	// 发布者的作品数 + 1
	err = tx.Model(&UserProfile{}).Where("user_id = ?", video.UserID).
		UpdateColumn("work_count", gorm.Expr("work_count + 1")).Error
	if err != nil {
		return err
	}
	return nil
}

// AfterDelete hook for the Video model.
func (video *Video) AfterDelete(tx *gorm.DB) (err error) {
	// 发布者的作品数 - 1
	err = tx.Model(&UserProfile{}).Where("user_id = ?", video.UserID).
		UpdateColumn("work_count", gorm.Expr(
			"CASE WHEN work_count > 0 THEN work_count - 1 ELSE 0 END")).Error
	if err != nil {
		return err
	}
	return nil
}
