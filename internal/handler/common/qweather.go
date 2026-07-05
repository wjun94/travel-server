package common

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"

	"travel-server/pkg/config"
	"travel-server/pkg/response"
)

type qWeatherDaily struct {
	FxDate       string `json:"fxDate"`
	TextDay      string `json:"textDay"`
	TextNight    string `json:"textNight"`
	TempMax      string `json:"tempMax"`
	TempMin      string `json:"tempMin"`
	WindDirDay   string `json:"windDirDay"`
	WindScaleDay string `json:"windScaleDay"`
}

type qWeatherResponse struct {
	Code       string          `json:"code"`
	UpdateTime string          `json:"updateTime"`
	FxLink     string          `json:"fxLink"`
	Daily      []qWeatherDaily `json:"daily"`
}

type qWeatherCityItem struct {
	Name string `json:"name"`
	ID   string `json:"id"`
	Lat  string `json:"lat"`
	Lon  string `json:"lon"`
	Rank string `json:"rank"`
}

type qWeatherCityResponse struct {
	Code     string             `json:"code"`
	Location []qWeatherCityItem `json:"location"`
}

var commonCityCoords = map[string]string{
	"北京": "116.407396,39.904211",
	"上海": "121.473701,31.230416",
	"广州": "113.264385,23.129110",
	"深圳": "114.057868,22.543099",
	"杭州": "120.155070,30.274084",
	"成都": "104.066541,30.572269",
	"武汉": "114.305392,30.593099",
	"西安": "108.940174,34.261117",
	"重庆": "106.551556,29.563010",
	"南京": "118.796877,32.060255",
	"苏州": "120.584298,31.297833",
	"天津": "117.200983,39.084158",
	"长沙": "112.938814,28.228209",
	"郑州": "113.625368,34.746599",
	"青岛": "120.382639,36.067082",
	"沈阳": "123.431474,41.805698",
	"昆明": "102.832891,24.880095",
	"厦门": "118.089425,24.479833",
	"三亚": "109.508268,18.252847",
	"丽江": "100.229628,26.874098",
	"拉萨": "91.140856,29.645554",
}

func resolveCityByQWeather(city, key string) (string, error) {
	if city == "" || key == "" {
		return "", nil
	}
	apiURL, _ := url.Parse("https://geoapi.qweather.com/v2/city/lookup")
	q := apiURL.Query()
	q.Set("location", city)
	q.Set("key", key)
	q.Set("range", "cn")
	apiURL.RawQuery = q.Encode()

	resp, err := http.Get(apiURL.String())
	if err != nil {
		return "", fmt.Errorf("和风城市搜索请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("和风城市搜索返回非 200: %d", resp.StatusCode)
	}

	var cr qWeatherCityResponse
	if err := json.NewDecoder(resp.Body).Decode(&cr); err != nil {
		return "", fmt.Errorf("和风城市搜索响应解析失败: %w", err)
	}
	if cr.Code != "200" || len(cr.Location) == 0 {
		return "", fmt.Errorf("和风城市搜索未找到 %q", city)
	}
	// 取 rank 最小的结果
	best := cr.Location[0]
	for _, loc := range cr.Location[1:] {
		if loc.Rank < best.Rank {
			best = loc
		}
	}
	log.Printf("[QWeather] 城市搜索: %s → %s id=%s (%s,%s)", city, best.Name, best.ID, best.Lon, best.Lat)
	return best.ID, nil
}

func buildQWeatherPayload(city string, days int, data *qWeatherResponse) map[string]interface{} {
	p := map[string]interface{}{"city": city, "days": days}
	if data == nil {
		return p
	}
	if len(data.Daily) > 0 {
		first := data.Daily[0]
		p["current"] = map[string]interface{}{
			"weather":     first.TextDay,
			"temperature": first.TempMax,
			"reporttime":  data.UpdateTime,
		}
	}
	forecast := make([]map[string]interface{}, 0, days)
	limit := len(data.Daily)
	if days < limit {
		limit = days
	}
	for _, d := range data.Daily[:limit] {
		forecast = append(forecast, map[string]interface{}{
			"date":         d.FxDate,
			"dayweather":   d.TextDay,
			"nightweather": d.TextNight,
			"daytemp":      d.TempMax,
			"nighttemp":    d.TempMin,
			"daywind":      d.WindDirDay,
			"daypower":     d.WindScaleDay,
		})
	}
	p["forecast"] = forecast
	return p
}

// GetQWeather 查询城市天气（和风天气 API）
// @Summary 获取城市天气（和风）
// @Tags 公共
// @Param city query string true "城市名称"
// @Success 200 {object} response.Response
// @Router /api/v1/weather/qweather [get]
func GetQWeather(c *gin.Context) {
	city := strings.TrimSpace(c.Query("city"))
	if city == "" {
		response.Fail(c, http.StatusBadRequest, "缺少city参数")
		return
	}

	key := strings.TrimSpace(config.AppConfig.QWeatherKey)
	if key == "" {
		response.Fail(c, http.StatusInternalServerError, "和风天气服务未配置")
		return
	}

	// 用和风城市搜索把中文名转为 location ID
	location, err := resolveCityByQWeather(city, key)
	if err != nil {
		log.Printf("[QWeather] 城市搜索失败: %v", err)
		// 兜底：内置坐标表
		if coords, ok := commonCityCoords[city]; ok {
			location = coords
		} else {
			response.Fail(c, http.StatusBadGateway, "天气查询失败")
			return
		}
	}

	apiURL, _ := url.Parse("https://api.qweather.com/v7/weather/7d")
	q := apiURL.Query()
	q.Set("key", key)
	q.Set("location", location)
	apiURL.RawQuery = q.Encode()

	resp, err := http.Get(apiURL.String())
	if err != nil {
		log.Printf("[QWeather] HTTP 请求失败: %v", err)
		response.Fail(c, http.StatusInternalServerError, "天气查询失败")
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		errBody := make([]byte, 1024)
		n, _ := resp.Body.Read(errBody)
		log.Printf("[QWeather] API 异常: status=%d body=%s", resp.StatusCode, string(errBody[:n]))
		response.Fail(c, http.StatusBadGateway, "天气查询失败")
		return
	}

	var weatherData qWeatherResponse
	if err := json.NewDecoder(resp.Body).Decode(&weatherData); err != nil {
		log.Printf("[QWeather] JSON 解析失败: %v", err)
		response.Fail(c, http.StatusInternalServerError, "天气查询失败")
		return
	}
	if weatherData.Code != "200" {
		log.Printf("[QWeather] 业务错误: code=%s", weatherData.Code)
		response.Fail(c, http.StatusBadGateway, "和风天气查询失败")
		return
	}

	payload := buildQWeatherPayload(city, 7, &weatherData)
	response.Success(c, payload)
}
