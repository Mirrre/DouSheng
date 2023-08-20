package main

import (
  "app/middleware"
  "app/modules/models"
  "app/modules/user"
  "app/utils"
  "fmt"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
  "os"
)

func main() {
	// 连接到MySQL数据库
  User := os.Getenv("MYSQL_USER")
  Pass := os.Getenv("MYSQL_PASSWORD")
  Host := os.Getenv("MYSQL_HOST")
  Port := os.Getenv("MYSQL_PORT")
  
  dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/mysql?charset=utf8mb4&parseTime=True&loc=Local", User, Pass, Host, Port)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
  // TODO: defer close db
	if err != nil {
		panic("failed to connect database")
	}

	// 迁移数据库模式
	db.AutoMigrate(&models.User{}, &models.UserProfile{})

	r := gin.Default()

  // 注册中间件将db实例传递给每个处理函数
  r.Use(func(c *gin.Context) {
      c.Set("db", db)
      c.Next()
  })
  
	// 设置路由处理函数
	r.POST("/douyin/user/register/", user.Register)
	r.GET("/douyin/user/", middleware.Authentication(), user.GetUser)
  r.POST("/douyin/user/login/", user.Login)

	r.Run(":8080")
}
