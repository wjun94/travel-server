package miniapp

import (
	"strconv"
	"strings"

	"travel-server/internal/middleware"
	"travel-server/internal/model"
	"travel-server/internal/repository"
	"travel-server/pkg/response"

	"github.com/gin-gonic/gin"
)

// GetGuideFeed 攻略瀑布流（公开接口，登录后可标记当前用户是否点赞）
// @Summary 攻略瀑布流
// @Tags 小程序-攻略
// @Param page query int false "页码"
// @Param pageSize query int false "每页数量"
// @Param destination query string false "目的地筛选"
// @Success 200 {object} response.Response{data=object{list=[]model.GuideFeedItem,total=int}}
// @Router /api/v1/guides [get]
func GetGuideFeed(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	destination := c.Query("destination")

	// 可选登录：如果带有效 token 则获取 userID 用于标记 isLiked
	userID := ""
	authHeader := c.GetHeader("Authorization")
	if authHeader != "" {
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if claims, err := middleware.ParseMiniAppToken(tokenString); err == nil {
			userID = claims.UserID
		}
	}

	guides, total, err := repository.GetGuideFeed(page, pageSize, destination, userID)
	if err != nil {
		response.Fail(c, 500, "获取失败")
		return
	}
	response.Success(c, gin.H{"list": guides, "total": total})
}

// CreateGuide 创建攻略（含板块）
// @Summary 创建攻略
// @Security BearerAuth
// @Tags 小程序-攻略
// @Param body body model.CreateGuideReq true "攻略内容（含板块列表）"
// @Success 200 {object} response.Response
// @Router /api/v1/guide [post]
func CreateGuide(c *gin.Context) {
	var req model.CreateGuideReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 400, "参数格式错误")
		return
	}
	// 业务校验
	if ve := model.ValidateCreateGuideReq(&req); ve != nil {
		response.Fail(c, 400, ve.Msg)
		return
	}
	// 组装 Guide
	guide := model.Guide{
		UserID:          c.MustGet("userID").(string),
		Title:           req.Title,
		CoverImage:      req.CoverImage,
		Destination:     req.Destination,
		Summary:         req.Summary,
		BudgetMin:       req.BudgetMin,
		BudgetMax:       req.BudgetMax,
		BestSeason:      req.BestSeason,
		RecommendedDays: req.RecommendedDays,
		Tags:            req.Tags,
		Difficulty:      req.Difficulty,
		CrowdType:       req.CrowdType,
		VideoURL:        req.VideoURL,
		Images:          req.Images,
		IsOriginal:      req.IsOriginal,
		Status:          req.Status,
	}
	// 组装 Sections
	sections := make([]model.GuideSection, len(req.Sections))
	for i, s := range req.Sections {
		sections[i] = model.GuideSection{
			SectionType: s.SectionType,
			Title:       s.Title,
			Content:     s.Content,
		}
	}
	// 事务写入
	if err := repository.CreateGuideWithSections(&guide, sections); err != nil {
		response.Fail(c, 500, "创建失败")
		return
	}
	response.Success(c, gin.H{"id": guide.ID})
}

// GetGuideDetail 获取攻略详情（含板块）
// @Summary 攻略详情
// @Tags 小程序-攻略
// @Param id path string true "攻略ID"
// @Success 200 {object} response.Response{data=object{guide=model.Guide,sections=[]model.GuideSection}}
// @Router /api/v1/guide/{id} [get]
func GetGuideDetail(c *gin.Context) {
	id := c.Param("id")
	guide, err := repository.GetGuideByID(id)
	if err != nil {
		response.Fail(c, 404, "攻略不存在")
		return
	}
	sections, _ := repository.GetSectionsByGuideID(id)
	// 增加浏览量（非作者）
	if userID := c.MustGet("userID").(string); guide.UserID != userID {
		_ = repository.IncrementGuideViewCount(id)
	}
	response.Success(c, gin.H{"guide": guide, "sections": sections})
}

// UpdateGuide 更新攻略基本信息
// @Summary 更新攻略
// @Security BearerAuth
// @Tags 小程序-攻略
// @Param id path string true "攻略ID"
// @Param body body model.UpdateGuideReq true "更新数据"
// @Success 200 {object} response.Response
// @Router /api/v1/guide/{id} [put]
func UpdateGuide(c *gin.Context) {
	id := c.Param("id")
	guide, err := repository.GetGuideByID(id)
	if err != nil {
		response.Fail(c, 404, "攻略不存在")
		return
	}
	if guide.UserID != c.MustGet("userID").(string) {
		response.Fail(c, 403, "无权操作")
		return
	}
	var req model.UpdateGuideReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 400, "参数错误")
		return
	}
	updates := buildUpdateMap(&req)
	if len(updates) == 0 {
		response.Fail(c, 400, "无更新内容")
		return
	}
	if err := repository.UpdateGuide(id, updates); err != nil {
		response.Fail(c, 500, "更新失败")
		return
	}
	response.Success(c, nil)
}

