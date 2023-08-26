package video

import (
	"app/modules/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"net/http"
	"strconv"
	"time"
)

const MaxVideos = 30

func GetFeed(c *gin.Context) {
	var videos []models.Video

	latestTimeString := c.DefaultQuery("latest_time", "")
	if latestTimeString == "" {
		latestTimeString = time.Now().Format("2006-01-02 15:04:05")
	}

	var latestTime time.Time
	var err error

	// 尝试解析为 "2006-01-02 15:04:05" 格式
	latestTime, err = time.ParseInLocation("2006-01-02 15:04:05", latestTimeString, time.Local)
	if err != nil {
		// 如果失败，尝试解析为Unix时间戳
		unixTime, unixErr := strconv.ParseInt(latestTimeString, 10, 64)
		if unixErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid latest_time format. Expected '2006-01-02 15:04:05' or Unix timestamp.",
			})
			return
		}
		latestTime = time.Unix(unixTime, 0)
	}

	db := c.MustGet("db").(*gorm.DB)
	if err := db.Where("publish_time < ?", latestTime).Order("publish_time desc").
		Limit(MaxVideos).Find(&videos).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, videos)
}
