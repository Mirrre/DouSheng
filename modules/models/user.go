package models

import (
	"gorm.io/gorm"
)

// User 表示应用中的用户
type User struct {
	gorm.Model
	Username string `gorm:"unique"`
	Password string
	Profile  UserProfile `gorm:"foreignKey:UserID"`
}

// UserProfile 表示用户的额外信息
type UserProfile struct {
	gorm.Model
	UserID         uint
	Avatar         string
	Background     string
	Signature      string
	FollowCount    int `gorm:"default:0"` // 关注总数
	FollowerCount  int `gorm:"default:0"` // 粉丝总数
	TotalFavorited int `gorm:"default:0"` // 获赞数量
	WorkCount      int `gorm:"default:0"` // 作品数
	FavoriteCount  int `gorm:"default:0"` // 喜欢数
}

func (u *User) AfterCreate(tx *gorm.DB) (err error) {
	userProfile := UserProfile{
		UserID: u.ID,
	}
	if err = tx.Create(&userProfile).Error; err != nil {
		return err
	}
	return nil
}
