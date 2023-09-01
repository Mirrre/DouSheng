package comment

import (
	"app/consts"
	"app/modules/models"
	"app/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"net/http"
	"strconv"
	"time"
)

func Action(c *gin.Context) {
	// 验证user_id
	tokenString := c.DefaultQuery("token", "")
	userId, err := utils.ValidateToken(tokenString)
	if err != nil || userId <= 0 {
		c.JSON(http.StatusBadRequest, utils.CommentResponse{
			StatusCode: 1,
			StatusMsg:  "Invalid User ID.",
		})
		return
	}

	db := c.MustGet("db").(*gorm.DB)
	var user models.User
	if err := db.Preload("Profile").First(&user, userId).Error; err != nil {
		c.JSON(http.StatusBadRequest, utils.CommentResponse{
			StatusCode: 1,
			StatusMsg:  "Failed to fetch user information.",
		})
		return
	}

	// 验证video_id
	videoId := c.DefaultQuery("video_id", "0")
	videoIdInt, err := strconv.Atoi(videoId)
	if err != nil || videoIdInt < 1 {
		c.JSON(http.StatusBadRequest, utils.CommentResponse{
			StatusCode: 1,
			StatusMsg:  "Invalid Video ID.",
		})
		return
	}
	// 验证视频是否存在
	var video models.Video
	if err := db.First(&video, videoId).Error; err != nil {
		c.JSON(http.StatusNotFound, utils.CommentResponse{
			StatusCode: 1,
			StatusMsg:  "Video not found.",
		})
		return
	}

	actionType := c.DefaultQuery("action_type", "")
	switch actionType {
	case "1": // 评论
		commentText := c.DefaultQuery("comment_text", "")
		// 验证评论长度
		if len(commentText) == 0 {
			c.JSON(http.StatusBadRequest, utils.CommentResponse{
				StatusCode: 1,
				StatusMsg:  "Empty comment is not allowed",
			})
			return
		}
		if len(commentText) > consts.MaxCommentLength {
			c.JSON(http.StatusBadRequest, utils.CommentResponse{
				StatusCode: 1,
				StatusMsg:  "Comment is too long",
			})
			return
		}

		// 在数据库中创建评论
		comment := models.Comment{
			UserID:    userId,
			VideoID:   uint(videoIdInt),
			Content:   commentText,
			CreatedAt: time.Now(),
		}
		result := db.Create(&comment)
		if result.Error != nil {
			c.JSON(http.StatusInternalServerError, utils.CommentResponse{
				StatusCode: 1,
				StatusMsg:  "Failed to comment.",
			})
			return
		}

		author := utils.UserResponse{
			ID:            user.ID,
			Name:          user.Username,
			FollowCount:   user.Profile.FollowCount,
			FollowerCount: user.Profile.FollowerCount,
			// TODO: IsFollow
			Avatar:         user.Profile.Avatar,
			Background:     user.Profile.Background,
			Signature:      user.Profile.Signature,
			TotalFavorited: user.Profile.TotalFavorited,
			WorkCount:      user.Profile.WorkCount,
			FavoriteCount:  user.Profile.FavoriteCount,
		}
		c.JSON(http.StatusOK, utils.CommentResponse{
			StatusCode: 0,
			StatusMsg:  "Successfully commented.",
			Comment: &utils.CommentResItem{
				ID:         comment.ID,
				User:       author,
				Content:    comment.Content,
				CreateDate: comment.CreatedAt,
			},
		})
		return
	case "2": // 删除评论
		commentId := c.DefaultQuery("comment_id", "")
		// 验证comment_id
		if commentIdInt, err := strconv.Atoi(commentId); err != nil || commentIdInt <= 0 {
			c.JSON(http.StatusBadRequest, utils.CommentResponse{
				StatusCode: 1,
				StatusMsg:  "Invalid comment ID.",
			})
			return
		}
		// 验证要删除的评论是否存在
		var commentToDelete models.Comment
		if err := db.First(&commentToDelete, commentId).Error; err != nil {
			c.JSON(http.StatusNotFound, utils.CommentResponse{
				StatusCode: 1,
				StatusMsg:  "Target comment not found.",
			})
			return
		}
		// 验证要删除的评论的评论人是不是当前登录用户
		if commentToDelete.UserID != userId {
			c.JSON(http.StatusForbidden, utils.CommentResponse{
				StatusCode: 1,
				StatusMsg:  "You do not have permission to delete this comment.",
			})
			return
		}
		// 删除评论
		if err := db.Delete(&commentToDelete).Error; err != nil {
			c.JSON(http.StatusInternalServerError, utils.CommentResponse{
				StatusCode: 1,
				StatusMsg:  "Failed to delete comment.",
			})
			return
		}
	default: // action_type 不合法
		c.JSON(http.StatusBadRequest, utils.CommentResponse{
			StatusCode: 1,
			StatusMsg:  "Invalid action type.",
		})
		return
	}
	c.JSON(http.StatusOK, utils.CommentResponse{
		StatusCode: 0,
		StatusMsg:  "Successfully deleted comment.",
	})
}

