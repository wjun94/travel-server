package miniapp

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"travel-server/internal/ai"
	"travel-server/internal/model"
	"travel-server/internal/repository"
	"travel-server/internal/ws"
	"travel-server/pkg/database"
	"travel-server/pkg/response"
)

// AIGenerateTrip 调用 DeepSeek 智能生成行程
// @Summary AI生成行程
// @Description 根据目的地和天数调用AI生成完整行程，自动填充标题/国家/省市/预算/概述等字段并保存（每日基础1次，邀请好友成功1人可额外+1次，超出返回400）
// @Security BearerAuth
// @Tags 小程序-行程
// @Param body body object{destination=string,days=int} true "生成参数：destination=目的地, days=旅行天数"
// @Success 200 {object} response.Response{data=model.Trip} "返回完整行程（含国家/省市/预算/概述及每日行程项）"
// @Router /api/v1/trip/ai-generate [post]
func AIGenerateTrip(c *gin.Context) {
	var req struct {
		Destination string `json:"destination" binding:"required"`
		Days        int    `json:"days" binding:"required"`
		StartDate   string `json:"startDate"` // 用户选择的出发日期（可选，指定后每天日期按此顺延）
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 400, "参数错误")
		return
	}
	uid := c.MustGet("userID").(string)

	// 额度校验：管理员不限次数，其他用户今日基础1次 + 邀请成功奖励，超出拒绝
	user, _ := repository.GetUserByID(uid)
	if user == nil || user.Role != 2 {
		inviteCount, _ := repository.CountTodayInviteSuccess(uid)
		tripUsed, _ := repository.CountTodayAITrips(uid)
		if int(tripUsed) >= 1+int(inviteCount) {
			response.Fail(c, 400, "今日AI生成次数已用完，邀请好友可额外获得次数")
			return
		}
	}

	// 出发日期说明（未指定时提示从今天开始）
	startDesc := "未指定（从今天开始）"
	if req.StartDate != "" {
		startDesc = req.StartDate
	}
	prompt := fmt.Sprintf(ai.TripPrompt, req.Destination, req.Days, startDesc)
	result, err := ai.Chat(prompt)
	if err != nil {
		response.Fail(c, 500, "AI生成失败")
		return
	}

	// 解析 AI 返回的行程数据（含行程基础信息与每日行程）
	var aiResult struct {
		Title       string   `json:"title"`
		Countries   []string `json:"countries"`
		Provinces   []string `json:"provinces"`
		Cities      []string `json:"cities"`
		IsOverseas  int      `json:"isOverseas"`
		TotalBudget float64  `json:"totalBudget"`
		Summary     string   `json:"summary"`
		Days        []struct {
			Day   int    `json:"day"`
			Date  string `json:"date"`
			Items []struct {
				Time        string `json:"time"`
				Name        string `json:"name"`
				Type        string `json:"type"`
				Duration    string `json:"duration"`
				Address     string `json:"address"`
				Description string `json:"description"`
			} `json:"items"`
		} `json:"days"`
	}
	if err := json.Unmarshal([]byte(result), &aiResult); err != nil {
		response.Fail(c, 500, "AI返回格式异常")
		return
	}

	// AI 生成不落库（不保存草稿），仅返回生成数据；用户确认发布后通过发布接口新建
	trip := model.Trip{
		ID:           fmt.Sprintf("ai_%d", time.Now().UnixNano()), // 临时标识，仅用于前端流程串联，非库中记录ID
		UserID:       uid,
		Title:        aiResult.Title,
		Countries:    aiResult.Countries,
		Provinces:    aiResult.Provinces,
		Cities:       aiResult.Cities,
		Destinations: []string{req.Destination},
		IsOverseas:   aiResult.IsOverseas,
		TotalBudget:  aiResult.TotalBudget,
		Summary:      aiResult.Summary,
		Status:       0, // 非草稿
		IsPublic:     1, // 默认公开
		IsAI:         1, // AI生成
	}
	// 兜底：AI未返回标题时用目的地+天数拼接
	if trip.Title == "" {
		trip.Title = fmt.Sprintf("%s%d日游", req.Destination, req.Days)
	}

	// 行程日及行程项（用户指定出发日期时，每天日期按出发日期顺延，AI 返回的日期不采用）
	dayList := make([]model.TripDay, 0, len(aiResult.Days))
	for i, d := range aiResult.Days {
		dayNumber := d.Day
		if dayNumber <= 0 {
			dayNumber = i + 1
		}
		dayDate := d.Date
		if req.StartDate != "" {
			dayDate = addDaysStr(req.StartDate, dayNumber-1)
		}
		items := make([]model.TripItem, 0, len(d.Items))
		for _, item := range d.Items {
			items = append(items, model.TripItem{
				StartTime:   item.Time,
				SectionType: item.Type,
				Title:       item.Name,
				Address:     item.Address,
				Description: item.Description,
			})
		}
		dayList = append(dayList, model.TripDay{
			DayNumber: dayNumber,
			Date:      dayDate,
			Title:     fmt.Sprintf("第%d天", dayNumber),
			Items:     items,
		})
	}

	// 响应：平铺行程字段 + 每日行程数组（前端 AI 流程依赖 days 数组结构）
	respData := map[string]interface{}{
		"id":           trip.ID,
		"userId":       uid,
		"title":        trip.Title,
		"coverImage":   "",
		"countries":    aiResult.Countries,
		"provinces":    aiResult.Provinces,
		"cities":       aiResult.Cities,
		"destinations": []string{req.Destination},
		"isOverseas":   aiResult.IsOverseas,
		"totalBudget":  aiResult.TotalBudget,
		"summary":      aiResult.Summary,
		"status":       0, // 非草稿
		"isPublic":     1,
		"isAI":         1,
		"days":         dayList,
	}
	response.Success(c, respData)
}

