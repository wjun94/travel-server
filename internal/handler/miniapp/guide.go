package miniapp

import (
	"strconv"

	"travel-server/internal/model"
	"travel-server/internal/repository"
	"travel-server/pkg/response"

	"github.com/gin-gonic/gin"
)

// GetFeed 获取公开的攻略瀑布流
// @Summary 攻略瀑布流
// @Tags 小程序-攻略
// @Param page query int false "页码"
// @Param pageSize query int false "每页数量"
// @Param destination query string false "目的地筛选"
// @Success 200 {object} response.Response{data=object{list=[]model.Guide,total=int}}
// @Router /api/v1/feed [get]
func GetFeed(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	destination := c.Query("destination")
	guides, total, err := repository.GetFeedGuides(page, pageSize, destination)
	if err != nil {
		response.Fail(c, 500, "获取失败")
		return
	}
	response.Success(c, gin.H{"list": guides, "total": total})
}

// CreateGuide 发布一篇攻略
// @Summary 发布攻略
// @Security BearerAuth
// @Tags 小程序-攻略
// @Param body body model.Guide true "攻略内容"
// @Success 200 {object} response.Response
// @Router /api/v1/guide [post]
func CreateGuide(c *gin.Context) {
	var guide model.Guide
	if err := c.ShouldBindJSON(&guide); err != nil {
		response.Fail(c, 400, "参数错误")
		return
	}
	guide.UserID = c.GetUint("userID")
	guide.Status = 0 // 默认草稿
	if err := repository.CreateGuide(&guide); err != nil {
		response.Fail(c, 500, "发布失败")
		return
	}
	response.Success(c, guide)
}

// GetGuideDetail 获取攻略详情（含板块）
// @Summary 攻略详情
// @Tags 小程序-攻略
// @Param id path int true "攻略ID"
// @Success 200 {object} response.Response{data=object{guide=model.Guide,sections=[]model.GuideSection}}
// @Router /api/v1/guide/{id} [get]
func GetGuideDetail(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.Fail(c, 400, "参数错误")
		return
	}
	guide, err := repository.GetGuideByID(uint(id))
	if err != nil {
		response.Fail(c, 404, "攻略不存在")
		return
	}
	// 同时获取板块
	sections, _ := repository.GetSectionsByGuideID(uint(id))
	// 增加浏览量（非作者）
	userID := c.GetUint("userID")
	if guide.UserID != userID {
		_ = repository.IncrementGuideViewCount(uint(id))
	}
	response.Success(c, gin.H{"guide": guide, "sections": sections})
}

// CreateSection 添加攻略板块
// @Summary 添加攻略板块
// @Security BearerAuth
// @Tags 小程序-攻略
// @Param body body model.GuideSection true "板块内容"
// @Success 200 {object} response.Response
// @Router /api/v1/guide/section [post]
func CreateSection(c *gin.Context) {
	var section model.GuideSection
	if err := c.ShouldBindJSON(&section); err != nil {
		response.Fail(c, 400, "参数错误")
		return
	}
	if err := repository.CreateSection(&section); err != nil {
		response.Fail(c, 500, "添加失败")
		return
	}
	response.Success(c, section)
}
