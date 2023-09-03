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
		relationToDelete := models.Relation{
			FromUserId: fromUserId,
			ToUserId:   uint(toUserIdInt),
		}
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

// GetFollowings 查询关注列表
func GetFollowings(c *gin.Context) {
	// 获取 user_id 参数
	userIdString := c.DefaultQuery("user_id", "")
	// 验证 user_id
	userIdInt, err := strconv.Atoi(userIdString)
	if err != nil || userIdInt <= 0 {
		c.JSON(http.StatusBadRequest, utils.CommentResponse{
			StatusCode: 1,
			StatusMsg:  "Invalid target user ID.",
		})
		return
	}

	// 查找用户的关注列表
	var relationships []models.Relation
	db := c.MustGet("db").(*gorm.DB)
	if err := db.Preload("ToUser").Preload("ToUser.Profile").
		Where("from_user_id = ?", userIdString).
		Find(&relationships).Error; err != nil { // 如果查询失败
		c.JSON(http.StatusInternalServerError, gin.H{
			"status_code": 1,
			"status_msg":  "Failed to fetch followings.",
		})
		return
	}

	// 生成 user_list
	var userList []utils.UserResponse
	for _, relation := range relationships {
		userList = append(userList, utils.UserResponse{
			ID:             relation.ToUser.ID,
			Name:           relation.ToUser.Username,
			FollowCount:    relation.ToUser.Profile.FollowCount,
			FollowerCount:  relation.ToUser.Profile.FollowerCount,
			IsFollow:       true,
			Avatar:         relation.ToUser.Profile.Avatar,
			Background:     relation.ToUser.Profile.Background,
			Signature:      relation.ToUser.Profile.Signature,
			TotalFavorited: relation.ToUser.Profile.TotalFavorited,
			WorkCount:      relation.ToUser.Profile.WorkCount,
			FavoriteCount:  relation.ToUser.Profile.FavoriteCount,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"status_code": 1,
		"status_msg":  "Success",
		"user_list":   userList,
	})
}

// GetFollowers 查询粉丝列表
func GetFollowers(c *gin.Context) {
	// 获取 user_id 参数
	userIdString := c.DefaultQuery("user_id", "")
	// 验证 user_id
	userIdInt, err := strconv.Atoi(userIdString)
	if err != nil || userIdInt <= 0 {
		c.JSON(http.StatusBadRequest, utils.CommentResponse{
			StatusCode: 1,
			StatusMsg:  "Invalid target user ID.",
		})
		return
	}

	// 查找用户的粉丝列表
	var relationships []models.Relation
	db := c.MustGet("db").(*gorm.DB)
	if err := db.Preload("FromUser").Preload("FromUser.Profile").
		Where("to_user_id = ?", userIdString).
		Find(&relationships).Error; err != nil { // 如果查询失败
		c.JSON(http.StatusInternalServerError, gin.H{
			"status_code": 1,
			"status_msg":  "Failed to fetch followers.",
		})
		return
	}

	// 查询当前用户关注了哪些他的粉丝
	var followerIds []uint // 先将粉丝ID放进一个列表，用来传给数据库查询的 IN 操作
	for _, relation := range relationships {
		followerIds = append(followerIds, relation.FromUserId)
	}
	// userFollowedFansIds 将查询结果放入这个列表
	var userFollowedFansIds []uint
	db.Table("relations").
		Where("from_user_id = ? AND to_user_id IN (?)", userIdString, followerIds).
		Pluck("to_user_id", &userFollowedFansIds)
	// 将 userFollowedFansIds 放入 userFollowedFansSet 哈希表以便在 O(1) 时间内查询
	var userFollowedFansSet = make(map[uint]bool)
	for _, id := range userFollowedFansIds {
		userFollowedFansSet[id] = true
	}

	// 生成 user_list
	var userList []utils.UserResponse
	for _, relation := range relationships {
		_, isFollowed := userFollowedFansSet[relation.FromUserId]
		userList = append(userList, utils.UserResponse{
			ID:             relation.FromUser.ID,
			Name:           relation.FromUser.Username,
			FollowCount:    relation.FromUser.Profile.FollowCount,
			FollowerCount:  relation.FromUser.Profile.FollowerCount,
			IsFollow:       isFollowed,
			Avatar:         relation.FromUser.Profile.Avatar,
			Background:     relation.FromUser.Profile.Background,
			Signature:      relation.FromUser.Profile.Signature,
			TotalFavorited: relation.FromUser.Profile.TotalFavorited,
			WorkCount:      relation.FromUser.Profile.WorkCount,
			FavoriteCount:  relation.FromUser.Profile.FavoriteCount,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"status_code": 0,
		"status_msg":  "Success",
		"user_list":   userList,
	})
}

func GetFriends(c *gin.Context) {
	// 获取 user_id 参数
	userIdString := c.DefaultQuery("user_id", "")
	// 验证 user_id
	userIdInt, err := strconv.Atoi(userIdString)
	if err != nil || userIdInt <= 0 {
		c.JSON(http.StatusBadRequest, utils.CommentResponse{
			StatusCode: 1,
			StatusMsg:  "Invalid target user ID.",
		})
		return
	}

	db := c.MustGet("db").(*gorm.DB)

	// 获取 user_id 的关注列表
	var followingIds []uint
	if err := db.Table("relations").Select("to_user_id").
		Where("from_user_id = ?", userIdString).
		Pluck("to_user_id", &followingIds).Error; err != nil {
		c.JSON(http.StatusInternalServerError, utils.CommentResponse{
			StatusCode: 1,
			StatusMsg:  "Failed to fetch data.",
		})
		return
	}

	// 如果用户的关注列表为空，那么他肯定没有好友
	if len(followingIds) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"status_code": 0,
			"status_msg":  "Success",
			"user_list":   followingIds,
		})
		return
	}

	// 查询 followingIds 中有哪些人关注了 user_id
	var friends []models.Relation
	if tx := db.Preload("ToUser").Preload("ToUser.Profile").
		Where("from_user_id = ? AND to_user_id IN (?)", userIdString, followingIds).
		Find(&friends); tx.Error != nil {
		c.JSON(http.StatusInternalServerError, utils.CommentResponse{
			StatusCode: 1,
			StatusMsg:  "Failed to fetch data.",
		})
		return
	}

	if len(friends) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"status_code": 0,
			"status_msg":  "Success",
			"user_list":   friends,
		})
		return
	}

	// 生成 user_list
	var userList []utils.UserResponse
	for _, friend := range friends {
		userList = append(userList, utils.UserResponse{
			ID:             friend.ToUser.ID,
			Name:           friend.ToUser.Username,
			FollowCount:    friend.ToUser.Profile.FollowCount,
			FollowerCount:  friend.ToUser.Profile.FollowerCount,
			IsFollow:       true,
			Avatar:         friend.ToUser.Profile.Avatar,
			Background:     friend.ToUser.Profile.Background,
			Signature:      friend.ToUser.Profile.Signature,
			TotalFavorited: friend.ToUser.Profile.TotalFavorited,
			WorkCount:      friend.ToUser.Profile.WorkCount,
			FavoriteCount:  friend.ToUser.Profile.FavoriteCount,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"status_code": 0,
		"status_msg":  "Success",
		"user_list":   userList,
	})
}
