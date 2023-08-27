package favorite

import (
	"app/modules/models"
	"app/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"net/http"
	"strconv"
)

func Action(c *gin.Context) {
	videoId := c.DefaultQuery("video_id", "0")
	tokenString := c.DefaultQuery("token", "")
	actionType := c.DefaultQuery("action_type", "")

	// validate video_id
	videoIdInt, err := strconv.Atoi(videoId)
	if err != nil || videoIdInt < 1 {
		c.JSON(http.StatusBadRequest, gin.H{
			"status_code":    1,
			"status_message": "Invalid video_id",
		})
		return
	}

	userId, _ := utils.ValidateToken(tokenString)
	db := c.MustGet("db").(*gorm.DB)

	// validate action type and perform action accordingly
	switch actionType {
	case "1": // Favorite
		favorite := models.Favorite{
			UserID:  userId,
			VideoID: uint(videoIdInt),
		}
		// Check if this video has been liked by current user
		var count int64
		db.Model(&models.Favorite{}).Where("user_id = ? AND video_id = ?", userId, videoIdInt).Count(&count)
		if count > 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"status_code": 1,
				"status_msg":  "Already favortited",
			})
			return
		}
		// Failed to like for some reason
		if err := db.Create(&favorite).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status_code": 1,
				"status_msg":  "Failed to favorite",
			})
			return
		}
		// Add 1 to video.FavoriteCount
		if err := db.Model(&models.Video{}).Where("id = ?", videoIdInt).
			UpdateColumn("favorite_count", gorm.Expr("favorite_count + ?", 1)).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status_code": 1,
				"status_msg":  "Failed to update video's favorite count",
			})
			return
		}
	case "2": // Un-favorite
		// If current user hasn't liked this video, unlike is not allowed
		var count int64
		db.Model(&models.Favorite{}).Where("user_id = ? AND video_id = ?", userId, videoIdInt).Count(&count)
		if count == 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"status_code": 1,
				"status_msg":  "Not favorite yet",
			})
			return
		}
		// Failed to delete the like record for some reason
		if err := db.Where("user_id = ? AND video_id = ?", userId, videoIdInt).
			Delete(&models.Favorite{}).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status_code": 1,
				"status_msg":  "Failed to un-favorite",
			})
			return
		}
		// Video.FavoriteCount - 1
		if err := db.Model(&models.Video{}).Where("id = ?", videoIdInt).
			UpdateColumn("favorite_count", gorm.Expr("favorite_count - ?", 1)).Error; err != nil { // Note the subtraction here
			c.JSON(http.StatusInternalServerError, gin.H{
				"status_code": 1,
				"status_msg":  "Failed to update video's favorite count",
			})
			return
		}
	default:
		// If actionType is not 1 nor 2
		c.JSON(http.StatusBadRequest, gin.H{
			"status_code": 1,
			"status_msg":  "Invalid action type",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status_code": 0,
		"status_msg":  "Success",
	})
}
