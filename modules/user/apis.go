package user

import (
	"app/modules/models"
	"app/utils"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"net/http"
	"strconv"
)

// Register 处理用户注册的API请求
func Register(c *gin.Context) {
	var user models.User

	user.Username = c.Query("username")
	user.Password = c.Query("password")
	if len(user.Username) < 6 || len(user.Password) < 6 || len(user.Username) > 25 || len(user.Password) > 25 {
		c.JSON(http.StatusBadRequest, gin.H{
			"status_code": 1,
			"status_msg":  "Username and Password should be non-empty and between 6 - 25 characters.",
			"user_id":     nil,
			"username":    user.Username,
		})
		fmt.Println(http.StatusBadRequest, "Username and Password should be non-empty and between 6 - 25 characters.")
		return
	}

	// 使用GORM将用户数据存储到数据库中
	db := c.MustGet("db").(*gorm.DB)
	if err := db.Create(&user).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status_code": 1,
			"status_msg":  "Failed to register.",
			"user_id":     nil,
		})
		fmt.Println(http.StatusBadRequest, "Failed to register.")
		return
	}

	// 生成新Token
	newToken, err := utils.GenerateToken(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status_code": 1,
			"status_msg":  "Failed to generate token.",
			"user_id":     user.ID,
		})
		fmt.Println(http.StatusInternalServerError, "Failed to generate token.")
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"status_code": 0,
		"status_msg":  "Registered!",
		"user_id":     user.ID,
		"token":       newToken,
	})
	fmt.Println(http.StatusCreated, "Registered!")
}

// GetUser 处理获取单个用户的API请求
func GetUser(c *gin.Context) {
	userIdString := c.DefaultQuery("user_id", "") // 从路径中获取用户ID
	// 验证 user_id
	userIdInt, err := strconv.Atoi(userIdString)
	if err != nil || userIdInt <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"status_code": 1,
			"status_msg":  "Invalid target user ID.",
			"user":        nil,
		})
		return
	}

	// 获取用户信息
	var user models.User
	db := c.MustGet("db").(*gorm.DB)
	if err := db.Preload("Profile").First(&user, userIdString).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) { // 目标用户不存在
			c.JSON(http.StatusNotFound, gin.H{
				"status_code": 1,
				"status_msg":  "User not found.",
				"user":        nil,
			})
			fmt.Println(http.StatusNotFound, "User not found.")
		} else {
			c.JSON(http.StatusNotFound, gin.H{
				"status_code": 1,
				"status_msg":  "User not found.",
				"user":        nil,
			})
			fmt.Println(http.StatusInternalServerError, "Database error.")
		}
		return
	}

	// 查看当前登录用户是否关注目标用户
	var relation models.Relation
	tokenString := c.DefaultQuery("token", "")
	currentUserId, _ := utils.ValidateToken(tokenString)
	result := db.Where(
		"from_user_id = ? AND to_user_id = ?", currentUserId, userIdString).First(&relation)
	isFollowed := result.RowsAffected > 0

	userResponse := map[string]interface{}{
		"id":               user.ID,
		"name":             user.Username,
		"follow_count":     user.Profile.FollowCount,
		"is_follow":        isFollowed,
		"avatar":           user.Profile.Avatar,
		"background_image": user.Profile.Background,
		"signature":        user.Profile.Signature,
		"total_favorited":  user.Profile.TotalFavorited,
		"work_count":       user.Profile.WorkCount,
		"favorite_count":   user.Profile.FavoriteCount,
	}

	fmt.Println(http.StatusOK, userResponse)
	c.JSON(http.StatusOK, gin.H{
		"status_code": 0,
		"status_msg":  "OK",
		"user":        userResponse,
	})
}

// Login 处理用户登录请求
func Login(c *gin.Context) {
	var user models.User
	var inputUser models.User

	// 从 PostForm 中提取用户名和密码
	inputUser.Username = c.Query("username")
	inputUser.Password = c.Query("password")

	// 验证用户名密码非空
	if inputUser.Username == "" || inputUser.Password == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status_code": 1,
			"status_msg":  "Username or Password is missing.",
		})
		fmt.Println(http.StatusUnauthorized, "Username or Password is missing.")
		return
	}

	// 使用GORM检索用户是否存在
	db := c.MustGet("db").(*gorm.DB)
	if err := db.Where("username = ?", inputUser.Username).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status_code": 1,
			"status_msg":  "User not found.",
		})
		fmt.Println(http.StatusUnauthorized, "User not found.")
		return
	}

	// 验证密码 (这里验证的是明文密码，实际生产中应该使用哈希后的密码)
	if user.Password != inputUser.Password {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status_code": 1,
			"status_msg":  "Incorrect password.",
		})
		fmt.Println(http.StatusUnauthorized, "Incorrect password.")
		return
	}

	// 生成新Token
	newToken, err := utils.GenerateToken(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status_code": 1,
			"status_msg":  "Failed to generate token.",
			"user_id":     user.ID,
		})
		fmt.Println(http.StatusInternalServerError, "Failed to generate token.")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status_code": 0,
		"status_msg":  "Logged in successfully.",
		"user_id":     user.ID,
		"token":       newToken,
	})
	fmt.Println(http.StatusOK, "Logged in successfully.")
}
