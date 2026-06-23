package middleware

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	"travel-server/internal/repository"
)

// 小程序用户 JWT 密钥
var miniAppJwtSecret = []byte("miniapp-secret-key")

// 后台管理员 JWT 密钥（独立）
var adminJwtSecret = []byte("admin-secret-key")

// ---------- 小程序用户 Claims ----------
type MiniAppClaims struct {
	UserID uint `json:"user_id"`
	jwt.RegisteredClaims
}

func GenerateMiniAppToken(userID uint) (string, error) {
	claims := MiniAppClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)),
			Issuer:    "travel-miniapp",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(miniAppJwtSecret)
}

func ParseMiniAppToken(tokenString string) (*MiniAppClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &MiniAppClaims{}, func(token *jwt.Token) (interface{}, error) {
		return miniAppJwtSecret, nil
	})
	if claims, ok := token.Claims.(*MiniAppClaims); ok && token.Valid {
		return claims, nil
	}
	return nil, err
}

// JWTAuth 小程序用户认证中间件
func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"msg": "未登录"})
			return
		}
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		claims, err := ParseMiniAppToken(tokenString)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"msg": "token无效"})
			return
		}
		c.Set("userID", claims.UserID)
		c.Next()
	}
}

// ---------- 后台管理员 Claims ----------
type AdminClaims struct {
	AdminUserID uint `json:"admin_user_id"`
	jwt.RegisteredClaims
}

func GenerateAdminToken(adminUserID uint) (string, error) {
	claims := AdminClaims{
		AdminUserID: adminUserID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)),
			Issuer:    "travel-admin",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(adminJwtSecret)
}

func ParseAdminToken(tokenString string) (*AdminClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &AdminClaims{}, func(token *jwt.Token) (interface{}, error) {
		return adminJwtSecret, nil
	})
	if claims, ok := token.Claims.(*AdminClaims); ok && token.Valid {
		return claims, nil
	}
	return nil, err
}

// AdminOnly 后台管理员认证中间件（验证 Admin JWT，并注入 adminUserID 和角色）
func AdminOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"msg": "未登录"})
			return
		}
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		claims, err := ParseAdminToken(tokenString)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"msg": "token无效"})
			return
		}

		// 查询后台用户是否存在且启用
		adminUser, err := repository.GetAdminUserByID(claims.AdminUserID)
		if err != nil || adminUser.Status != 1 {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"msg": "用户已被禁用或不存在"})
			return
		}

		// 注入上下文
		c.Set("adminUserID", adminUser.ID)
		c.Set("adminRole", adminUser.Role)
		c.Next()
	}
}