// addDaysStr 按偏移天数计算日期字符串（YYYY-MM-DD + offset 天）
func addDaysStr(dateStr string, offset int) string {
	t, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return dateStr
	}
	return t.AddDate(0, 0, offset).Format("2006-01-02")
}

// CreateTrip 手动创建行程
// @Summary 创建手动行程
// @Security BearerAuth
// @Tags 小程序-行程
// @Param body body model.Trip true "行程信息"
// @Success 200 {object} response.Response
// @Router /api/v1/trip [post]
func CreateTrip(c *gin.Context) {
	var trip model.Trip
	if err := c.ShouldBindJSON(&trip); err != nil {
		response.Fail(c, 400, "参数错误")
		return
	}
	// 内容安全检测（论坛场景：标题/摘要/目的地文本）
	uid := c.MustGet("userID").(string)
	if !secGuard(c, uid, secSceneForum, secText(trip.Title, trip.Summary,
		strings.Join(trip.Countries, " "), strings.Join(trip.Provinces, " "),
		strings.Join(trip.Cities, " "), strings.Join(trip.Destinations, " "))) {
		return
	}
	trip.UserID = uid
	if err := repository.CreateTrip(&trip); err != nil {
		response.Fail(c, 500, "创建失败")
		return
	}
	// 作者信息
	author, _ := repository.GetUserByID(trip.UserID)
	authorName := ""
	authorAvatar := ""
	if author != nil {
		authorName = author.Nickname
		authorAvatar = author.AvatarURL
	}
	response.Success(c, gin.H{
		"id":            trip.ID,
		"userId":        trip.UserID,
		"guideId":       trip.GuideID,
		"title":         trip.Title,
		"coverImage":    trip.CoverImage,
		"countries":     trip.Countries,
		"provinces":     trip.Provinces,
		"cities":        trip.Cities,
		"destinations":  trip.Destinations,
		"totalBudget":   trip.TotalBudget,
		"isOverseas":    trip.IsOverseas,
		"summary":       trip.Summary,
		"viewCount":     trip.ViewCount,
		"likeCount":     trip.LikeCount,
		"favoriteCount": trip.FavoriteCount,
		"status":        trip.Status,
		"isPublic":      trip.IsPublic,
		"createdAt":     trip.CreatedAt,
		"updatedAt":     trip.UpdatedAt,
		"authorName":    authorName,
		"authorAvatar":  authorAvatar,
		"isFollowed":    false,
		"isSelf":        true,
	})
}