func buildUpdateMap(req *model.UpdateGuideReq) map[string]interface{} {
	m := make(map[string]interface{})
	if req.Title != nil {
		m["title"] = *req.Title
	}
	if req.CoverImage != nil {
		m["cover_image"] = *req.CoverImage
	}
	if req.Destination != nil {
		m["destination"] = *req.Destination
	}
	if req.Summary != nil {
		m["summary"] = *req.Summary
	}
	if req.BudgetMin != nil {
		m["budget_min"] = *req.BudgetMin
	}
	if req.BudgetMax != nil {
		m["budget_max"] = *req.BudgetMax
	}
	if req.BestSeason != nil {
		m["best_season"] = *req.BestSeason
	}
	if req.RecommendedDays != nil {
		m["recommended_days"] = *req.RecommendedDays
	}
	if req.Tags != nil {
		m["tags"] = *req.Tags
	}
	if req.Difficulty != nil {
		m["difficulty"] = *req.Difficulty
	}
	if req.CrowdType != nil {
		m["crowd_type"] = *req.CrowdType
	}
	if req.VideoURL != nil {
		m["video_url"] = *req.VideoURL
	}
	if req.Images != nil {
		m["images"] = *req.Images
	}
	if req.IsOriginal != nil {
		m["is_original"] = *req.IsOriginal
	}
	if req.Status != nil {
		m["status"] = *req.Status
	}
	return m
}

// ==================== 板块管理 ====================

// CreateSection 添加攻略板块
// @Summary 添加攻略板块
// @Security BearerAuth
// @Tags 小程序-攻略
// @Param body body model.CreateSectionReq true "板块内容"
// @Success 200 {object} response.Response
// @Router /api/v1/guide/section [post]
func CreateSection(c *gin.Context) {
	var req model.CreateSectionReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 400, "参数错误")
		return
	}
	if !model.ValidSectionTypes[req.SectionType] {
		response.Fail(c, 400, "无效的板块类型")
		return
	}
	// 从原始 body 提取 guideId
	var body struct {
		GuideID string `json:"guideId"`
	}
	c.ShouldBindJSON(&body)

	section := model.GuideSection{
		GuideID:     body.GuideID,
		SectionType: req.SectionType,
		Title:       req.Title,
		Content:     req.Content,
	}
	if err := repository.CreateSection(&section); err != nil {
		response.Fail(c, 500, "添加失败")
		return
	}
	response.Success(c, section)
}

// UpdateSection 更新攻略板块
// @Summary 更新攻略板块
// @Security BearerAuth
// @Tags 小程序-攻略
// @Param id path string true "板块ID"
// @Param body body object true "更新数据"
// @Success 200 {object} response.Response
// @Router /api/v1/guide/section/{id} [put]
func UpdateSection(c *gin.Context) {
	id := c.Param("id")
	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		response.Fail(c, 400, "参数错误")
		return
	}
	if st, ok := updates["sectionType"].(string); ok {
		if !model.ValidSectionTypes[st] {
			response.Fail(c, 400, "无效的板块类型")
			return
		}
	}
	if err := repository.UpdateSection(id, updates); err != nil {
		response.Fail(c, 500, "更新失败")
		return
	}
	response.Success(c, nil)
}

// DeleteSection 删除攻略板块
// @Summary 删除攻略板块
// @Security BearerAuth
// @Tags 小程序-攻略
// @Param id path string true "板块ID"
// @Success 200 {object} response.Response
// @Router /api/v1/guide/section/{id} [delete]
func DeleteSection(c *gin.Context) {
	id := c.Param("id")
	if err := repository.DeleteSection(id); err != nil {
		response.Fail(c, 500, "删除失败")
		return
	}
	response.Success(c, nil)
}

// ReorderSections 板块拖拽排序
// @Summary 板块排序
// @Security BearerAuth
// @Tags 小程序-攻略
// @Param body body object{guideId=string,sectionIds=[]string} true "排序数据"
// @Success 200 {object} response.Response
// @Router /api/v1/guide/sections/reorder [put]
func ReorderSections(c *gin.Context) {
	var req struct {
		GuideID    string   `json:"guideId"`
		SectionIDs []string `json:"sectionIds"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 400, "参数错误")
		return
	}
	if err := repository.ReorderSections(req.GuideID, req.SectionIDs); err != nil {
		response.Fail(c, 500, "排序失败")
		return
	}
	response.Success(c, nil)
}
