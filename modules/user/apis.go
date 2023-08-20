package user

import (
  "app/modules/models"
  "app/utils"
	"errors"
	"net/http"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
  "strconv"
)

// Register 处理用户注册的API请求
func Register(c *gin.Context) {
  	var user models.User
  
    user.Username = c.PostForm("username")
    user.Password = c.PostForm("password")
    if len(user.Username) < 6 || len(user.Password) < 6 || len(user.Username) > 25 || len(user.Password) > 25 {
        c.JSON(http.StatusBadRequest, gin.H{
          "status_code": 1,
          "status_msg": "Username and Password should be non-empty and between 6 - 25 characters.",
          "user_id": nil,
          "username": user.Username,
        })
        return
    }
  
  	// 使用GORM将用户数据存储到数据库中
  	db := c.MustGet("db").(*gorm.DB)
  	if err := db.Create(&user).Error; err != nil {
  		c.JSON(http.StatusBadRequest, gin.H{
          "status_code": 1,
          "status_msg": "Failed to register",
          "user_id": nil,
        })
  		return
  	}

    // 生成新Token
    newToken, err := utils.GenerateToken(user.ID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "status_code": 1,
            "status_msg": "Failed to generate token",
            "user_id": user.ID,
        })
        return
    }
  
  	c.JSON(http.StatusCreated, gin.H{
      "status_code": 0,
      "status_msg": "Registered!",
      "user_id": user.ID,
      "token": newToken,
    })
}

// GetUser 处理获取单个用户的API请求
func GetUser(c *gin.Context) {
    id := c.Query("id")  // 从路径中获取用户ID

    // 将ID从字符串转换为uint
    userID, err := strconv.ParseUint(id, 10, 32)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
        return
    }

    var user models.User
    db := c.MustGet("db").(*gorm.DB)
    if err := db.First(&user, userID).Error; err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
        } else {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
        }
        return
    }

    c.JSON(http.StatusOK, gin.H{
      "status_code": 0,
      "status_msg": "OK",
      "user": user,
    })
}

// Login 处理用户登录请求
func Login(c *gin.Context) {
    var user models.User
    var inputUser models.User

    // 从 PostForm 中提取用户名和密码
    inputUser.Username = c.PostForm("username")
    inputUser.Password = c.PostForm("password")

    // 验证用户名密码非空
    if inputUser.Username == "" || inputUser.Password == "" {
        c.JSON(http.StatusUnauthorized, gin.H{
            "status_code": 1,
            "status_msg": "Username of Password is missing.",
        })
        return
    }

    // 使用GORM检索用户是否存在
    db := c.MustGet("db").(*gorm.DB)
    if err := db.Where("username = ?", inputUser.Username).First(&user).Error; err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{
            "status_code": 1,
            "status_msg": "User not found",
        })
        return
    }

    // 验证密码 (这里验证的是明文密码，实际生产中应该使用哈希后的密码)
    if user.Password != inputUser.Password {
        c.JSON(http.StatusUnauthorized, gin.H{
            "status_code": 1,
            "status_msg": "Incorrect password",
        })
        return
    }

    // 生成新Token
    newToken, err := utils.GenerateToken(user.ID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "status_code": 1,
            "status_msg": "Failed to generate token",
            "user_id": user.ID,
        })
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "status_code": 0,
        "status_msg": "Logged in successfully",
        "user_id": user.ID,
        "token": newToken,
    })
}

