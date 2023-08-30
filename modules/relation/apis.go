package relation

import (
	"app/modules/models"
	"app/utils"
	"errors"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"net/http"
	"strconv"
	"time"
)

func Action(c *gin.Context) {
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

	// 验证 to_user_id 是不是一个数字
	toUserIdString := c.DefaultQuery("to_user_id", "")
	toUserIdInt, err := strconv.Atoi(toUserIdString)
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
			StatusMsg:  "You can't follow yourself.",
		})
		return
	}

	db := c.MustGet("db").(*gorm.DB)

	// 验证 to_user_id 是一个存在的用户
	var toUser models.User
	if err := db.First(&toUser, toUserIdString).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status_code": 1,
			"status_msg":  "to_user_id not found.",
		})
		return
	}

	actionType := c.DefaultQuery("action_type", "")
	// 执行关注/取关操作
	switch actionType {
	case "1": // 关注
		relation := models.Relation{
			FromUserId: fromUserId,
			ToUserId:   uint(toUserIdInt),
			CreatedAt:  time.Time{},
		}
		if err := db.Create(&relation).Error; err != nil {
			if errors.Is(err, gorm.ErrDuplicatedKey) { // 重复关注
				c.JSON(http.StatusBadRequest, gin.H{
					"status_code": 1,
					"status_msg":  "You've already followed this user.",
				})
				return
			}
			// 其它创建失败错误
			c.JSON(http.StatusInternalServerError, gin.H{
				"status_code": 1,
				"status_msg":  "Failed to follow.",
			})
			return
		}
	case "2": // 取关
		var relationToDelete models.Relation
		tx := db.Where("from_user_id = ? and to_user_id = ?", fromUserId, toUserIdString).
			Delete(&relationToDelete)
		if tx.RowsAffected == 0 { // 删除了0条记录，说明这条关注关系不存在
			c.JSON(http.StatusNotFound, gin.H{
				"status_code": 1,
				"status_msg":  "You haven't followed this user.",
			})
			return
		}
		if tx.Error != nil { // 其它删除失败错误
			c.JSON(http.StatusInternalServerError, gin.H{
				"status_code": 1,
				"status_msg":  "Failed to unfollow.",
			})
			return
		}
	default: // 错误的 action_type
		c.JSON(http.StatusBadRequest, gin.H{
			"status_code": 1,
			"status_msg":  "Invalid action_type.",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status_code": 0,
		"status_msg":  "Success",
	})
	return
}
