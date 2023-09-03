package main

import (
	"app/config"
	"app/middleware"
	"app/modules/comment"
	"app/modules/favorite"
	"app/modules/message"
	"app/modules/relation"
	"app/modules/user"
	"app/modules/video"
)

func main() {
	dsn := config.SetDsn()
	db, err := config.InitDatabase(dsn)
	config.InitAwsSession()
	if err != nil {
		panic("failed to connect database.")
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
	r.POST("/douyin/publish/action/", middleware.Authentication(), video.Publish)
	r.POST("/douyin/relation/action/", middleware.Authentication(), relation.Action)
	r.POST("/douyin/user/login/", user.Login)
	r.POST("/douyin/user/register/", user.Register)

	err = r.Run(":8080")
	if err != nil {
		panic("failed to run server.")
	}
}