// GetTrip 获取行程详情
// @Summary 获取行程详情
// @Tags 小程序-行程
// @Param id path string true "行程ID"
// @Success 200 {object} response.Response{data=model.Trip}
// @Router /api/v1/trip/{id} [get]
func GetTrip(c *gin.Context) {
	id := c.Param("id")
	trip, err := repository.GetTripByID(id)
	if err != nil {
		response.Fail(c, 500, "行程不存在")
		return
	}
	userID := c.MustGet("userID").(string)

	// 增加浏览量（非作者）
	if trip.UserID != userID {
		_ = repository.IncrementTripViewCount(id)
	}

	// 作者信息
	author, _ := repository.GetUserByID(trip.UserID)
	authorName := ""
	authorAvatar := ""
	if author != nil {
		authorName = author.Nickname
		authorAvatar = author.AvatarURL
	}
	// 关注状态
	followStatus, _ := repository.GetFollowStatus(userID, trip.UserID)
	isFollowed := followStatus == 1 || followStatus == 2

	// 收藏数、评论数、点赞状态（点赞与收藏独立，action 区分）
	favoriteCount := repository.GetTripFavoriteCount(id)
	commentCount := repository.GetTripCommentCount(id)
	isLiked := repository.IsTripLikedByUser(userID, id)
	isFavorited := repository.IsFavorited(userID, id, "trip")

	// 将 trip 展开到顶层，不嵌套在单独的 trip 字段里
	result := gin.H{
		"id":            trip.ID,
		"userId":        trip.UserID,
		"guideId":       trip.GuideID,
		"title":         trip.Title,
		"coverImage":    trip.CoverImage,
		"countries":     trip.Countries,
		"provinces":     trip.Provinces,
		"cities":        trip.Cities,
		"destinations":  trip.Destinations,
		"totalBudget":   trip.TotalBudget,
		"isOverseas":    trip.IsOverseas,
		"summary":       trip.Summary,
		"viewCount":     trip.ViewCount,
		"likeCount":     trip.LikeCount,
		"favoriteCount": favoriteCount,
		"commentCount":  commentCount,
		"status":        trip.Status,
		"isPublic":      trip.IsPublic,
		"createdAt":     trip.CreatedAt,
		"updatedAt":     trip.UpdatedAt,
		"days":          trip.Days,
		"members":       trip.Members,
		"isLiked":       isLiked,
		"isFavorited":   isFavorited,
		"authorName":    authorName,
		"authorAvatar":  authorAvatar,
		"isFollowed":    isFollowed,
		"isSelf":        userID == trip.UserID,
	}

	response.Success(c, result)
}

