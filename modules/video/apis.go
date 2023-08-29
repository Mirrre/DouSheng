package video

import (
	"app/consts"
	"app/modules/models"
	"app/utils"
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
)

// GetFeed 视频流接口，返回早于latest_time发布的MaxVideos个视频
func GetFeed(c *gin.Context) {
	latestTimeString := c.DefaultQuery("latest_time", "")
	if latestTimeString == "" {
		// 将当前时间转换为毫秒单位的Unix时间戳
		latestTimeString = fmt.Sprintf("%d", time.Now().UnixNano()/1e6)
	}

	// 尝试将输入的字符串解析为毫秒为单位的Unix时间戳
	unixTimeMs, err := strconv.ParseInt(latestTimeString, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.VideoResponse{
			StatusCode: 1,
			StatusMsg:  "Error: Invalid latest_time format. Expected Unix timestamp in milliseconds.",
		})
		return
	}

	// 将毫秒单位的Unix时间戳转换为time.Time对象
	latestTime := time.Unix(0, unixTimeMs*1e6)

	// 找出所有发布时间早于latestTime的视频
	var videos []models.Video
	db := c.MustGet("db").(*gorm.DB)
	err = db.Preload("User").Preload("User.Profile").
		Where("publish_time < ?", latestTime).Order("publish_time desc").
		Limit(consts.MaxVideos).Find(&videos).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.VideoResponse{
			StatusCode: 1,
			StatusMsg:  "Error: Can't fetch videos.",
		})
		return
	}

	// 检查当前登录状态
	tokenString := c.DefaultQuery("token", "")
	userId, _ := utils.ValidateToken(tokenString)
	isLoggedIn := userId > 0

	// 如果当前已登录，我们需要知道返回的MaxVideos个视频中哪些被用户已经点赞过
	var likedVideoIdSet = make(map[uint]bool)
	if isLoggedIn == true {
		// 生成视频ID列表
		var videoIds []uint
		for _, video := range videos {
			videoIds = append(videoIds, video.ID)
		}
		// 查询favorites表，看看哪些视频被用户点赞过
		var likedVideoIds []uint
		db.Table("favorites").
			Where("user_id = ? AND video_id in (?)\n", userId, videoIds).
			Pluck("video_id", &likedVideoIds)

		for _, id := range likedVideoIds {
			likedVideoIdSet[id] = true
		}
	}

	var videoResList []utils.VideoResItem
	for _, v := range videos {
		// 将查询的数据填充到返回的结构体中
		_, isLiked := likedVideoIdSet[v.ID]
		videoResList = append(videoResList, utils.VideoResItem{
			ID:            v.ID,
			PlayUrl:       v.PlayUrl,
			CoverUrl:      v.CoverUrl,
			FavoriteCount: v.FavoriteCount,
			CommentCount:  v.CommentCount,
			Title:         v.Title,
			IsFavorite:    isLiked,
			Author: utils.Author{
				ID:             v.User.ID,
				Name:           v.User.Username,
				Avatar:         v.User.Profile.Avatar,
				Background:     v.User.Profile.Background,
				Signature:      v.User.Profile.Signature,
				FollowCount:    v.User.Profile.FollowCount,
				FollowerCount:  v.User.Profile.FollowerCount,
				TotalFavorited: v.User.Profile.TotalFavorited,
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

	resp := utils.VideoResponse{
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
	userId, _ := utils.ValidateToken(tokenString)
	fmt.Println(title)
	fmt.Println("user_id:", userId)

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
	if err := os.MkdirAll(storagePath, os.ModePerm); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status_code": 1,
			"status_msg":  "Failed to create video path",
		})
		fmt.Println("无法创建视频存储路径")
		return
	}

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
	objectName := fmt.Sprintf("%d/%s.mp4", userId, title)
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
		UserID:      userId,
		Title:       title,
		PlayUrl:     playUrl,
		CoverUrl:    coverUrl,
		PublishTime: time.Now(),
	}
	db := c.MustGet("db").(*gorm.DB)
	db.Create(&newVideo)

	c.JSON(http.StatusOK, gin.H{
		"status_code": 0,
		"status_msg":  "video upload successfully.",
	})
}

func GetUserVideos(c *gin.Context) {
	// 验证 user_id
	userId := c.DefaultQuery("user_id", "0")
	if userIdInt, err := strconv.Atoi(userId); err != nil || userIdInt < 1 {
		c.JSON(http.StatusBadRequest, utils.VideoResponse{
			StatusCode: 1,
			StatusMsg:  "Invalid user_id.",
		})
		return
	}

	db := c.MustGet("db").(*gorm.DB)

	// 获取用户的投稿列表
	var videos []models.Video

	err := db.Preload("User").Preload("User.Profile").
		Where("user_id = ?", userId).Order("publish_time desc").
		Find(&videos).Error
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.VideoResponse{
			StatusCode: 1,
			StatusMsg:  "Error fetching videos.",
		})
		return
	}

	// 获取投稿视频的ID列表
	var videoIds []uint
	for _, v := range videos {
		videoIds = append(videoIds, v.ID)
	}

	// 在这些视频ID中，查询哪些被自己点赞过
	var likedVideoIds []uint
	db.Table("favorites").
		Where("user_id = ? AND video_id IN (?)", userId, videoIds).
		Pluck("video_id", &likedVideoIds)

	// 将这些ID放进哈希表，以便在O(1)时间内查询某个视频是否被自己点赞过
	var likedVideoIdSet = make(map[uint]bool)
	for _, id := range likedVideoIds {
		likedVideoIdSet[id] = true
	}

	var videoResList []utils.VideoResItem
	for _, v := range videos {
		_, isLiked := likedVideoIdSet[v.ID]
		videoResList = append(videoResList, utils.VideoResItem{
			ID:            v.ID,
			PlayUrl:       v.PlayUrl,
			CoverUrl:      v.CoverUrl,
			FavoriteCount: v.FavoriteCount,
			CommentCount:  v.CommentCount,
			Title:         v.Title,
			IsFavorite:    isLiked,
			Author: utils.Author{
				ID:             v.User.ID,
				Name:           v.User.Username,
				Avatar:         v.User.Profile.Avatar,
				Background:     v.User.Profile.Background,
				Signature:      v.User.Profile.Signature,
				FollowCount:    v.User.Profile.FollowCount,
				FollowerCount:  v.User.Profile.FollowerCount,
				TotalFavorited: v.User.Profile.TotalFavorited,
				WorkCount:      v.User.Profile.WorkCount,
				FavoriteCount:  v.User.Profile.FavoriteCount,
			},
		})
	}

	c.JSON(http.StatusOK, utils.VideoResponse{
		StatusCode: 0,
		StatusMsg:  "Success",
		VideoList:  videoResList,
	})
}
