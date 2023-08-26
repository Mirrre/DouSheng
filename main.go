package main

import (
	"app/config"
	"app/middleware"
	"app/modules/user"
)

func main() {
	dsn := config.SetDsn()
	db, err := config.InitDatabase(dsn)
	if err != nil {
		panic("failed to connect database")
	}

	r := config.InitGinEngine(db)

	r.POST("/douyin/user/register/", user.Register)
	r.GET("/douyin/user/", middleware.Authentication(), user.GetUser)
	r.POST("/douyin/user/login/", user.Login)

	r.Run(":8080")
}
