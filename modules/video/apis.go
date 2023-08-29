package video

import (
	"app/modules/models"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"gorm.io/gorm"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"app/util"
)

const MaxVideos = 5

type FeedResponse struct {
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
	var videos []models.Video

	latestTimeString := c.DefaultQuery("latest_time", "")
	if latestTimeString == "" {
		// 将当前时间转换为毫秒单位的Unix时间戳
		latestTimeString = fmt.Sprintf("%d", time.Now().UnixNano()/1e6)
	}

	// 尝试将输入的字符串解析为毫秒为单位的Unix时间戳
	unixTimeMs, err := strconv.ParseInt(latestTimeString, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, FeedResponse{
			StatusCode: 1,
			StatusMsg:  "Error: Invalid latest_time format. Expected Unix timestamp in milliseconds.",
		})
		return
	}

	// 将毫秒单位的Unix时间戳转换为time.Time对象
	latestTime := time.Unix(0, unixTimeMs*1e6)
	fmt.Println("Latest time: ", latestTime)
	db := c.MustGet("db").(*gorm.DB)
	err = db.Preload("User").Preload("User.Profile").
		Where("publish_time < ?", latestTime).Order("publish_time desc").
		Limit(10).Find(&videos).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, FeedResponse{
			StatusCode: 1,
			StatusMsg:  "Error: Can't fetch videos.",
		})
		return
	}

	var videoResList []FeedVideoRes
	for _, v := range videos {
		// 将查询的数据填充到返回的结构体中
		videoResList = append(videoResList, FeedVideoRes{
			ID:      v.ID,
			PlayUrl: v.PlayUrl,
			// "sdcard/DCIM/Camera/TG-2023-05-22-1541367851684741298128.mp4"
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
		//fmt.Println(v.PlayUrl, "+", v.CoverUrl)
	}
	// 计算nextTime
	var nextTime int64
	if len(videos) > 0 {
		nextTime = videos[len(videos)-1].PublishTime.Unix()
	}

	resp := FeedResponse{
		StatusCode: 0,
		StatusMsg:  "Success",
		NextTime:   nextTime,
		VideoList:  videoResList,
	}

	c.JSON(http.StatusOK, resp)
}

func Submission(c *gin.Context) {
	tokenString := c.PostForm("token")
	title := c.PostForm("title")
	user_id, _ := util.ValidateToken(tokenString)
	fmt.Println(title)
	fmt.Println("user_id:", user_id)

	// 获取上传的视频文件
	videoFile, err := c.FormFile("data")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status_code": 1,
			"status_msg":  "upload video data error",
		})
		fmt.Println("上传视频数据错误")
		return
	}
	// 创建视频存储路径
	storagePath := filepath.Join("temp", "videos")
	os.MkdirAll(storagePath, os.ModePerm)

	// 保存视频至本地
	videoFilePath := filepath.Join(storagePath, videoFile.Filename)
	if err := c.SaveUploadedFile(videoFile, videoFilePath); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status_code": 1,
			"status_msg":  "Failed to save video file",
		})
		fmt.Println("无法保存视频数据到本地")
		return
	}
	// 创建MinIO客户端
	minioClient, err := minio.New("192.168.10.10:9000", &minio.Options{
		Creds:  credentials.NewStaticV4("ROOTUSER", "CHANGEME123", ""),
		Secure: false,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status_code": 1,
			"status_msg":  "Failed to connect to MinIO",
		})
		fmt.Println(err)
		fmt.Println("连接MinlO失败")
		return
	}
	// 设置存储桶名称
	bucketName := "videos"
	// 检查存储桶是否存在，如果不存在则创建
	found, err := minioClient.BucketExists(c.Request.Context(), bucketName)
	fmt.Println(found)
	if found == false {
		err := minioClient.MakeBucket(c.Request.Context(), bucketName, minio.MakeBucketOptions{})
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"status_code": 1,
				"status_msg":  "Failed to create bucket",
			})
			fmt.Println(err)
			return
		}
	}
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status_code": 1,
			"status_msg":  "Failed to found bucket",
		})
		fmt.Println(err)
	}

	// 上传视频至Minio
	objectName := fmt.Sprintf("%d/%s.mp4", user_id, title)
	n, err := minioClient.FPutObject(c.Request.Context(), bucketName, objectName, videoFilePath, minio.PutObjectOptions{
		ContentType: "video/mp4",
	})
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{
			"status_code": 1,
			"status_msg":  "Failed to uploda video",
		})
		fmt.Println("无法上传视频数据")
		return
	}
	fmt.Println("Successfully uploaded bytes: ", n)
	presignedURL, err := minioClient.PresignedGetObject(c.Request.Context(), bucketName, objectName, time.Hour*24, nil)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status_code": 1,
			"status_msg":  "Failed to create presignedURL",
		})
		fmt.Println("生成视频链接的错误:", err)
		return
	}

	//fmt.Println("生成的视频链接:", presignedURL.String())
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status_code": 1,
			"status_msg":  "Failed to upload to Minio",
		})
		return
	}

	//视频地址
	playUrl := presignedURL.String()
	//TODO:
	coverUrl := presignedURL.String()
	// 在数据库中创建新的视频记录
	newVideo := models.Video{
		UserID:      user_id,
		Title:       title,
		PlayUrl:     playUrl,
		CoverUrl:    coverUrl,
		PublishTime: time.Now(),
	}
	db := c.MustGet("db").(*gorm.DB)
	db.Create(&newVideo)

	c.JSON(http.StatusOK, gin.H{
		"status_code": 0,
		"status_msg":  "video upload uccess",
	})

}
