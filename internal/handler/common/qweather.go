package common

import (
	"context"
	"crypto/ed25519"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"
	"github.com/golang-jwt/jwt/v5"

	"travel-server/pkg/config"
	"travel-server/pkg/database"
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

// QWeatherDaily 和风7天预报每日数据
// @Description 每日天气预报字段
// @Description fxDate: 日期, tempMax: 最高温, tempMin: 最低温
// @Description textDay: 白天天气, textNight: 夜间天气, iconDay: 天气图标编号
// @Description windDirDay: 风向, humidity: 湿度百分比, precip: 降水量mm
// @Description vis: 能见度km, uvIndex: 紫外线指数
// @Description cloud: 云量（可能为空）
type QWeatherDaily struct {
	FxDate       string `json:"fxDate"`       // 日期
	Sunrise      string `json:"sunrise"`      // 日出时间
	Sunset       string `json:"sunset"`       // 日落时间
	TempMax      string `json:"tempMax"`      // 最高温度
	TempMin      string `json:"tempMin"`      // 最低温度
	IconDay      string `json:"iconDay"`      // 白天天气图标代码
	TextDay      string `json:"textDay"`      // 白天天气描述
	IconNight    string `json:"iconNight"`    // 夜间天气图标代码
	TextNight    string `json:"textNight"`    // 夜间天气描述
	WindDirDay   string `json:"windDirDay"`   // 白天风向
	WindScaleDay string `json:"windScaleDay"` // 白天风力等级
	Humidity     string `json:"humidity"`     // 相对湿度
	Precip       string `json:"precip"`       // 降水量
	Vis          string `json:"vis"`          // 能见度
	Cloud        string `json:"cloud"`        // 云量
	UvIndex      string `json:"uvIndex"`      // 紫外线指数
}

// QWeatherResponse 和风天气7天预报完整返回
// @Description code: 状态码200代表成功, updateTime: 接口更新时间
// @Description fxLink: 天气网页跳转链接, daily: 7天预报数组
type QWeatherResponse struct {
	Code       string          `json:"code"`       // 状态码
	UpdateTime string          `json:"updateTime"` // 最近更新时间
	FxLink     string          `json:"fxLink"`     // 天气网页链接
	Daily      []QWeatherDaily `json:"daily"`      // 7天预报列表
}

// GetQWeather 查询城市天气（和风天气 API）
// @Summary 获取城市天气（和风）
// @Tags 公共
// @Param city query string true "城市名称"
// @Success 200 {object} response.Response{data=QWeatherResponse}
// @Router /api/v1/weather/qweather [get]
func GetQWeather(c *gin.Context) {
	city := c.Query("city")
	if city == "" {
		response.Fail(c, http.StatusBadRequest, "缺少城市参数")
		return
	}

	// Redis 缓存键
	cacheKey := "weather:city:" + city
	ctx := context.Background()

	// 1. 尝试从 Redis 读取缓存
	cached, err := database.RedisClient.Get(ctx, cacheKey).Result()
	if err == nil && cached != "" {
		var cachedResult map[string]interface{}
		if json.Unmarshal([]byte(cached), &cachedResult) == nil {
			response.Success(c, cachedResult)
			return
		}
	}

	kid := config.AppConfig.QWeatherKID
	privateKeyPEM := config.AppConfig.QWeatherKey
	qweatherPID := config.AppConfig.QWeatherPID

	apiBaseURL := "https://p86yw35uvq.re.qweatherapi.com"

	// 2. 生成 JWT Token（两个接口通用）
	token, err := generateQWeatherJWT(kid, qweatherPID, privateKeyPEM)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "生成天气认证凭证失败: "+err.Error())
		return
	}

	// 创建带 JWT Header 的客户端
	client := resty.New().SetHeader("Authorization", "Bearer "+token)

	// 3. 第一步：调用 GeoAPI 获取城市的 Location ID
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

	// 4. 第二步：调用 7 天天气预报接口
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

	// 5. 写入 Redis 缓存，6 小时过期
	if data, err := json.Marshal(weatherResult); err == nil {
		database.RedisClient.Set(ctx, cacheKey, string(data), 6*time.Hour)
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
