package miniapp

import (
	"time"

	"github.com/gin-gonic/gin"

	"travel-server/internal/model"
	"travel-server/pkg/database"
	"travel-server/pkg/response"
)

// GetFootprints 获取用户的足迹（已点亮城市）
// @Summary 获取足迹
// @Security BearerAuth
// @Tags 小程序-足迹
// @Success 200 {object} response.Response{data=[]model.Footprint}
// @Router /api/v1/footprint [get]
func GetFootprints(c *gin.Context) {
	uid := c.MustGet("userID").(string)
	var footprints []model.Footprint
	if err := database.DB.Where("user_id = ?", uid).Find(&footprints).Error; err != nil {
		response.Fail(c, 500, "获取足迹失败")
		return
	}
	response.Success(c, footprints)
}

// SyncFootprint 同步/点亮一个新城市
// @Summary 同步足迹
// @Security BearerAuth
// @Tags 小程序-足迹
// @Param body body object{city=string,province=string,lat=float64,lng=float64} true "城市信息"
// @Success 200 {object} response.Response
// @Router /api/v1/footprint/sync [post]
func SyncFootprint(c *gin.Context) {
	var req struct {
		City     string  `json:"city"`
		Province string  `json:"province"`
		Lat      float64 `json:"lat"`
		Lng      float64 `json:"lng"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 400, "参数错误")
		return
	}
	fp := model.Footprint{
		UserID:    c.MustGet("userID").(string),
		City:      req.City,
		Province:  req.Province,
		Lat:       req.Lat,
		Lng:       req.Lng,
		VisitedAt: time.Now(),
	}
	if err := database.DB.Create(&fp).Error; err != nil {
		response.Fail(c, 500, "同步失败")
		return
	}
	response.Success(c, nil)
}

// GeneratePoster 生成足迹海报（模拟）
// @Summary 生成足迹海报
// @Security BearerAuth
// @Tags 小程序-足迹
// @Success 200 {object} response.Response{data=object{poster_url=string}}
// @Router /api/v1/footprint/poster [get]
func GeneratePoster(c *gin.Context) {
	// TODO: 实际生成海报逻辑（调用图片合成服务），这里返回模拟链接
	response.Success(c, gin.H{"posterUrl": "https://cdn.example.com/poster/user123.png"})
}
