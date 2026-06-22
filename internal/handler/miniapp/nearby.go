package miniapp

import (
	"encoding/json"
	"fmt"
	"net/http"

	_ "travel-server/internal/model" // swagger 类型引用需要

	"github.com/gin-gonic/gin"

	"travel-server/internal/repository"
	"travel-server/pkg/config"
	"travel-server/pkg/response"
)

// GetNearby 根据经纬度获取周边景点/民宿推荐（调用高德POI搜索）
// @Summary 周边推荐
// @Tags 小程序-周边游
// @Param lat query string true "纬度"
// @Param lng query string true "经度"
// @Success 200 {object} response.Response
// @Router /api/v1/nearby [get]
func GetNearby(c *gin.Context) {
	lat := c.Query("lat")
	lng := c.Query("lng")
	if lat == "" || lng == "" {
		response.Fail(c, 400, "缺少经纬度参数")
		return
	}
	key := config.AppConfig.AmapKey
	url := fmt.Sprintf("https://restapi.amap.com/v3/place/around?key=%s&location=%s,%s&keywords=景区|民宿&radius=50000",
		key, lng, lat)
	resp, err := http.Get(url)
	if err != nil {
		response.Fail(c, 500, "周边搜索失败")
		return
	}
	defer resp.Body.Close()
	var data interface{}
	json.NewDecoder(resp.Body).Decode(&data)
	response.Success(c, data)
}

// GetTopRecommend 获取后台配置的“本周TOP推荐”
// @Summary 本周TOP推荐
// @Tags 小程序-周边游
// @Param city query string false "城市（可选）"
// @Success 200 {object} response.Response{data=[]model.Recommendation}
// @Router /api/v1/nearby/recommend [get]
func GetTopRecommend(c *gin.Context) {
	city := c.DefaultQuery("city", "")
	recs, err := repository.GetRecommendations(city)
	if err != nil {
		response.Fail(c, 500, "获取推荐失败")
		return
	}
	response.Success(c, recs)
}