// GetMyTrips 我的行程列表
// @Summary 我的行程列表
// @Security BearerAuth
// @Tags 小程序-行程
// @Param page query int false "页码"
// @Param pageSize query int false "每页数量"
// @Param status query int false "状态筛选（1草稿 2已发布，-1或不传为全部）" default(-1)
// @Success 200 {object} response.Response{data=object{list=[]model.Trip,total=int}}
// @Router /api/v1/my/trips [get]
func GetMyTrips(c *gin.Context) {
	userID := c.MustGet("userID").(string)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	// 容错：status 解析失败或非数字时视为全部（-1），避免前端误传对象导致误筛草稿
	status := -1
	if v := c.Query("status"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			status = n
		}
	}

	trips, total, err := repository.ListUserTrips(userID, page, pageSize, status)
	if err != nil {
		response.Fail(c, 500, "获取失败")
		return
	}

	// 批量查询行程的天数和行程项统计
	tripIDs := make([]string, len(trips))
	for i, t := range trips {
		tripIDs[i] = t.ID
	}

	type tripDayCount struct {
		TripID string
		Days   int
	}
	var tripDayCounts []tripDayCount
	if len(tripIDs) > 0 {
		database.DB.Model(&model.TripDay{}).
			Select("trip_id, COUNT(DISTINCT day_number) as days").
			Where("trip_id IN ?", tripIDs).
			Group("trip_id").
			Find(&tripDayCounts)
	}
	dayMap := make(map[string]int)
	for _, d := range tripDayCounts {
		dayMap[d.TripID] = d.Days
	}

	type tripSecCount struct {
		TripID string
		Count  int64
	}
	var tripSecCounts []tripSecCount
	if len(tripIDs) > 0 {
		database.DB.Table("trip_items ti").
			Select("td.trip_id, COUNT(ti.id) as count").
			Joins("LEFT JOIN trip_days td ON td.id = ti.trip_day_id").
			Where("td.trip_id IN ?", tripIDs).
			Group("td.trip_id").
			Find(&tripSecCounts)
	}
	secMap := make(map[string]int64)
	for _, s := range tripSecCounts {
		secMap[s.TripID] = s.Count
	}

	// 组装返回数据
	list := make([]gin.H, 0, len(trips))
	for _, t := range trips {
		list = append(list, gin.H{
			"id":            t.ID,
			"userId":        t.UserID,
			"guideId":       t.GuideID,
			"title":         t.Title,
			"coverImage":    t.CoverImage,
			"countries":     t.Countries,
			"provinces":     t.Provinces,
			"cities":        t.Cities,
			"destinations":  t.Destinations,
			"totalBudget":   t.TotalBudget,
			"isOverseas":    t.IsOverseas,
			"summary":       t.Summary,
			"viewCount":     t.ViewCount,
			"likeCount":     t.LikeCount,
			"favoriteCount": t.FavoriteCount,
			"tripDays":      dayMap[t.ID],
			"sectionCount":  secMap[t.ID],
			"status":        t.Status,
			"isPublic":      t.IsPublic,
			"createdAt":     t.CreatedAt,
			"updatedAt":     t.UpdatedAt,
		})
	}

	response.Success(c, gin.H{"list": list, "total": total})
}

// DeleteTrip 删除行程（仅作者本人，级联删除行程日、行程项及同行者）
// @Summary 删除行程
// @Security BearerAuth
// @Tags 小程序-行程
// @Param id path string true "行程ID"
// @Success 200 {object} response.Response
// @Router /api/v1/trip/{id} [delete]
func DeleteTrip(c *gin.Context) {
	id := c.Param("id")
	trip, err := repository.GetTripByID(id)
	if err != nil {
		response.Fail(c, 404, "行程不存在")
		return
	}
	if trip.UserID != c.MustGet("userID").(string) {
		response.Fail(c, 403, "无权限")
		return
	}
	if err := repository.DeleteTripCascade(id); err != nil {
		response.Fail(c, 500, "删除失败")
		return
	}
	response.Success(c, nil)
}

