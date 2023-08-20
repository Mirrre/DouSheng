package middleware

import (
    "app/utils"
  	"github.com/gin-gonic/gin"
  	"net/http"
)

func Authentication() gin.HandlerFunc {
    return func(c *gin.Context) {
        // tokenString := c.GetHeader("Authorization")
        tokenString := c.Query("token")

        userID, err := utils.ValidateToken(tokenString)
        if err != nil {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
            c.Abort()  // 验证不通过，阻止API处理函数继续执行
            return
        }

    c.Set("userIDFromToken", userID)

    c.Next()  // 验证通过，继续处理API
    }
}