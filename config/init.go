package config

import (
	"app/modules/models"
	"fmt"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"os"
)

// IsTesting 通过设置环境变量来让程序判断当前是测试环境还是生产环境
var IsTesting = os.Getenv("GO_TESTING") == "true"

// InitGinEngine 初始化路由函数
func InitGinEngine(db *gorm.DB) *gin.Engine {
	r := gin.Default()
	r.Use(func(c *gin.Context) {
		c.Set("db", db)
		c.Next()
	})
	return r
}

// InitDatabase 通过dsn来初始化db链接
func InitDatabase(dsn string) (*gorm.DB, error) {
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	// 自动将表单模型结构体迁移成数据库表单
	// TODO: Can we automatically input db models here?
	err = db.AutoMigrate(&models.User{}, &models.UserProfile{}, &models.Video{})
	if err != nil {
		return nil, err
	}
	return db, nil
}

// SetDsn 自动判断测试/生产环境来生成不同的dsn，对应不同的数据库
func SetDsn() string {
	User := os.Getenv("MYSQL_USER")
	Pass := os.Getenv("MYSQL_PASSWORD")
	Host := os.Getenv("MYSQL_HOST")
	Port := os.Getenv("MYSQL_PORT")
	if IsTesting {
		return fmt.Sprintf("%s:%s@tcp(%s:%s)/gotest?charset=utf8mb4&parseTime=True&loc=Local", User, Pass, Host, Port)
	}
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/mysql?charset=utf8mb4&parseTime=True&loc=Local", User, Pass, Host, Port)
}

// Router 设置一个全局路由
var Router *gin.Engine

// SetupRouter 调用初始化路由函数，赋值给Router
func SetupRouter(db *gorm.DB) {
	Router = InitGinEngine(db)
}
