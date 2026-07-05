package common

import (
	"crypto/ed25519"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2" // 推荐使用 resty 方便处理 HTTP 请求，也可以用原生 net/http
	"github.com/golang-jwt/jwt/v5"

	"travel-server/pkg/config"
	"travel-server/pkg/response"
)

// QWeatherGeoResponse 城市搜索返回结构
type QWeatherGeoResponse struct {
	Code     string `json:"code"`
	Location []struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"location"`
}

// GetQWeather 查询城市天气（和风天气 API）
// @Summary 获取城市天气（和风）
// @Tags 公共
// @Param city query string true "城市名称"
// @Success 200 {object} response.Response
// @Router /api/v1/weather/qweather [get]
func GetQWeather(c *gin.Context) {
	city := c.Query("city")
	if city == "" {
		response.Fail(c, http.StatusBadRequest, "缺少城市参数")
		return
	}

	kid := config.AppConfig.QWeatherKID
	privateKeyPEM := config.AppConfig.QWeatherKey
	qweatherPID := config.AppConfig.QWeatherPID

	// 【修改点 1】统一使用你的专属 Host
	apiBaseURL := "https://p86yw35uvq.re.qweatherapi.com"

	// 1. 生成 JWT Token（两个接口通用）
	token, err := generateQWeatherJWT(kid, qweatherPID, privateKeyPEM)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "生成天气认证凭证失败: "+err.Error())
		return
	}

	// 创建带 JWT Header 的客户端
	client := resty.New().SetHeader("Authorization", "Bearer "+token)

	// 2. 第一步：调用 GeoAPI 获取城市的 Location ID
	// 【修改点 2】严格按照文档拼接专属 Host 和 /geo/v2/city/lookup 路径
	geoURL := fmt.Sprintf("%s/geo/v2/city/lookup", apiBaseURL)
	var geoResult QWeatherGeoResponse

	respGeo, err := client.R().
		SetQueryParam("location", city).
		SetResult(&geoResult).
		Get(geoURL)

	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "请求城市搜索接口失败: "+err.Error())
		return
	}

	// 如果返回的 Code 不是 200，打印出原始 Body 方便排查
	if geoResult.Code != "200" || len(geoResult.Location) == 0 {
		if geoResult.Code == "" {
			fmt.Println("GeoAPI 原始返回内容:", respGeo.String())
			response.Fail(c, http.StatusInternalServerError, fmt.Sprintf("城市搜索失败，HTTP状态码: %d", respGeo.StatusCode()))
			return
		}
		response.Fail(c, http.StatusInternalServerError, "城市搜索失败，错误码: "+geoResult.Code)
		return
	}
	locationID := geoResult.Location[0].ID

	// 3. 第二步：调用 7 天天气预报接口
	weatherURL := fmt.Sprintf("%s/v7/weather/7d", apiBaseURL)
	var weatherResult map[string]interface{}

	respWeather, err := client.R().
		SetQueryParam("location", locationID).
		SetResult(&weatherResult).
		Get(weatherURL)

	if err != nil || respWeather.StatusCode() != http.StatusOK {
		response.Fail(c, http.StatusInternalServerError, "获取天气预报失败")
		return
	}

	response.Success(c, weatherResult)
}

// generateQWeatherJWT 解析 Ed25519 密钥并生成和风规范的 JWT
func generateQWeatherJWT(kid, sub, keyPEM string) (string, error) {
	// 解析 PEM 格式的私钥
	block, _ := pem.Decode([]byte(keyPEM))
	if block == nil {
		return "", errors.New("failed to parse PEM block containing the private key")
	}

	// 转换为 ASN.1 DER 编码的私钥结构
	privKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return "", fmt.Errorf("failed to parse PKCS8 private key: %v", err)
	}

	// 确保是 Ed25519 私钥
	ed25519Key, ok := privKey.(ed25519.PrivateKey)
	if !ok {
		return "", errors.New("private key is not of type Ed25519")
	}

	// 组装 Payload 载荷
	now := time.Now().Unix()
	claims := jwt.MapClaims{
		"sub": sub,
		"iat": now - 30,  // 容错提前 30 秒
		"exp": now + 600, // 10 分钟有效期
	}

	// 创建 Token 对象
	token := jwt.NewWithClaims(jwt.SigningMethodEdDSA, claims)

	// 显式在 Header 中注入 kid
	token.Header["kid"] = kid

	// 使用私钥签名
	tokenString, err := token.SignedString(ed25519Key)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}
