package main

import (
	"app/config"
	"app/consts"
	"app/middleware"
	"app/modules/comment"
	"app/modules/favorite"
	"app/modules/message"
	"app/modules/relation"
	"app/modules/user"
	"app/modules/video"
	"context"
	"github.com/minio/minio-go/v7"
	"log"
)

func main() {
	ctx := context.Background()
	endpoint := "play.min.io"
	accessKeyID := "Q3AM3UQ867SPQQA43P2F"
	secretAccessKey := "zuf+tfteSlswRu7BJ86wekitnifILbZam1KYY3TG"
	useSSL := true

	// Initialize minio client object.
	err := config.InitMinioClient(endpoint, accessKeyID, secretAccessKey, useSSL)
	if err != nil {
		log.Fatalln("Failed to initialize minIO client: ", err)
	}

	// Create bucket if not exists
	bucketName := consts.MinIOBucketName
	err = config.MinioClient.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{})
	if err != nil {
		// Check to see if we already own this bucket (which happens if you run this twice)
		exists, errBucketExists := config.MinioClient.BucketExists(ctx, bucketName)
		if errBucketExists == nil && exists {
			log.Printf("We already own %s\n", bucketName)
		} else {
			log.Fatalf("Failed to create bucket: %s\n", err)
		}
	} else {
		log.Printf("Successfully created %s\n", bucketName)
	}

	dsn := config.SetDsn()
	db, err := config.InitDatabase(dsn)
	config.InitAwsSession()
	if err != nil {
		log.Fatalf("failed to connect database: %s\n", err)
	}

	r := config.InitGinEngine(db)

	r.GET("/douyin/comment/list/", middleware.Authentication(), comment.List)
	r.GET("/douyin/favorite/list/", middleware.Authentication(), favorite.GetLikeVideos)
	r.GET("/douyin/feed/", video.GetFeed)
	r.GET("/douyin/message/chat/", middleware.Authentication(), message.GetHistory)
	r.GET("/douyin/publish/list/", middleware.Authentication(), video.GetUserVideos)
	r.GET("/douyin/relation/follow/list/", middleware.Authentication(), relation.GetFollowings)
	r.GET("/douyin/relation/follower/list/", middleware.Authentication(), relation.GetFollowers)
	r.GET("/douyin/relation/friend/list/", middleware.Authentication(), relation.GetFriends)
	r.GET("/douyin/user/", middleware.Authentication(), user.GetUser)
	r.POST("/douyin/comment/action/", middleware.Authentication(), comment.Action)
	r.POST("/douyin/favorite/action/", middleware.Authentication(), favorite.Action)
	r.POST("/douyin/message/action/", middleware.Authentication(), message.Send)
	r.POST("/douyin/publish/action/", middleware.Authentication(), video.PublishToMinIO)
	//r.POST("/douyin/publish/action/", middleware.Authentication(), video.Publish)
	r.POST("/douyin/relation/action/", middleware.Authentication(), relation.Action)
	r.POST("/douyin/user/login/", user.Login)
	r.POST("/douyin/user/register/", user.Register)

	err = r.Run(":8080")
	if err != nil {
		panic("failed to run server.")
	}
}
