package middleware

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	"travel-server/internal/model"
	"travel-server/pkg/database"
)

// jwtSecret JWT 签名密钥（生产环境应从配置读取）
var jwtSecret = []byte("travel-secret-key")

// Claims 自定义 JWT 载荷
type Claims struct {
	UserID uint `json:"user_id"`
	jwt.RegisteredClaims
}

// GenerateToken 为用户生成 JWT Token
func GenerateToken(userID uint) (string, error) {
	claims := Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)), // 7天过期
			Issuer:    "travel-server",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

// ParseToken 解析并验证 JWT Token
func ParseToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})
	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}
	return nil, err
}

// JWTAuth 验证 JWT Token 并注入 userID 到上下文
func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"msg": "未登录"})
			return
		}
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		claims, err := ParseToken(tokenString)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"msg": "token无效"})
			return
		}
		c.Set("userID", claims.UserID)
		c.Next()
	}
}

// AdminOnly 检查用户是否为管理员（role >= 2）
func AdminOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		uid, exists := c.Get("userID")
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"msg": "未登录"})
			return
		}
		var user model.User
		if err := database.DB.First(&user, uid).Error; err != nil || user.Role < 2 {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"msg": "无权限"})
			return
		}
		c.Next()
	}
}
