package video

import (
	"app/config"
	"app/consts"
	"app/modules/models"
	"app/utils"
	"bytes"
	"fmt"
	"github.com/disintegration/imaging"
	"github.com/gin-gonic/gin"
	"github.com/h2non/filetype"
	ffmpeg "github.com/u2takey/ffmpeg-go"
	"gorm.io/gorm"
	"io"
	"net/http"
	"os"
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

	// 如果当前已登录，我们需要：1. 知道返回的MaxVideos个视频中哪些被用户已经点赞过
	// 2. 知道其中哪些视频发布者是当前登录用户关注的
	var likedVideoIdSet = make(map[uint]bool)
	var followedVideoCreatorIdSet = make(map[uint]bool)
	if isLoggedIn == true {
		// 生成视频 ID 列表和视频发布者 ID 列表
		var videoIds []uint
		var creatorIds []uint
		for _, video := range videos {
			videoIds = append(videoIds, video.ID)
			creatorIds = append(creatorIds, video.UserID)
		}
		// 查询 favorites 表，看看哪些视频被用户点赞过
		var likedVideoIds []uint
		db.Table("favorites").
			Where("user_id = ? AND video_id IN ?", userId, videoIds).
			Pluck("video_id", &likedVideoIds)

		// 查询 relations 表，看看当前用户关注了哪些视频发布者
		var followedCreatorIds []uint
		db.Table("relations").
			Where("from_user_id = ? AND to_user_id IN ?", userId, creatorIds).
			Pluck("to_user_id", &followedCreatorIds)

		// 将视频 ID 放入哈希表
		for _, id := range likedVideoIds {
			likedVideoIdSet[id] = true
		}

		// 将被关注的视频发布者 ID 放入哈希表
		for _, id := range followedCreatorIds {
			followedVideoCreatorIdSet[id] = true
		}
	}

	var videoResList []utils.VideoResItem
	for _, v := range videos {
		// 将查询的数据填充到返回的结构体中
		_, isLiked := likedVideoIdSet[v.ID]
		_, isFollowed := followedVideoCreatorIdSet[v.UserID]
		videoResList = append(videoResList, utils.VideoResItem{
			ID:            v.ID,
			PlayUrl:       v.PlayUrl,
			CoverUrl:      v.CoverUrl,
			FavoriteCount: v.FavoriteCount,
			CommentCount:  v.CommentCount,
			Title:         v.Title,
			IsFavorite:    isLiked,
			Author: utils.UserResponse{
				ID:             v.User.ID,
				Name:           v.User.Username,
				IsFollow:       isFollowed,
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

	// 计算nextTime
	var nextTime int64
	if len(videos) > 0 {
		nextTime = videos[len(videos)-1].PublishTime.UnixMilli()
	}

	resp := utils.VideoResponse{
		StatusCode: 0,
		StatusMsg:  "Success",
		NextTime:   nextTime,
		VideoList:  videoResList,
	}

	c.JSON(http.StatusOK, resp)
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

	// 在这些视频ID中，查询哪些被当前用户点赞过
	tokenString := c.DefaultQuery("token", "")
	currentUserId, _ := utils.ValidateToken(tokenString)
	var likedVideoIds []uint
	db.Table("favorites").
		Where("user_id = ? AND video_id IN (?)", currentUserId, videoIds).
		Pluck("video_id", &likedVideoIds)

	// 将这些ID放进哈希表，以便在O(1)时间内查询某个视频是否被自己点赞过
	var likedVideoIdSet = make(map[uint]bool)
	for _, id := range likedVideoIds {
		likedVideoIdSet[id] = true
	}

	// 查询当前登录用户是否关注了列表的发布者
	var relation models.Relation
	result := db.Where("from_user_id = ? AND to_user_id = ?", currentUserId, userId).First(&relation)
	isFollowed := result.RowsAffected > 0

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
			Author: utils.UserResponse{
				ID:             v.User.ID,
				Name:           v.User.Username,
				IsFollow:       isFollowed,
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

func Publish(c *gin.Context) {
	// 验证视频标题
	title := c.DefaultPostForm("title", "")
	if len(title) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"status_code": 1,
			"status_msg":  "Missing title",
		})
		return
	}

	// 验证视频文件
	file, err := c.FormFile("data")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status_code": 1,
			"status_msg":  "Invalid data",
		})
		fmt.Println("Invalid data")
		return
	}

	// TODO: 验证文件类型为视频类型
	//fileType := file.Header.Get("Content-Type")
	//if !strings.Contains(fileType, "video/") {
	//	c.JSON(http.StatusBadRequest, gin.H{
	//		"status_code": 1,
	//		"status_msg":  "Invalid video data",
	//	})
	//	fmt.Println("Invalid video data")
	//	return
	//}
	openedFile, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status_code": 1,
			"status_msg":  "Failed to open file",
		})
		return
	}
	defer openedFile.Close()

	// 读取文件的前261字节来验证类型
	fileHead := make([]byte, 261)
	_, err = openedFile.Read(fileHead)
	if err != nil && err != io.EOF {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status_code": 1,
			"status_msg":  "Failed to read file",
		})
		return
	}

	// 用 filetype 库验证文件类型
	if !filetype.IsVideo(fileHead) {
		c.JSON(http.StatusBadRequest, gin.H{
			"status_code": 1,
			"status_msg":  "Please provide a video file",
		})
		return
	}

	// 检查文件大小
	if file.Size > consts.MaxVideoSize {
		c.JSON(http.StatusBadRequest, gin.H{
			"status_code": 1,
			"status_msg":  "Video size limit exceeds",
		})
		fmt.Println("Video size limit exceeds")
		return
	}

	// 验证 user_id
	tokenString := c.DefaultPostForm("token", "")
	userId, _ := utils.ValidateToken(tokenString)

	// 生成文件名
	now := time.Now()
	nowUnix := now.UnixMilli()
	filename := fmt.Sprintf("%d-%d", userId, nowUnix)
	videoKey := filename + ".mp4"
	coverKey := filename + ".jpg"
	videoPath := "media/" + filename + ".mp4"
	coverPath := "media/" + filename + ".jpg"

	// TODO: 转码成mp4并压缩

	// 保存文件
	if err := c.SaveUploadedFile(file, videoPath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status_code": 1,
			"status_msg":  "Failed to upload video",
		})
		return
	}
	defer os.Remove(videoPath)
	defer os.Remove(coverPath)

	// 生成视频封面
	if err := GenerateCover(videoPath, coverPath); err != nil {
		os.Remove(videoPath)
		c.JSON(http.StatusInternalServerError, gin.H{
			"status_code": 1,
			"status_msg":  err.Error(),
		})
		return
	}

	videoUrl := fmt.Sprintf(
		"https://s3.%s.amazonaws.com/%s/%s",
		config.AwsBucketRegion,
		consts.AwsBucketName,
		videoKey,
	)

	coverUrl := fmt.Sprintf(
		"https://s3.%s.amazonaws.com/%s/%s",
		config.AwsBucketRegion,
		consts.AwsBucketName,
		coverKey,
	)

	// 更新 videos 表
	videoRecord := models.Video{
		UserID:      userId,
		Title:       title,
		PlayUrl:     videoUrl,
		CoverUrl:    coverUrl,
		PublishTime: now,
	}
	db := c.MustGet("db").(*gorm.DB)
	tx := db.Create(&videoRecord)
	if tx.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status_code": 1,
			"status_msg":  "Failed to create video record",
		})
		return
	}

	err = utils.UploadFileToS3(videoPath, videoKey)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status_code": 1,
			"status_msg":  err.Error(),
		})
		tx.Rollback()
		return
	}

	err = utils.UploadFileToS3(coverPath, coverKey)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status_code": 1,
			"status_msg":  err.Error(),
		})
		tx.Rollback()
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status_code": 0,
		"status_msg":  "Success",
	})
}

func GenerateCover(videoPath, coverPath string) (err error) {
	buf := bytes.NewBuffer(nil)
	if err := ffmpeg.Input(videoPath).
		Filter("select", ffmpeg.Args{fmt.Sprintf("gte(n,%d)", 1)}).
		Output("pipe:", ffmpeg.KwArgs{
			"vframes": 1, "format": "image2", "vcodec": "mjpeg", "pix_fmt": "yuv420p"}).
		WithOutput(buf, os.Stdout).
		Run(); err != nil {
		return fmt.Errorf("failed to process video file")
	}

	img, err := imaging.Decode(buf)
	if err != nil {
		return fmt.Errorf("failed to decode image")
	}

	if err := imaging.Save(img, coverPath); err != nil {
		return fmt.Errorf("failed to save image")
	}

	return nil
}
