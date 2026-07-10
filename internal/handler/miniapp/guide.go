package miniapp

import (
	"log"
	"strconv"
	"strings"

	"travel-server/internal/middleware"
	"travel-server/internal/model"
	"travel-server/internal/repository"
	"travel-server/pkg/response"
	"travel-server/pkg/snowflake"

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

// CreateGuide 创建攻略（含每日行程）
// @Summary 创建攻略
// @Security BearerAuth
// @Tags 小程序-攻略
// @Param body body model.CreateGuideReq true "攻略基本信息（可选传入每日行程）"
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
		ID:              snowflake.GenerateID(),
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
		IsOriginal:      req.IsOriginal,
		Status:          req.Status,
	}
	// 组装每日行程（如果未传 days，自动创建第1天空天）
	days := make([]model.GuideSection, 0)
	if len(req.Days) > 0 {
		for _, d := range req.Days {
			day := model.GuideSection{
				Title: d.Title,
				Date:  d.Date,
			}
			// 组装行程项
			items := make([]model.GuideDayItem, len(d.Items))
			for j, it := range d.Items {
				// 图片最多9张
				images := it.Images
				if len(images) > 9 {
					images = images[:9]
				}
				items[j] = model.GuideDayItem{
					SectionType:     it.SectionType,
					Title:           it.Title,
					Description:     it.Description,
					StartTime:       it.StartTime,
					EndTime:         it.EndTime,
					Latitude:        it.Latitude,
					Longitude:       it.Longitude,
					Address:         it.Address,
					Images:          images,
					NeedReservation: it.NeedReservation,
					TicketChannel:   it.TicketChannel,
					TicketPrice:     it.TicketPrice,
					TransportMode:   it.TransportMode,
					StartPoint:      it.StartPoint,
					EndPoint:        it.EndPoint,
					StartLat:        it.StartLat,
					StartLng:        it.StartLng,
					EndLat:          it.EndLat,
					EndLng:          it.EndLng,
				}
			}
			day.Items = items
			days = append(days, day)
		}
	} else {
		// 未传 days，自动创建第1天空天
		days = append(days, model.GuideSection{
			Title: "第1天",
		})
	}

	// 事务写入
	if err := repository.CreateGuideWithDays(&guide, days); err != nil {
		log.Printf("创建攻略失败: %v", err)
		response.Fail(c, 500, "创建失败: "+err.Error())
		return
	}
	response.Success(c, gin.H{"id": guide.ID})
}

// GetGuideDetail 获取攻略详情（含每日行程和行程项）
// @Summary 攻略详情
// @Tags 小程序-攻略
// @Param id path string true "攻略ID"
// @Success 200 {object} response.Response{data=object{guide=model.Guide,days=[]model.GuideSection,favoriteCount=int,commentCount=int,isLiked=bool,isFavorited=bool}}
// @Router /api/v1/guide/{id} [get]
func GetGuideDetail(c *gin.Context) {
	id := c.Param("id")
	guide, err := repository.GetGuideByID(id)
	if err != nil {
		response.Fail(c, 404, "攻略不存在")
		return
	}
	days, _ := repository.GetDaysByGuideID(id)
	userID := c.MustGet("userID").(string)
	// 增加浏览量（非作者）
	if guide.UserID != userID {
		_ = repository.IncrementGuideViewCount(id)
	}
	// 收藏数、评论数、点赞状态、收藏状态
	favoriteCount := repository.GetGuideFavoriteCount(id)
	commentCount := repository.GetGuideCommentCount(id)
	isLiked := repository.IsGuideLikedByUser(userID, id)
	isFavorited := repository.IsFavorited(userID, id, "guide")
	response.Success(c, gin.H{
		"guide":         guide,
		"days":          days,
		"likeCount":     guide.LikeCount,
		"favoriteCount": favoriteCount,
		"commentCount":  commentCount,
		"isLiked":       isLiked,
		"isFavorited":   isFavorited,
	})
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
	if req.IsOriginal != nil {
		m["is_original"] = *req.IsOriginal
	}
	if req.Status != nil {
		m["status"] = *req.Status
	}
	return m
}

// ==================== 每日行程管理 ====================

