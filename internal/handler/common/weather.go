package common

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"

	"travel-server/pkg/config"
	"travel-server/pkg/response"
)

type amapGeocodeResponse struct {
	Status   string `json:"status"`
	Info     string `json:"info"`
	Infocode string `json:"infocode"`
	Geocodes []struct {
		Adcode string `json:"adcode"`
	} `json:"geocodes"`
}

type amapLiveWeather struct {
	City          string `json:"city"`
	Weather       string `json:"weather"`
	Temperature   string `json:"temperature"`
	Winddirection string `json:"winddirection"`
	Windpower     string `json:"windpower"`
	Humidity      string `json:"humidity"`
	Reporttime    string `json:"reporttime"`
}

type amapForecastDay struct {
	Date         string `json:"date"`
	Dayweather   string `json:"dayweather"`
	Nightweather string `json:"nightweather"`
	Daytemp      string `json:"daytemp"`
	Nighttemp    string `json:"nighttemp"`
	Daywind      string `json:"daywind"`
	Nightwind    string `json:"nightwind"`
	Daypower     string `json:"daypower"`
	Nightpower   string `json:"nightpower"`
}

type amapForecastWeather struct {
	City  string            `json:"city"`
	Casts []amapForecastDay `json:"casts"`
}

type amapWeatherResponse struct {
	Status    string                `json:"status"`
	Info      string                `json:"info"`
	Infocode  string                `json:"infocode"`
	Lives     []amapLiveWeather     `json:"lives"`
	Forecasts []amapForecastWeather `json:"forecasts"`
}

func buildWeatherPayload(city string, days int, weatherData *amapWeatherResponse) map[string]interface{} {
	payload := map[string]interface{}{
		"city": city,
		"days": days,
	}
	if weatherData == nil {
		return payload
	}

	if len(weatherData.Lives) > 0 {
		live := weatherData.Lives[0]
		payload["current"] = map[string]interface{}{
			"weather":       live.Weather,
			"temperature":   live.Temperature,
			"winddirection": live.Winddirection,
			"windpower":     live.Windpower,
			"humidity":      live.Humidity,
			"reporttime":    live.Reporttime,
		}
	}

	forecastDays := make([]map[string]interface{}, 0, days)
	if len(weatherData.Forecasts) > 0 {
		casts := weatherData.Forecasts[0].Casts
		if days > len(casts) {
			days = len(casts)
		}
		for _, cast := range casts[:days] {
			forecastDays = append(forecastDays, map[string]interface{}{
				"date":         cast.Date,
				"dayweather":   cast.Dayweather,
				"nightweather": cast.Nightweather,
				"daytemp":      cast.Daytemp,
				"nighttemp":    cast.Nighttemp,
				"daywind":      cast.Daywind,
				"nightwind":    cast.Nightwind,
				"daypower":     cast.Daypower,
				"nightpower":   cast.Nightpower,
			})
		}
	}
	payload["forecast"] = forecastDays
	return payload
}

// resolveCityToAdcode 使用高德地理编码服务把城市名解析为 adcode（区划代码）
// 如果解析失败，返回空字符串和错误，调用方可决定回退到原始 city
func resolveCityToAdcode(city, key string) (string, error) {
	if city == "" || key == "" {
		return "", nil
	}
	apiURL, err := url.Parse("https://restapi.amap.com/v3/geocode/geo")
	if err != nil {
		return "", err
	}
	q := apiURL.Query()
	q.Set("key", key)
	q.Set("address", city)
	apiURL.RawQuery = q.Encode()

	resp, err := http.Get(apiURL.String())
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	var geo amapGeocodeResponse
	if err := json.NewDecoder(resp.Body).Decode(&geo); err != nil {
		return "", err
	}
	if geo.Status != "1" || len(geo.Geocodes) == 0 {
		return "", nil
	}
	return strings.TrimSpace(geo.Geocodes[0].Adcode), nil
}

// GetWeather 查询城市天气（高德 API）
// @Summary 获取城市天气
// @Tags 公共
// @Param city query string true "城市名称"
// @Success 200 {object} response.Response
// @Router /api/v1/weather [get]
func GetWeather(c *gin.Context) {
	city := strings.TrimSpace(c.Query("city"))
	if city == "" {
		response.Fail(c, http.StatusBadRequest, "缺少city参数")
		return
	}

	key := strings.TrimSpace(config.AppConfig.AmapKey)
	if key == "" {
		response.Fail(c, http.StatusInternalServerError, "天气服务未配置")
		return
	}

	// 尝试把城市名解析为高德的 adcode，以便获取更完整的预报
	cityParam := city
	if adcode, err := resolveCityToAdcode(city, config.AppConfig.AmapKey); err == nil && adcode != "" {
		cityParam = adcode
	}

	apiURL, err := url.Parse("https://restapi.amap.com/v3/weather/weatherInfo")
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "天气查询失败")
		return
	}
	query := apiURL.Query()
	query.Set("key", key)
	query.Set("city", cityParam)
	query.Set("extensions", "all")
	apiURL.RawQuery = query.Encode()

	resp, err := http.Get(apiURL.String())
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "天气查询失败")
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		response.Fail(c, http.StatusBadGateway, "天气查询失败")
		return
	}

	var weatherData amapWeatherResponse
	if err := json.NewDecoder(resp.Body).Decode(&weatherData); err != nil {
		response.Fail(c, http.StatusInternalServerError, "天气查询失败")
		return
	}
	if weatherData.Status != "1" {
		response.Fail(c, http.StatusBadGateway, weatherData.Info)
		return
	}

	payload := buildWeatherPayload(city, 7, &weatherData)
	response.Success(c, payload)
}
