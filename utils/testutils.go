// 定义了一些单元测试的辅助函数，在测试之前将测试数据库的表单迁移好，
// 在测试之后清除测试数据。
// 对于每一个包的测试，顺序是 Setup -> 测试 -> Teardown

package utils

import (
	"app/config"
	"app/modules/models"
	"fmt"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"log"
)

var TestRouter *gin.Engine
var dsn = config.SetDsn()
var db, err = config.InitDatabase(dsn)

func Setup() {
	if err != nil {
		log.Fatal("Failed to connect database.")
	}
	config.SetupRouter(db)
}

func Teardown() {
	TestRouter = nil
	err := db.Migrator().DropTable(&models.User{}, &models.UserProfile{}, &models.Message{}, &models.Relation{})
	if err != nil {
		fmt.Println("Failed to drop DB table.")
	}
}

func GetDb() *gorm.DB {
	return db
}