// UpdateTrip 更新行程基本信息
// @Summary 更新行程
// @Security BearerAuth
// @Tags 小程序-行程
// @Param id path string true "行程ID"
// @Param body body object true "更新数据"
// @Success 200 {object} response.Response
// @Router /api/v1/trip/{id} [put]
func UpdateTrip(c *gin.Context) {
	id := c.Param("id")
	trip, err := repository.GetTripByID(id)
	if err != nil {
		response.Fail(c, 404, "行程不存在")
		return
	}
	// 权限检查：仅创建者可编辑
	userID := c.MustGet("userID").(string)
	if trip.UserID != userID {
		response.Fail(c, 403, "无编辑权限")
		return
	}

	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		response.Fail(c, 400, "参数错误")
		return
	}
	// 过滤不可更新的字段
	delete(updates, "id")
	delete(updates, "user_id")
	delete(updates, "created_at")
	delete(updates, "view_count")
	delete(updates, "like_count")
	delete(updates, "favorite_count")

	// JSON 数组字段：map 更新不应用 serializer，需显式序列化，避免存明文
	for _, f := range []string{"countries", "provinces", "cities", "destinations"} {
		if v, ok := updates[f]; ok && v != nil {
			if b, err := json.Marshal(v); err == nil {
				updates[f] = string(b)
			}
		}
	}

	// 内容安全检测（仅检测本次更新的文本字段）
	secParts := make([]string, 0, 6)
	for _, k := range []string{"title", "summary", "countries", "provinces", "cities", "destinations"} {
		if v, ok := updates[k]; ok && v != nil {
			secParts = append(secParts, fmt.Sprintf("%v", v))
		}
	}
	if !secGuard(c, userID, secSceneForum, secText(secParts...)) {
		return
	}

	// 传入 days 时全量替换行程日（删除旧行程重建）
	days, hasDays := updates["days"].([]interface{})
	if hasDays {
		delete(updates, "days")
		dayList := make([]model.TripDay, 0, len(days))
		b, err := json.Marshal(days)
		if err == nil {
			_ = json.Unmarshal(b, &dayList)
		}
		if err := repository.UpdateTripWithDays(id, updates, dayList); err != nil {
			response.Fail(c, 500, "更新失败")
			return
		}
		response.Success(c, nil)
		return
	}

	if err := repository.UpdateTrip(id, updates); err != nil {
		response.Fail(c, 500, "更新失败")
		return
	}
	response.Success(c, nil)
}

// ==================== TripDay 行程日 ====================

// AddTripDay 添加行程日
// @Summary 添加行程日
// @Security BearerAuth
// @Tags 小程序-行程
// @Param body body model.TripDay true "行程日信息"
// @Success 200 {object} response.Response
// @Router /api/v1/trip/day [post]
func AddTripDay(c *gin.Context) {
	var day model.TripDay
	if err := c.ShouldBindJSON(&day); err != nil {
		response.Fail(c, 400, "参数错误")
		return
	}
	if err := repository.CreateTripDay(&day); err != nil {
		response.Fail(c, 500, "添加失败")
		return
	}
	response.Success(c, day)
}

// UpdateTripDay 更新行程日
// @Summary 更新行程日
// @Security BearerAuth
// @Tags 小程序-行程
// @Param id path string true "行程日ID"
// @Param body body object true "更新数据"
// @Success 200 {object} response.Response
// @Router /api/v1/trip/day/{id} [put]
func UpdateTripDay(c *gin.Context) {
	id := c.Param("id")
	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		response.Fail(c, 400, "参数错误")
		return
	}
	if err := repository.UpdateTripDay(id, updates); err != nil {
		response.Fail(c, 500, "更新失败")
		return
	}
	response.Success(c, nil)
}

// DeleteTripDay 删除行程日
// @Summary 删除行程日
// @Security BearerAuth
// @Tags 小程序-行程
// @Param id path string true "行程日ID"
// @Success 200 {object} response.Response
// @Router /api/v1/trip/day/{id} [delete]
func DeleteTripDay(c *gin.Context) {
	id := c.Param("id")
	if err := repository.DeleteTripDay(id); err != nil {
		response.Fail(c, 500, "删除失败")
		return
	}
	response.Success(c, nil)
}

// ==================== TripItem 行程项 ====================

// AddTripItem 添加行程项
// @Summary 添加行程项
// @Security BearerAuth
// @Tags 小程序-行程
// @Param body body model.TripItem true "行程项信息"
// @Success 200 {object} response.Response
// @Router /api/v1/trip/item [post]
func AddTripItem(c *gin.Context) {
	var item model.TripItem
	if err := c.ShouldBindJSON(&item); err != nil {
		response.Fail(c, 400, "参数错误")
		return
	}
	if item.SectionType != "" && !model.ValidSectionTypes[item.SectionType] {
		response.Fail(c, 400, "无效的行程项类型，可选：transport/hotel/attraction/food/shopping/tips")
		return
	}
	if err := repository.CreateTripItem(&item); err != nil {
		response.Fail(c, 500, "添加失败")
		return
	}
	response.Success(c, item)
}

