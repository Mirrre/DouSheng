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
	FollowCount    int
	FollowerCount  int
	TotalFavorited int
	WorkCount      int
}
