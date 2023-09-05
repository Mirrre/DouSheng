package config

import (
	"app/modules/models"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/gin-gonic/gin"
	"github.com/minio/minio-go/v7"
	minioCred "github.com/minio/minio-go/v7/pkg/credentials"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"log"
	"os"
	"sync"
)

// IsTesting 通过设置环境变量来让程序判断当前是单元测试还是运行后端服务
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
	err = db.AutoMigrate(&models.User{}, &models.UserProfile{},
		&models.Video{}, &models.Favorite{},
		&models.Comment{}, &models.Message{},
		&models.Relation{},
	)
	if err != nil {
		return nil, err
	}
	return db, nil
}

// SetDsn 自动判断测试/生产环境来生成不同的dsn，对应不同的数据库
func SetDsn() string {
	//User := os.Getenv("MYSQL_USER")
	//Pass := os.Getenv("MYSQL_PASSWORD")
	//Host := os.Getenv("MYSQL_HOST")
	//Port := os.Getenv("MYSQL_PORT")
	User := "root"
	Pass := "lps123456"
	Host := "127.0.0.1"
	Port := "3306"
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

// 设置全局单例 AWS 客户端
var (
	onceAwsSession  sync.Once
	sess            *session.Session
	S3Client        *s3.S3
	err             error
	AwsBucketRegion = os.Getenv("AWS_BUCKET_REGION")
)

func InitAwsSession() {
	onceAwsSession.Do(func() {
		sess, err = session.NewSession(&aws.Config{
			Region: aws.String(AwsBucketRegion),
			Credentials: credentials.NewStaticCredentials(
				os.Getenv("AWS_ACCESS_KEY_ID"),
				os.Getenv("AWS_SECRET_ACCESS_KEY"),
				"",
			),
		})
		if err != nil {
			log.Fatalf("Failed to initialize AWS session: %v", err)
		}

		S3Client = s3.New(sess)
	})
}

// 设置全局单例 minIO 客户端
var (
	MinioClient *minio.Client
	once        sync.Once
)

func InitMinioClient(endpoint, accessKeyID, secretAccessKey string, useSSL bool) error {
	var err error

	once.Do(func() {
		MinioClient, err = minio.New(endpoint, &minio.Options{
			Creds:  minioCred.NewStaticV4(accessKeyID, secretAccessKey, ""),
			Secure: useSSL,
		})
	})

	return err
}
