package middleware

import (
	"app/util"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
)

func Authentication() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := c.Query("token")
		if tokenString == "" {
			tokenString = c.PostForm("token")
		}

		userID, err := util.ValidateToken(tokenString)
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
