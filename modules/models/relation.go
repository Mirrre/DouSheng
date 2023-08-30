package models

import "time"

type Relation struct {
	ID         uint      `gorm:"primaryKey"`
	FromUserId uint      `gorm:"index:idx_from_user;index:idx_relationship,unique"`
	ToUserId   uint      `gorm:"index:idx_to_user;index:idx_relationship,unique"`
	CreatedAt  time.Time `gorm:"autoCreateTime"`
	ToUser     User      `gorm:"foreignKey:ToUserId"`
	FromUser   User      `gorm:"foreignKey:FromUserId"`
}

// TODO: hooks
