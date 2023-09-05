package models

import (
	"github.com/u2takey/go-utils/rand"
	"gorm.io/gorm"
	"time"
)

func getRandomIndex() int {
	// 使用当前时间的纳秒数来初始化随机数生成器
	rand.Seed(time.Now().UnixNano())
	// 使用随机索引选择切片中的一个字符串
	return rand.Intn(6)
}

var AvatarUrls = []string{
	"https://c-ssl.duitang.com/uploads/item/202102/04/20210204151116_snwek.jpg",
	"https://c-ssl.duitang.com/uploads/item/202102/04/20210204151117_ufhjb.jpg",
	"https://c-ssl.duitang.com/uploads/item/202102/04/20210204151118_ytkwe.jpg",
	"https://c-ssl.duitang.com/uploads/item/202102/04/20210204151119_mtxcj.jpg",
	"https://c-ssl.duitang.com/uploads/item/202102/04/20210204151120_xxfqi.jpg",
	"https://c-ssl.duitang.com/uploads/item/202102/04/20210204151123_jungh.jpg",
}

var BackgroundUrls = []string{
	"https://c-ssl.duitang.com/uploads/blog/202205/16/20220516164359_e2a6f.jpg",
	"https://c-ssl.duitang.com/uploads/blog/202205/15/20220515195808_2491f.jpeg",
	"https://c-ssl.duitang.com/uploads/blog/202205/15/20220515195807_45e27.jpeg",
	"https://c-ssl.duitang.com/uploads/blog/202207/28/20220728175554_dc1f0.jpeg",
	"https://c-ssl.duitang.com/uploads/blog/202207/28/20220728175555_8db48.jpeg",
	"https://c-ssl.duitang.com/uploads/blog/202207/28/20220728175600_e25d2.jpeg",
}

var Signatures = []string{
	"Injustice anywhere is a threat to justice everywhere.",
	"Be the change that you wish to see in the world.",
	"Imperfection is beauty, madness is genius and it's better to be absolutely ridiculous than absolutely boring.",
	"Success is not final, failure is not fatal: It is the courage to continue that counts.",
	"Be yourself; everyone else is already taken.",
	"The secret of getting ahead is getting started.",
}

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
		UserID:     u.ID,
		Avatar:     AvatarUrls[getRandomIndex()],
		Background: BackgroundUrls[getRandomIndex()],
		Signature:  Signatures[getRandomIndex()],
	}
	if err = tx.Create(&userProfile).Error; err != nil {
		return err
	}
	return nil
}
