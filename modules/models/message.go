package models

type Message struct {
	ID         uint `gorm:"primary_key"`
	FromUserID uint `gorm:"index:idx_from_to_created,priority:1;index:idx_to_from_created,priority:2"`
	ToUserID   uint `gorm:"index:idx_from_to_created,priority:2;index:idx_to_from_created,priority:1"`
	Content    string
	CreatedAt  int64 `gorm:"index:idx_from_to_created,priority:3;index:idx_to_from_created,priority:3"`
}