// UpdateTripItem 更新行程项
// @Summary 更新行程项
// @Security BearerAuth
// @Tags 小程序-行程
// @Param id path string true "行程项ID"
// @Param body body object true "更新数据"
// @Success 200 {object} response.Response
// @Router /api/v1/trip/item/{id} [put]
func UpdateTripItem(c *gin.Context) {
	id := c.Param("id")
	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		response.Fail(c, 400, "参数错误")
		return
	}
	if sectionType, ok := updates["sectionType"].(string); ok && sectionType != "" && !model.ValidSectionTypes[sectionType] {
		response.Fail(c, 400, "无效的行程项类型，可选：transport/hotel/attraction/food/shopping/tips")
		return
	}
	if err := repository.UpdateTripItem(id, updates); err != nil {
		response.Fail(c, 500, "更新失败")
		return
	}
	response.Success(c, nil)
}

// DeleteTripItem 删除行程项
// @Summary 删除行程项
// @Security BearerAuth
// @Tags 小程序-行程
// @Param id path string true "行程项ID"
// @Success 200 {object} response.Response
// @Router /api/v1/trip/item/{id} [delete]
func DeleteTripItem(c *gin.Context) {
	id := c.Param("id")
	if err := repository.DeleteTripItem(id); err != nil {
		response.Fail(c, 500, "删除失败")
		return
	}
	response.Success(c, nil)
}

// ==================== TripMember 同行者 ====================

// InviteMember 邀请同行者
// @Summary 邀请同行者
// @Security BearerAuth
// @Tags 小程序-行程
// @Param body body model.TripMember true "同行者信息"
// @Success 200 {object} response.Response
// @Router /api/v1/trip/member [post]
func InviteMember(c *gin.Context) {
	var member model.TripMember
	if err := c.ShouldBindJSON(&member); err != nil {
		response.Fail(c, 400, "参数错误")
		return
	}
	if err := repository.AddTripMember(&member); err != nil {
		response.Fail(c, 500, "邀请失败")
		return
	}
	response.Success(c, member)
}

// RemoveMember 移除同行者
// @Summary 移除同行者
// @Security BearerAuth
// @Tags 小程序-行程
// @Param id path string true "同行者记录ID"
// @Success 200 {object} response.Response
// @Router /api/v1/trip/member/{id} [delete]
func RemoveMember(c *gin.Context) {
	id := c.Param("id")
	if err := repository.RemoveTripMember(id); err != nil {
		response.Fail(c, 500, "移除失败")
		return
	}
	response.Success(c, nil)
}

// LikeTrip 点赞行程
// @Summary 点赞行程
// @Security BearerAuth
// @Tags 小程序-行程
// @Param id path string true "行程ID"
// @Success 200 {object} response.Response
// @Router /api/v1/trip/{id}/like [post]
func LikeTrip(c *gin.Context) {
	id := c.Param("id")
	userID := c.MustGet("userID").(string)
	if err := repository.LikeTrip(userID, id); err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	// 通知行程作者（非本人点赞才通知）+ 实时推送
	trip, _ := repository.GetTripByID(id)
	if trip != nil && trip.UserID != userID {
		notification := model.Notification{
			UserID:     trip.UserID,
			FromUserID: userID,
			Type:       2,
			RelatedID:  id,
			Content:    "您的行程收到一个赞",
		}
		if err := repository.CreateNotification(&notification); err == nil {
			ws.WsHub.PushToUser(trip.UserID, map[string]interface{}{
				"action":  "new_notification",
				"type":    2, // 点赞
				"content": notification.Content,
			})
		}
	}
	response.Success(c, nil)
}

// UnlikeTrip 取消点赞行程
// @Summary 取消点赞行程
// @Security BearerAuth
// @Tags 小程序-行程
// @Param id path string true "行程ID"
// @Success 200 {object} response.Response
// @Router /api/v1/trip/{id}/like [delete]
func UnlikeTrip(c *gin.Context) {
	id := c.Param("id")
	userID := c.MustGet("userID").(string)
	if err := repository.UnlikeTrip(userID, id); err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	response.Success(c, nil)
}
