package models

import (
	"gorm.io/gorm"
	"time"
)

type Favorite struct {
	ID        uint `gorm:"primary_key"`
	UserID    uint `gorm:"index:idx_user_video,unique"`
	VideoID   uint `gorm:"index:idx_user_video,unique"`
	CreatedAt time.Time
}

// 创建Hook，在点赞/取消点赞记录生成后，自动给: 1. 视频的favorite_count +/- 1
// 2. 为视频发布者的Profile中的获赞数量TotalFavorited +/- 1
// 3. 为点赞者的Profile中的喜欢数FavoriteCount +/- 1
// 如果更新失败，点赞/取消点赞也会被回滚，保持数据一致性

func (f *Favorite) AfterCreate(tx *gorm.DB) (err error) {
	// 加载Video，以获取发布该视频的用户
	var video Video
	if err = tx.First(&video, f.VideoID).Error; err != nil {
		return err
	}

	// 更新视频的favorite_count
	if err = tx.Model(&Video{}).Where("id = ?", f.VideoID).
		UpdateColumn("favorite_count", gorm.Expr("favorite_count + 1")).Error; err != nil {
		return err
	}

	// 更新发布视频用户的Profile中的获赞数量TotalFavorited
	if err = tx.Model(&UserProfile{}).Where("user_id = ?", video.UserID).
		UpdateColumn("total_favorited", gorm.Expr("total_favorited + 1")).Error; err != nil {
		return err
	}

	// 更新点赞者的Profile中的喜欢数FavoriteCount
	if err = tx.Model(&UserProfile{}).Where("user_id = ?", f.UserID).
		UpdateColumn("favorite_count", gorm.Expr("favorite_count + 1")).Error; err != nil {
		return err
	}

	return nil
}

func (f *Favorite) AfterDelete(tx *gorm.DB) (err error) {
	// 加载Video，以获取发布该视频的用户
	var video Video
	if err = tx.First(&video, f.VideoID).Error; err != nil {
		return err
	}

	// 更新视频的favorite_count
	if err = tx.Model(&Video{}).Where("id = ?", f.VideoID).
		UpdateColumn("favorite_count", gorm.Expr(
			"CASE WHEN favorite_count > 0 THEN favorite_count - 1 ELSE 0 END")).Error; err != nil {
		return err
	}

	// 更新发布视频用户的Profile中的TotalFavorited
	if err = tx.Model(&UserProfile{}).Where("user_id = ?", video.UserID).
		UpdateColumn("total_favorited", gorm.Expr(
			"CASE WHEN total_favorited > 0 THEN total_favorited - 1 ELSE 0 END")).Error; err != nil {
		return err
	}

	// 更新点赞者的Profile中的喜欢数FavoriteCount
	if err = tx.Model(&UserProfile{}).Where("user_id = ?", f.UserID).
		UpdateColumn("favorite_count", gorm.Expr(
			"CASE WHEN favorite_count > 0 THEN favorite_count - 1 ELSE 0 END")).Error; err != nil {
		return err
	}

	return nil
}