// List 列出视频下面的所有评论（时间倒序）
func List(c *gin.Context) {
	// 验证video_id
	videoId := c.DefaultQuery("video_id", "0")
	videoIdInt, err := strconv.Atoi(videoId)
	if err != nil || videoIdInt < 1 {
		c.JSON(http.StatusBadRequest, utils.CommentListResponse{
			StatusCode: 1,
			StatusMsg:  "Invalid Video ID.",
		})
		return
	}

	db := c.MustGet("db").(*gorm.DB)

	// 验证视频是否存在
	var video models.Video
	if err := db.First(&video, videoId).Error; err != nil {
		c.JSON(http.StatusNotFound, utils.CommentResponse{
			StatusCode: 1,
			StatusMsg:  "Video not found.",
		})
		return
	}

	var commentList []models.Comment
	result := db.Preload("User").Where("video_id = ?", videoId).
		Find(&commentList).Order("created_at desc")
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, utils.CommentListResponse{
			StatusCode: 1,
			StatusMsg:  "Failed to fetch comments",
		})
		return
	}

	// 查询评论列表中有哪些评论者是当前用户关注的
	tokenString := c.DefaultQuery("token", "")
	currentUserId, _ := utils.ValidateToken(tokenString)
	var commenterIdsSet = make(map[uint]bool)
	for _, c := range commentList {
		commenterIdsSet[c.UserID] = true
	}
	var commenterIds []uint
	for id := range commenterIdsSet {
		commenterIds = append(commenterIds, id)
	}
	var followedIds []uint
	db.Table("relations").
		Where("from_user_id = ? AND to_user_id IN ?", currentUserId, commenterIds).
		Pluck("to_user_id", &followedIds)
	var followedIdSet = make(map[uint]bool)
	for _, id := range followedIds {
		followedIdSet[id] = true
	}

	var commentListResponses []utils.CommentResItem
	for _, comment := range commentList {
		isFollowed := followedIdSet[comment.UserID]
		commentListResponses = append(commentListResponses, utils.CommentResItem{
			ID: comment.ID,
			User: utils.UserResponse{
				ID:             comment.User.ID,
				Name:           comment.User.Username,
				FollowCount:    comment.User.Profile.FollowCount,
				FollowerCount:  comment.User.Profile.FollowerCount,
				IsFollow:       isFollowed,
				Avatar:         comment.User.Profile.Avatar,
				Background:     comment.User.Profile.Background,
				Signature:      comment.User.Profile.Signature,
				TotalFavorited: comment.User.Profile.TotalFavorited,
				WorkCount:      comment.User.Profile.WorkCount,
				FavoriteCount:  comment.User.Profile.FavoriteCount,
			},
			Content:    comment.Content,
			CreateDate: comment.CreatedAt,
		})
	}
	c.JSON(http.StatusOK, utils.CommentListResponse{
		StatusCode:  0,
		StatusMsg:   "Success",
		CommentList: &commentListResponses,
	})
}
