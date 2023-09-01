package models

import (
	"gorm.io/gorm"
	"time"
)

type Relation struct {
	ID         uint      `gorm:"primaryKey"`
	FromUserId uint      `gorm:"index:idx_from_user;index:idx_relationship,unique"`
	ToUserId   uint      `gorm:"index:idx_to_user;index:idx_relationship,unique"`
	CreatedAt  time.Time `gorm:"autoCreateTime"`
	ToUser     User      `gorm:"foreignKey:ToUserId"`
	FromUser   User      `gorm:"foreignKey:FromUserId"`
}

// AfterCreate hook for the Relation model.
func (relation *Relation) AfterCreate(tx *gorm.DB) (err error) {
	// 1. 被关注者的粉丝数+1
	err = tx.Model(&UserProfile{}).Where("user_id = ?", relation.ToUserId).
		UpdateColumn("follower_count", gorm.Expr("follower_count + 1")).Error
	if err != nil {
		return err
	}

	// 2. 关注者的关注数+1
	err = tx.Model(&UserProfile{}).Where("user_id = ?", relation.FromUserId).
		UpdateColumn("follow_count", gorm.Expr("follow_count + 1")).Error
	return err
}

// AfterDelete hook for the Relation model.
func (relation *Relation) AfterDelete(tx *gorm.DB) (err error) {
	// 1. 被关注者的粉丝数-1 (确保不会小于0)
	err = tx.Model(&UserProfile{}).Where("user_id = ?", relation.ToUserId).
		UpdateColumn("follower_count", gorm.Expr(
			"CASE WHEN follower_count > 0 THEN follower_count - 1 ELSE 0 END")).Error
	if err != nil {
		return err
	}

	// 2. 关注者的关注数-1 (确保不会小于0)
	err = tx.Model(&UserProfile{}).Where("user_id = ?", relation.FromUserId).
		UpdateColumn("follow_count", gorm.Expr(
			"CASE WHEN follow_count > 0 THEN follow_count - 1 ELSE 0 END")).Error
	return err
}
