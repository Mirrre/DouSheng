package video

import (
	"app/modules/models"
	"fmt"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"net/http"
	"strconv"
	"time"
)

const MaxVideos = 5

type Response struct {
	StatusCode int            `json:"status_code"`
	StatusMsg  string         `json:"status_msg"`
	NextTime   int64          `json:"next_time"`
	VideoList  []FeedVideoRes `json:"video_list"`
}

type FeedVideoRes struct {
	ID            uint      `json:"id"`
	Author        AuthorRes `json:"author"`
	PlayUrl       string    `json:"play_url"`
	CoverUrl      string    `json:"cover_url"`
	FavoriteCount uint      `json:"favorite_count"`
	CommentCount  uint      `json:"comment_count"`
	IsFavorite    bool      `json:"is_favorite"`
	Title         string    `json:"title"`
}

type AuthorRes struct {
	ID             uint   `json:"id"`
	Name           string `json:"name"`
	FollowCount    int    `json:"follow_count"`
	FollowerCount  int    `json:"follower_count"`
	IsFollow       bool   `json:"is_follow"`
	Avatar         string `json:"avatar"`
	Background     string `json:"background_image"`
	Signature      string `json:"signature"`
	TotalFavorited string `json:"total_favorited"`
	WorkCount      int    `json:"work_count"`
	FavoriteCount  int    `json:"favorite_count"`
}

func GetFeed(c *gin.Context) {
	latestTimeString := c.DefaultQuery("latest_time", "")
	if latestTimeString == "" {
		// 将当前时间转换为毫秒单位的Unix时间戳
		latestTimeString = fmt.Sprintf("%d", time.Now().UnixNano()/1e6)
	}

	// 尝试将输入的字符串解析为毫秒为单位的Unix时间戳
	unixTimeMs, err := strconv.ParseInt(latestTimeString, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, Response{
			StatusCode: 1,
			StatusMsg:  "Error: Invalid latest_time format. Expected Unix timestamp in milliseconds.",
		})
		return
	}

	// 将毫秒单位的Unix时间戳转换为time.Time对象
	latestTime := time.Unix(0, unixTimeMs*1e6)

	var videos []models.Video
	db := c.MustGet("db").(*gorm.DB)
	err = db.Preload("User").Preload("User.Profile").
		Where("publish_time < ?", latestTime).Order("publish_time desc").
		Limit(MaxVideos).Find(&videos).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			StatusCode: 1,
			StatusMsg:  "Error: Can't fetch videos.",
		})
		return
	}

	var videoResList []FeedVideoRes
	for _, v := range videos {
		// 将查询的数据填充到返回的结构体中
		videoResList = append(videoResList, FeedVideoRes{
			ID:            v.ID,
			PlayUrl:       v.PlayUrl,
			CoverUrl:      v.CoverUrl,
			FavoriteCount: v.FavoriteCount,
			CommentCount:  v.CommentCount,
			Title:         v.Title,
			Author: AuthorRes{
				ID:             v.User.ID,
				Name:           v.User.Username,
				Avatar:         v.User.Profile.Avatar,
				Background:     v.User.Profile.Background,
				Signature:      v.User.Profile.Signature,
				FollowCount:    v.User.Profile.FollowCount,
				FollowerCount:  v.User.Profile.FollowerCount,
				TotalFavorited: strconv.Itoa(v.User.Profile.TotalFavorited),
				WorkCount:      v.User.Profile.WorkCount,
				FavoriteCount:  v.User.Profile.FavoriteCount,
			},
		})
	}

	// 计算nextTime
	var nextTime int64
	if len(videos) > 0 {
		nextTime = videos[len(videos)-1].PublishTime.Unix()
	}

	resp := Response{
		StatusCode: 0,
		StatusMsg:  "Success",
		NextTime:   nextTime,
		VideoList:  videoResList,
	}

	c.JSON(http.StatusOK, resp)
}

func GetUserVideos(c *gin.Context) {
	// validate user_id
	userId := c.DefaultQuery("user_id", "0")
	if userIdInt, err := strconv.Atoi(userId); err != nil || userIdInt < 1 {
		c.JSON(http.StatusBadRequest, Response{
			StatusCode: 1,
			StatusMsg:  "Invalid user_id.",
		})
		return
	}

	// fetch videos published by user_id
	db := c.MustGet("db").(*gorm.DB)

	var videos []models.Video
	err := db.Preload("User").Preload("User.Profile").
		Where("user_id = ?", userId).Order("publish_time desc").
		Find(&videos).Error
	if err != nil {
		c.JSON(http.StatusBadRequest, Response{
			StatusCode: 1,
			StatusMsg:  "Error fetching videos.",
		})
		return
	}

	var videoResList []FeedVideoRes
	for _, v := range videos {
		videoResList = append(videoResList, FeedVideoRes{
			ID:            v.ID,
			PlayUrl:       v.PlayUrl,
			CoverUrl:      v.CoverUrl,
			FavoriteCount: v.FavoriteCount,
			CommentCount:  v.CommentCount,
			Title:         v.Title,
			Author: AuthorRes{
				ID:             v.User.ID,
				Name:           v.User.Username,
				Avatar:         v.User.Profile.Avatar,
				Background:     v.User.Profile.Background,
				Signature:      v.User.Profile.Signature,
				FollowCount:    v.User.Profile.FollowCount,
				FollowerCount:  v.User.Profile.FollowerCount,
				TotalFavorited: strconv.Itoa(v.User.Profile.TotalFavorited),
				WorkCount:      v.User.Profile.WorkCount,
				FavoriteCount:  v.User.Profile.FavoriteCount,
			},
		})
	}

	c.JSON(http.StatusOK, Response{
		StatusCode: 0,
		StatusMsg:  "Success",
		VideoList:  videoResList,
	})
}
