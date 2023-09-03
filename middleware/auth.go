package middleware

import (
	"app/utils"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
)

func Authentication() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := c.DefaultQuery("token", "")
		if tokenString == "" {
			tokenString = c.DefaultPostForm("token", "")
		}

		userID, err := utils.ValidateToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			fmt.Println(http.StatusUnauthorized, "Invalid token")
			c.Abort() // 验证不通过，阻止API处理函数继续执行
			return
		}

		c.Set("userIDFromToken", userID)

		c.Next() // 验证通过，继续处理API
	}
}
