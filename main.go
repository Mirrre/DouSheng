package main

import (
	"app/config"
	"app/middleware"
	"app/modules/comment"
	"app/modules/favorite"
	"app/modules/user"
	"app/modules/video"
)

func main() {
	dsn := config.SetDsn()
	db, err := config.InitDatabase(dsn)
	if err != nil {
		panic("failed to connect database.")
	}

	r := config.InitGinEngine(db)

	r.POST("/douyin/user/register/", user.Register)
	r.GET("/douyin/user/", middleware.Authentication(), user.GetUser)
	r.POST("/douyin/user/login/", user.Login)
	r.GET("/douyin/feed/", video.GetFeed)
	r.GET("/douyin/publish/list/", middleware.Authentication(), video.GetUserVideos)
	r.POST("/douyin/favorite/action/", middleware.Authentication(), favorite.Action)
	r.GET("/douyin/favorite/list/", middleware.Authentication(), favorite.GetLikeVideos)
	r.POST("/douyin/comment/action/", middleware.Authentication(), comment.Action)
	r.GET("/douyin/comment/list/", middleware.Authentication(), comment.List)

	err = r.Run(":8080")
	if err != nil {
		panic("failed to run server.")
	}
}
