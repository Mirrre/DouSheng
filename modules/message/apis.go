package message

import (
	"app/modules/models"
	"app/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"net/http"
	"strconv"
	"time"
)

func Send(c *gin.Context) {
	// 获取token
	tokenString := c.DefaultQuery("token", "")
	// 验证 from_user_id
	fromUserId, err := utils.ValidateToken(tokenString)
	if err != nil || fromUserId <= 0 {
		c.JSON(http.StatusBadRequest, utils.CommentResponse{
			StatusCode: 1,
			StatusMsg:  "Invalid user ID.",
		})
		return
	}

	db := c.MustGet("db").(*gorm.DB)

	// 验证 to_user_id 是不是一个数字
	toUserId := c.DefaultQuery("to_user_id", "")
	toUserIdInt, err := strconv.Atoi(toUserId)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.CommentResponse{
			StatusCode: 1,
			StatusMsg:  "Invalid target user ID.",
		})
		return
	}
	// 验证 from_user_id 和 to_user_id 是不是同一个 id
	if uint(toUserIdInt) == fromUserId {
		c.JSON(http.StatusBadRequest, utils.CommentResponse{
			StatusCode: 1,
			StatusMsg:  "You can't send message to yourself.",
		})
		return
	}
	// 验证 to_user_id 是一个存在的用户
	var toUser models.User
	if err := db.First(&toUser, toUserId).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status_code": 1,
			"status_msg":  "to_user_id not found.",
		})
		return
	}

	// 验证 action_type
	actionType := c.DefaultQuery("action_type", "")
	if actionType != "1" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status_code": 1,
			"status_msg":  "Invalid action type.",
		})
		return
	}

	// 验证 content
	content := c.DefaultQuery("content", "")
	if len(content) == 0 || len(content) > 512 {
		c.JSON(http.StatusBadRequest, gin.H{
			"status_code": 1,
			"status_msg":  "Message length must be between 1 and 512.",
		})
		return
	}

	// 发送消息
	// 创建消息对象
	message := models.Message{
		FromUserID: fromUserId,
		ToUserID:   toUser.ID,
		Content:    content,
		CreatedAt:  time.Time{},
	}
	// 消息存入数据库
	if db.Create(&message).Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status_code": 1,
			"status_msg":  "Failed to send message.",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"status_code": 0,
		"status_msg":  "Message sent.",
	})
}
