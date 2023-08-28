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

// 创建Hook，在点赞/取消点赞记录生成后，自动给视频的favorite_count +/- 1
// 如果更新favorite_count失败，点赞/取消点赞也会被回滚，保持数据一致性

func (f *Favorite) AfterCreate(tx *gorm.DB) (err error) {
	tx.Model(&Video{}).Where("id = ?", f.VideoID).
		UpdateColumn("favorite_count", gorm.Expr("favorite_count + ?", 1))
	//fmt.Println("increment favorite_count by 1")
	return
}

func (f *Favorite) AfterDelete(tx *gorm.DB) (err error) {
	tx.Model(&Video{}).Where("id = ?", f.VideoID).
		UpdateColumn("favorite_count", gorm.Expr("favorite_count - ?", 1))
	//fmt.Println("decrement favorite_count by 1")
	return
}