// CreateGuideDay 添加一天行程到攻略
// @Summary 添加攻略天数
// @Security BearerAuth
// @Tags 小程序-攻略
// @Param id path string true "攻略ID"
// @Param body body object{title=string,date=string} true "天数信息"
// @Success 200 {object} response.Response
// @Router /api/v1/guide/{id}/day [post]
func CreateGuideDay(c *gin.Context) {
	guideID := c.Param("id")
	// 校验攻略存在且是作者
	guide, err := repository.GetGuideByID(guideID)
	if err != nil {
		response.Fail(c, 404, "攻略不存在")
		return
	}
	if guide.UserID != c.MustGet("userID").(string) {
		response.Fail(c, 403, "无权操作")
		return
	}
	var req struct {
		Date  string `json:"date"`
		Title string `json:"title"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 400, "参数错误")
		return
	}
	day := model.GuideSection{
		GuideID: guideID,
		Title:   req.Title,
		Date:    req.Date,
	}
	if err := repository.CreateDay(&day); err != nil {
		response.Fail(c, 500, "添加天数失败")
		return
	}
	response.Success(c, day)
}

// DeleteGuideDay 删除一天行程（含其下所有行程项）
// @Summary 删除天数
// @Security BearerAuth
// @Tags 小程序-攻略
// @Param id path string true "天数ID"
// @Success 200 {object} response.Response
// @Router /api/v1/guide/day/{id} [delete]
func DeleteGuideDay(c *gin.Context) {
	dayID := c.Param("id")
	day, err := repository.GetDayByID(dayID)
	if err != nil {
		response.Fail(c, 404, "天数不存在")
		return
	}
	// 校验攻略归属
	guide, err := repository.GetGuideByID(day.GuideID)
	if err != nil || guide.UserID != c.MustGet("userID").(string) {
		response.Fail(c, 403, "无权操作")
		return
	}
	if err := repository.DeleteDay(dayID); err != nil {
		response.Fail(c, 500, "删除失败")
		return
	}
	response.Success(c, nil)
}

// ==================== 行程项管理 ====================

// CreateGuideDayItem 添加行程项
// @Summary 添加行程项
// @Security BearerAuth
// @Tags 小程序-攻略
// @Param id path string true "天数ID"
// @Param body body model.DayItemReq true "行程项内容"
// @Success 200 {object} response.Response
// @Router /api/v1/guide/day/{id}/item [post]
func CreateGuideDayItem(c *gin.Context) {
	dayID := c.Param("id")
	// 校验天存在且属于当前用户
	day, err := repository.GetDayByID(dayID)
	if err != nil {
		response.Fail(c, 404, "天数不存在")
		return
	}
	guide, err := repository.GetGuideByID(day.GuideID)
	if err != nil || guide.UserID != c.MustGet("userID").(string) {
		response.Fail(c, 403, "无权操作")
		return
	}

	var req model.DayItemReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 400, "参数错误")
		return
	}
	if !model.ValidSectionTypes[req.SectionType] {
		response.Fail(c, 400, "无效的板块类型")
		return
	}

	// 图片最多9张
	images := req.Images
	if len(images) > 9 {
		images = images[:9]
	}
	item := model.GuideDayItem{
		DayID:       dayID,
		SectionType: req.SectionType,
		Title:       req.Title,
		Description: req.Description,
		StartTime:   req.StartTime,
		EndTime:     req.EndTime,
		Latitude:    req.Latitude,
		Longitude:   req.Longitude,
		Address:     req.Address,
		Images:      images,
	}
	if err := repository.CreateDayItem(&item); err != nil {
		response.Fail(c, 500, "添加行程项失败")
		return
	}
	response.Success(c, item)
}

// UpdateGuideDayItem 更新行程项
// @Summary 更新行程项
// @Security BearerAuth
// @Tags 小程序-攻略
// @Param id path string true "行程项ID"
// @Param body body object true "更新数据"
// @Success 200 {object} response.Response
// @Router /api/v1/guide/day/item/{id} [put]
func UpdateGuideDayItem(c *gin.Context) {
	id := c.Param("id")
	item, err := repository.GetDayItemByID(id)
	if err != nil {
		response.Fail(c, 404, "行程项不存在")
		return
	}
	// 校验归属（通过 day -> guide -> user）
	day, err := repository.GetDayByID(item.DayID)
	if err != nil {
		response.Fail(c, 404, "所属天数不存在")
		return
	}
	guide, err := repository.GetGuideByID(day.GuideID)
	if err != nil || guide.UserID != c.MustGet("userID").(string) {
		response.Fail(c, 403, "无权操作")
		return
	}

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
	if err := repository.UpdateDayItem(id, updates); err != nil {
		response.Fail(c, 500, "更新失败")
		return
	}
	response.Success(c, nil)
}

// DeleteGuideDayItem 删除行程项
// @Summary 删除行程项
// @Security BearerAuth
// @Tags 小程序-攻略
// @Param id path string true "行程项ID"
// @Success 200 {object} response.Response
// @Router /api/v1/guide/day/item/{id} [delete]
func DeleteGuideDayItem(c *gin.Context) {
	id := c.Param("id")
	item, err := repository.GetDayItemByID(id)
	if err != nil {
		response.Fail(c, 404, "行程项不存在")
		return
	}
	// 校验归属
	day, err := repository.GetDayByID(item.DayID)
	if err != nil {
		response.Fail(c, 404, "所属天数不存在")
		return
	}
	guide, err := repository.GetGuideByID(day.GuideID)
	if err != nil || guide.UserID != c.MustGet("userID").(string) {
		response.Fail(c, 403, "无权操作")
		return
	}
	if err := repository.DeleteDayItem(id); err != nil {
		response.Fail(c, 500, "删除失败")
		return
	}
	response.Success(c, nil)
}

// ==================== 点赞 / 取消点赞 ====================

// LikeGuide 点赞攻略
// @Summary 点赞攻略
// @Security BearerAuth
// @Tags 小程序-攻略
// @Param id path string true "攻略ID"
// @Success 200 {object} response.Response
// @Router /api/v1/guide/{id}/like [post]
func LikeGuide(c *gin.Context) {
	guideID := c.Param("id")
	userID := c.MustGet("userID").(string)
	if err := repository.LikeGuide(userID, guideID); err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	response.Success(c, nil)
}

// UnlikeGuide 取消点赞攻略
// @Summary 取消点赞攻略
// @Security BearerAuth
// @Tags 小程序-攻略
// @Param id path string true "攻略ID"
// @Success 200 {object} response.Response
// @Router /api/v1/guide/{id}/like [delete]
func UnlikeGuide(c *gin.Context) {
	guideID := c.Param("id")
	userID := c.MustGet("userID").(string)
	if err := repository.UnlikeGuide(userID, guideID); err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	response.Success(c, nil)
}
