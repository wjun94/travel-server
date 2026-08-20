package miniapp

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"travel-server/internal/ai"
	"travel-server/internal/model"
	"travel-server/internal/repository"
	"travel-server/internal/service"
	"travel-server/internal/ws"
	"travel-server/pkg/response"
)

// CreatePartnerReq 发布搭子请求
type CreatePartnerReq struct {
	Title           string          `json:"title"`           // 搭子标题
	Cover           string          `json:"cover"`           // 封面图
	Images          []string        `json:"images"`          // 多图列表
	Category        string          `json:"category"`        // 活动分类：旅游/美食/运动/学习/探店/看展/桌游
	Destination     string          `json:"destination"`     // 目的地
	Longitude       float64         `json:"longitude"`       // 经度
	Latitude        float64         `json:"latitude"`        // 纬度
	Address         string          `json:"address"`         // 详细地址
	LocationType    int             `json:"locationType"`    // 0线下 1线上
	OnlineLink      string          `json:"onlineLink"`      // 线上链接
	StartDate       string          `json:"startDate"`       // YYYY-MM-DD
	EndDate         string          `json:"endDate"`         // YYYY-MM-DD
	TravelTags      string          `json:"travelTags"`      // 逗号分隔（兼容旧版）
	Tags            string          `json:"tags"`            // 多选标签JSON数组
	Desc            string          `json:"desc"`            // 行程简述
	RichDesc        string          `json:"richDesc"`        // 详细介绍（富文本）
	Requirement     string          `json:"requirement"`     // 人员要求
	MaxMembers      int             `json:"maxMembers"`      // 招募上限
	MinMembers      int             `json:"minMembers"`      // 最小成团人数
	GenderLimit     int             `json:"genderLimit"`     // 0不限 1仅男生 2仅女生
	MaleCount       int             `json:"maleCount"`       // 男生需求数
	FemaleCount     int             `json:"femaleCount"`     // 女生需求数
	MinAge          int             `json:"minAge"`          // 年龄下限
	MaxAge          int             `json:"maxAge"`          // 年龄上限
	FeeMode         int             `json:"feeMode"`         // 0免费 1AA 2组织者全包 3人均固定预算
	BudgetPerPerson int             `json:"budgetPerPerson"` // 人均预算
	OfficialPrice   float64         `json:"officialPrice"`   // 官方活动定价
	FeeInclude      string          `json:"feeInclude"`      // 费用包含
	FeeExclude      string          `json:"feeExclude"`      // 费用不含
	EstTotal        int             `json:"estTotal"`        // 预估总价
	Visibility      int             `json:"visibility"`      // 0全部可见 1同城可见 2好友可见
	JoinMode        int             `json:"joinMode"`        // 0需审核 1直接加入
	AutoClose       int             `json:"autoClose"`       // 满员自动关闭：0否 1是
	AllowShare      int             `json:"allowShare"`      // 允许转发：0否 1是
	AllowCollect    int             `json:"allowCollect"`    // 允许收藏：0否 1是
	IsDraft         int             `json:"isDraft"`         // 0已发布 1草稿
	IsPublic        int             `json:"isPublic"`        // 0仅自己可见 1公开招募
	TripID          string          `json:"tripId"`          // 关联行程ID（可选，已有行程时直接关联）
	Days            []model.TripDay `json:"days"`            // 行程安排（可选，自动创建关联草稿行程）
}

// parseDate 将 YYYY-MM-DD 格式的字符串解析为 time.Time
func parseDate(s string) *time.Time {
	if s == "" {
		return nil
	}
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return nil
	}
	return &t
}

// toPartnerDays 将行程日数据（TripDay）转换为搭子行程日列表（PartnerDay），用于落库关联表
func toPartnerDays(days []model.TripDay) []model.PartnerDay {
	res := make([]model.PartnerDay, 0, len(days))
	for _, d := range days {
		pd := model.PartnerDay{
			DayNumber: d.DayNumber,
			Date:      d.Date,
			Title:     d.Title,
		}
		for _, it := range d.Items {
			pd.Items = append(pd.Items, model.PartnerDayItem{
				SectionType:     it.SectionType,
				Title:           it.Title,
				Description:     it.Description,
				StartTime:       it.StartTime,
				EndTime:         it.EndTime,
				Latitude:        it.Latitude,
				Longitude:       it.Longitude,
				Address:         it.Address,
				Images:          it.Images,
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
			})
		}
		res = append(res, pd)
	}
	return res
}

// CreatePartner 发布搭子信息
// @Summary 发布搭子
// @Security BearerAuth
// @Tags 小程序-搭子
// @Param body body CreatePartnerReq true "搭子信息"
// @Success 200 {object} response.Response
// @Router /api/v1/partner [post]
func CreatePartner(c *gin.Context) {
	var req CreatePartnerReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 400, "参数错误")
		return
	}

	userID := c.MustGet("userID").(string)

	p := model.Partner{
		Title:           req.Title,
		Cover:           req.Cover,
		Images:          req.Images,
		Category:        req.Category,
		Destination:     req.Destination,
		Longitude:       req.Longitude,
		Latitude:        req.Latitude,
		Address:         req.Address,
		LocationType:    req.LocationType,
		OnlineLink:      req.OnlineLink,
		StartDate:       parseDate(req.StartDate),
		EndDate:         parseDate(req.EndDate),
		TravelTags:      req.TravelTags,
		Tags:            req.Tags,
		Desc:            req.Desc,
		RichDesc:        req.RichDesc,
		Requirement:     req.Requirement,
		MaxMembers:      req.MaxMembers,
		MinMembers:      req.MinMembers,
		GenderLimit:     req.GenderLimit,
		MaleCount:       req.MaleCount,
		FemaleCount:     req.FemaleCount,
		MinAge:          req.MinAge,
		MaxAge:          req.MaxAge,
		FeeMode:         req.FeeMode,
		BudgetPerPerson: req.BudgetPerPerson,
		OfficialPrice:   req.OfficialPrice,
		FeeInclude:      req.FeeInclude,
		FeeExclude:      req.FeeExclude,
		EstTotal:        req.EstTotal,
		Visibility:      req.Visibility,
		JoinMode:        req.JoinMode,
		AutoClose:       req.AutoClose,
		AllowShare:      req.AllowShare,
		AllowCollect:    req.AllowCollect,
		IsDraft:         req.IsDraft,
		IsPublic:        req.IsPublic,
	}
	p.UserID = userID
	p.Status = 0         // 默认招募中
	p.CurrentMembers = 1 // 发起人计入

	// 行程安排：转搭子行程日列表（关联表 partner_days，参考行程表结构）
	if len(req.Days) > 0 {
		p.Days = toPartnerDays(req.Days)
	}
	if err := repository.CreatePartner(&p); err != nil {
		response.Fail(c, 500, "发布失败")
		return
	}
	response.Success(c, p)
}

// GetPartnerList 获取搭子列表（公共接口，未登录也可浏览）
// @Summary 搭子列表
// @Tags 小程序-搭子
// @Param page query int false "页码"
// @Param pageSize query int false "每页数量"
// @Param keyword query string false "关键词搜索（标题/目的地/简述/标签）"
// @Success 200 {object} response.Response{data=object{list=[]object{id=string,userId=string,tripId=string,type=int,category=string,title=string,cover=string,images=string,destination=string,longitude=float64,latitude=float64,address=string,locationType=int,onlineLink=string,startDate=string,endDate=string,days=int,travelTags=string,tags=string,desc=string,richDesc=string,requirement=string,maxMembers=int,minMembers=int,currentMembers=int,genderLimit=int,maleCount=int,femaleCount=int,minAge=int,maxAge=int,feeMode=int,budgetPerPerson=int,officialPrice=float64,feeInclude=string,feeExclude=string,estTotal=int,visibility=int,joinMode=int,autoClose=int,allowShare=int,allowCollect=int,isDraft=int,status=int,isPublic=int,viewCount=int,sortWeight=int,createdAt=string,updatedAt=string,authorId=string,authorName=string,authorAvatar=string,isApplied=bool,isSelf=bool,isFollowed=bool},total=int64}}
// @Router /api/v1/partner/list [get]
func GetPartnerList(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	// 清洗关键词：去掉首尾空格，忽略 null/undefined 等无效值
	keyword := strings.TrimSpace(c.Query("keyword"))
	if keyword == "" || keyword == "null" || keyword == "undefined" {
		keyword = ""
	}
	list, total, err := repository.GetPartnerList(page, pageSize, keyword)
	if err != nil {
		response.Fail(c, 500, "获取失败")
		return
	}

	// 批量查询用户信息
	userIDs := make([]string, 0, len(list))
	for _, p := range list {
		userIDs = append(userIDs, p.UserID)
	}
	userMap := repository.GetUsersByIDs(userIDs)

	// 如果已登录，额外查询申请状态和关注状态
	userIDVal, loggedIn := c.Get("userID")
	var partnerIDs []string
	var appliedMap map[string]bool
	var followMap map[string]int
	currentUserID := ""
	if loggedIn {
		currentUserID = userIDVal.(string)
		partnerIDs = make([]string, len(list))
		for i, p := range list {
			partnerIDs[i] = p.ID
		}
		appliedMap, _ = repository.GetUserAppliedPartnerIDs(currentUserID, partnerIDs)
		followMap = repository.GetFollowStatusMap(currentUserID, userIDs)
	}

	type partnerVO struct {
		model.Partner
		AuthorID     string `json:"authorId"`     // 作者ID
		AuthorName   string `json:"authorName"`   // 作者昵称
		AuthorAvatar string `json:"authorAvatar"` // 作者头像
		IsApplied    bool   `json:"isApplied"`    // 当前用户是否已申请
		IsSelf       bool   `json:"isSelf"`       // 是否是自己创建的
		IsFollowed   bool   `json:"isFollowed"`   // 是否已关注
		ItemCount    int64  `json:"itemCount"`    // 关联行程的行程项总数
		DayCount     int    `json:"dayCount"`     // 出行天数（由行程日列表长度派生）
		SectionCount int64  `json:"sectionCount"` // 自有行程安排的行程项总数（与首页 sectionCount 一致）
	}
	result := make([]partnerVO, len(list))
	for i, p := range list {
		u := userMap[p.UserID]
		authorName := ""
		authorAvatar := ""
		if u != nil {
			authorName = u.Nickname
			authorAvatar = u.AvatarURL
		}
		// 天数由行程日列表长度派生；置空 Days 避免列表响应携带完整行程数组
		dayCount := len(p.Days)
		p.Days = nil
		result[i] = partnerVO{
			Partner:      p,
			AuthorID:     p.UserID,
			AuthorName:   authorName,
			AuthorAvatar: authorAvatar,
			IsApplied:    loggedIn && appliedMap[p.ID],
			IsSelf:       loggedIn && currentUserID == p.UserID,
			IsFollowed:   loggedIn && (followMap[p.UserID] == 1 || followMap[p.UserID] == 2),
			DayCount:     dayCount,
		}
	}

	// 批量查询关联行程的行程项总数
	var tripIDs []string
	for _, p := range list {
		if p.TripID != "" {
			tripIDs = append(tripIDs, p.TripID)
		}
	}
	if len(tripIDs) > 0 {
		itemCountMap := repository.GetTripItemCounts(tripIDs)
		for i, p := range list {
			if cnt, ok := itemCountMap[p.TripID]; ok {
				result[i].ItemCount = cnt
			}
		}
	}

	// 批量查询实时评论数，覆盖数据库静态字段
	partnerIDs = make([]string, len(list))
	for i, p := range list {
		partnerIDs[i] = p.ID
	}
	commentCountMap := repository.GetPartnerCommentCounts(partnerIDs)
	for i, p := range list {
		if cnt, ok := commentCountMap[p.ID]; ok {
			result[i].CommentCount = int(cnt)
		}
	}

	// 批量查询搭子自有行程安排的行程项总数
	partnerItemCountMap := repository.GetPartnerItemCounts(partnerIDs)
	for i, p := range list {
		if cnt, ok := partnerItemCountMap[p.ID]; ok {
			result[i].SectionCount = cnt
		}
	}

	response.Success(c, gin.H{"list": result, "total": total})
}

// GetMyPartners 我发布的搭子列表
// @Summary 我发布的搭子
// @Security BearerAuth
// @Tags 小程序-搭子
// @Param page query int false "页码"
// @Param pageSize query int false "每页数量"
// @Param isDraft query int false "草稿筛选（1草稿 0已发布，-1或不传为全部）" default(-1)
// @Success 200 {object} response.Response{data=object{list=[]object{id=string,userId=string,tripId=string,type=int,category=string,title=string,cover=string,images=string,destination=string,longitude=float64,latitude=float64,address=string,locationType=int,onlineLink=string,startDate=string,endDate=string,days=int,travelTags=string,tags=string,desc=string,richDesc=string,requirement=string,maxMembers=int,minMembers=int,currentMembers=int,genderLimit=int,maleCount=int,femaleCount=int,minAge=int,maxAge=int,feeMode=int,budgetPerPerson=int,officialPrice=float64,feeInclude=string,feeExclude=string,estTotal=int,visibility=int,joinMode=int,autoClose=int,allowShare=int,allowCollect=int,isDraft=int,status=int,isPublic=int,viewCount=int,sortWeight=int,createdAt=string,updatedAt=string,authorId=string,authorName=string,authorAvatar=string,isApplied=bool,isSelf=bool,isFollowed=bool,itemCount=int64,commentCount=int,statusText=string},total=int64}}
// @Router /api/v1/my/partners [get]
func GetMyPartners(c *gin.Context) {
	userID := c.MustGet("userID").(string)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	// 容错：isDraft 解析失败或非数字时视为全部（-1），避免前端误传对象导致误筛
	isDraft := -1
	if v := c.Query("isDraft"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			isDraft = n
		}
	}
	list, total, err := repository.GetMyPartners(userID, page, pageSize, isDraft)
	if err != nil {
		response.Fail(c, 500, "获取失败")
		return
	}

	result := enrichPartnerList(list, false, true)
	response.Success(c, gin.H{"list": result, "total": total})
}

// partnerItemVO 搭子列表项（列表接口共用响应结构）
type partnerItemVO struct {
	model.Partner
	AuthorID     string `json:"authorId"`     // 作者ID
	AuthorName   string `json:"authorName"`   // 作者昵称
	AuthorAvatar string `json:"authorAvatar"` // 作者头像
	IsApplied    bool   `json:"isApplied"`    // 当前用户是否已申请
	IsSelf       bool   `json:"isSelf"`       // 是否是自己创建的
	IsFollowed   bool   `json:"isFollowed"`   // 是否已关注
	ItemCount    int64  `json:"itemCount"`    // 关联行程的行程项总数
	DayCount     int    `json:"dayCount"`     // 出行天数（由行程日列表长度派生）
	StatusText   string `json:"statusText"`   // 状态文案：草稿/仅自己可见/招募中/已满员/已解散/已过期/行程结束
	SectionCount int64  `json:"sectionCount"` // 自有行程安排的行程项总数（与首页 sectionCount 一致）
}

// partnerStatusText 搭子状态文案（草稿/仅自己可见优先于招募状态）
func partnerStatusText(p *model.Partner) string {
	if p.IsDraft == 1 {
		return "草稿"
	}
	if p.IsPublic == 0 {
		return "仅自己可见"
	}
	switch p.Status {
	case 0:
		return "招募中"
	case 1:
		return "已满员"
	case 2:
		return "已解散"
	case 3:
		return "已过期"
	case 4:
		return "行程结束"
	default:
		return "招募中"
	}
}

// enrichPartnerList 富化搭子列表：批量注入作者信息与关联行程项数（列表/我的/参与的共用）
func enrichPartnerList(list []model.Partner, isApplied, isSelf bool) []partnerItemVO {
	result := make([]partnerItemVO, len(list))

	// 批量查询用户信息
	userIDs := make([]string, 0, len(list))
	for _, p := range list {
		userIDs = append(userIDs, p.UserID)
	}
	userMap := repository.GetUsersByIDs(userIDs)

	for i, p := range list {
		u := userMap[p.UserID]
		authorName := ""
		authorAvatar := ""
		if u != nil {
			authorName = u.Nickname
			authorAvatar = u.AvatarURL
		}
		// 天数由行程日列表长度派生；置空 Days 避免列表响应携带完整行程数组
		dayCount := len(p.Days)
		p.Days = nil
		result[i] = partnerItemVO{
			Partner:      p,
			AuthorID:     p.UserID,
			AuthorName:   authorName,
			AuthorAvatar: authorAvatar,
			IsApplied:    isApplied,
			IsSelf:       isSelf,
			IsFollowed:   false,
			DayCount:     dayCount,
			StatusText:   partnerStatusText(&p),
		}
	}

	// 批量查询关联行程的行程项总数
	var tripIDs []string
	for _, p := range list {
		if p.TripID != "" {
			tripIDs = append(tripIDs, p.TripID)
		}
	}
	if len(tripIDs) > 0 {
		itemCountMap := repository.GetTripItemCounts(tripIDs)
		for i, p := range list {
			if cnt, ok := itemCountMap[p.TripID]; ok {
				result[i].ItemCount = cnt
			}
		}
	}

	// 批量查询搭子自有行程安排的行程项总数
	partnerIDs := make([]string, len(list))
	for i, p := range list {
		partnerIDs[i] = p.ID
	}
	partnerItemCountMap := repository.GetPartnerItemCounts(partnerIDs)
	for i, p := range list {
		if cnt, ok := partnerItemCountMap[p.ID]; ok {
			result[i].SectionCount = cnt
		}
	}
	return result
}

// GetMyJoinedPartners 我参与的搭子列表（申请已通过且搭子已发布）
// @Summary 我参与的搭子
// @Security BearerAuth
// @Tags 小程序-搭子
// @Param page query int false "页码"
// @Param pageSize query int false "每页数量"
// @Success 200 {object} response.Response{data=object{list=[]object{id=string,userId=string,tripId=string,type=int,category=string,title=string,cover=string,images=string,destination=string,address=string,locationType=int,startDate=string,endDate=string,days=int,tags=string,desc=string,requirement=string,maxMembers=int,minMembers=int,currentMembers=int,status=int,isDraft=int,isPublic=int,viewCount=int,createdAt=string,authorId=string,authorName=string,authorAvatar=string,isApplied=bool,isSelf=bool,isFollowed=bool,itemCount=int64,statusText=string},total=int64}}
// @Router /api/v1/my/joined-partners [get]
func GetMyJoinedPartners(c *gin.Context) {
	userID := c.MustGet("userID").(string)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))

	list, total, err := repository.GetJoinedPartners(userID, page, pageSize)
	if err != nil {
		response.Fail(c, 500, "获取失败")
		return
	}
	result := enrichPartnerList(list, true, false)
	response.Success(c, gin.H{"list": result, "total": total})
}

// DeletePartner 删除搭子（仅作者本人，级联删除申请记录）
// @Summary 删除搭子
// @Security BearerAuth
// @Tags 小程序-搭子
// @Param id path string true "搭子ID"
// @Success 200 {object} response.Response
// @Router /api/v1/partner/{id} [delete]
func DeletePartner(c *gin.Context) {
	id := c.Param("id")
	partner, err := repository.GetPartnerByID(id)
	if err != nil {
		response.Fail(c, 404, "搭子不存在")
		return
	}
	if partner.UserID != c.MustGet("userID").(string) {
		response.Fail(c, 403, "无权限")
		return
	}
	if err := repository.DeletePartnerCascade(id); err != nil {
		response.Fail(c, 500, "删除失败")
		return
	}
	response.Success(c, nil)
}

// UpdatePartnerReq 更新搭子请求（所有字段可选）
type UpdatePartnerReq struct {
	Title           *string         `json:"title"`           // 搭子标题
	Cover           *string         `json:"cover"`           // 封面图
	Images          []string        `json:"images"`          // 多图列表
	Category        *string         `json:"category"`        // 活动分类
	Destination     *string         `json:"destination"`     // 目的地
	Longitude       *float64        `json:"longitude"`       // 经度
	Latitude        *float64        `json:"latitude"`        // 纬度
	Address         *string         `json:"address"`         // 详细地址
	LocationType    *int            `json:"locationType"`    // 0线下 1线上
	OnlineLink      *string         `json:"onlineLink"`      // 线上链接
	StartDate       *string         `json:"startDate"`       // YYYY-MM-DD
	EndDate         *string         `json:"endDate"`         // YYYY-MM-DD
	TravelTags      *string         `json:"travelTags"`      // 逗号分隔（兼容旧版）
	Tags            *string         `json:"tags"`            // 多选标签JSON数组
	Desc            *string         `json:"desc"`            // 行程简述
	RichDesc        *string         `json:"richDesc"`        // 详细介绍（富文本）
	Requirement     *string         `json:"requirement"`     // 人员要求
	MaxMembers      *int            `json:"maxMembers"`      // 招募上限
	MinMembers      *int            `json:"minMembers"`      // 最小成团人数
	GenderLimit     *int            `json:"genderLimit"`     // 0不限 1仅男生 2仅女生
	MaleCount       *int            `json:"maleCount"`       // 男生需求数
	FemaleCount     *int            `json:"femaleCount"`     // 女生需求数
	MinAge          *int            `json:"minAge"`          // 年龄下限
	MaxAge          *int            `json:"maxAge"`          // 年龄上限
	FeeMode         *int            `json:"feeMode"`         // 0免费 1AA 2组织者全包 3人均固定预算
	BudgetPerPerson *int            `json:"budgetPerPerson"` // 人均预算
	OfficialPrice   *float64        `json:"officialPrice"`   // 官方活动定价
	FeeInclude      *string         `json:"feeInclude"`      // 费用包含
	FeeExclude      *string         `json:"feeExclude"`      // 费用不含
	EstTotal        *int            `json:"estTotal"`        // 预估总价
	Visibility      *int            `json:"visibility"`      // 0全部可见 1同城可见 2好友可见
	JoinMode        *int            `json:"joinMode"`        // 0需审核 1直接加入
	AutoClose       *int            `json:"autoClose"`       // 满员自动关闭：0否 1是
	AllowShare      *int            `json:"allowShare"`      // 允许转发：0否 1是
	AllowCollect    *int            `json:"allowCollect"`    // 允许收藏：0否 1是
	IsDraft         *int            `json:"isDraft"`         // 0已发布 1草稿
	IsPublic        *int            `json:"isPublic"`        // 0仅自己可见 1公开招募
	TripID          *string         `json:"tripId"`          // 关联行程ID（可选，已有行程时直接关联）
	Days            []model.TripDay `json:"days"`            // 行程安排（可选，传入则全量替换关联行程）
}

// UpdatePartner 更新搭子信息（仅作者，支持全量替换关联行程安排）
// @Summary 更新搭子
// @Security BearerAuth
// @Tags 小程序-搭子
// @Param id path string true "搭子ID"
// @Param body body UpdatePartnerReq true "更新数据"
// @Success 200 {object} response.Response
// @Router /api/v1/partner/{id} [put]
func UpdatePartner(c *gin.Context) {
	id := c.Param("id")
	partner, err := repository.GetPartnerByID(id)
	if err != nil {
		response.Fail(c, 404, "搭子不存在")
		return
	}
	if partner.UserID != c.MustGet("userID").(string) {
		response.Fail(c, 403, "无权限")
		return
	}
	var req UpdatePartnerReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 400, "参数错误")
		return
	}

	// 基础字段更新（指针非空才更新）
	if req.Title != nil {
		partner.Title = *req.Title
	}
	if req.Cover != nil {
		partner.Cover = *req.Cover
	}
	if req.Images != nil {
		partner.Images = req.Images
	}
	if req.Category != nil {
		partner.Category = *req.Category
	}
	if req.Destination != nil {
		partner.Destination = *req.Destination
	}
	if req.Longitude != nil {
		partner.Longitude = *req.Longitude
	}
	if req.Latitude != nil {
		partner.Latitude = *req.Latitude
	}
	if req.Address != nil {
		partner.Address = *req.Address
	}
	if req.LocationType != nil {
		partner.LocationType = *req.LocationType
	}
	if req.OnlineLink != nil {
		partner.OnlineLink = *req.OnlineLink
	}
	if req.StartDate != nil {
		partner.StartDate = parseDate(*req.StartDate)
	}
	if req.EndDate != nil {
		partner.EndDate = parseDate(*req.EndDate)
	}
	if req.TravelTags != nil {
		partner.TravelTags = *req.TravelTags
	}
	if req.Tags != nil {
		partner.Tags = *req.Tags
	}
	if req.Desc != nil {
		partner.Desc = *req.Desc
	}
	if req.RichDesc != nil {
		partner.RichDesc = *req.RichDesc
	}
	if req.Requirement != nil {
		partner.Requirement = *req.Requirement
	}
	if req.MaxMembers != nil {
		partner.MaxMembers = *req.MaxMembers
	}
	if req.MinMembers != nil {
		partner.MinMembers = *req.MinMembers
	}
	if req.GenderLimit != nil {
		partner.GenderLimit = *req.GenderLimit
	}
	if req.MaleCount != nil {
		partner.MaleCount = *req.MaleCount
	}
	if req.FemaleCount != nil {
		partner.FemaleCount = *req.FemaleCount
	}
	if req.MinAge != nil {
		partner.MinAge = *req.MinAge
	}
	if req.MaxAge != nil {
		partner.MaxAge = *req.MaxAge
	}
	if req.FeeMode != nil {
		partner.FeeMode = *req.FeeMode
	}
	if req.BudgetPerPerson != nil {
		partner.BudgetPerPerson = *req.BudgetPerPerson
	}
	if req.OfficialPrice != nil {
		partner.OfficialPrice = *req.OfficialPrice
	}
	if req.FeeInclude != nil {
		partner.FeeInclude = *req.FeeInclude
	}
	if req.FeeExclude != nil {
		partner.FeeExclude = *req.FeeExclude
	}
	if req.EstTotal != nil {
		partner.EstTotal = *req.EstTotal
	}
	if req.Visibility != nil {
		partner.Visibility = *req.Visibility
	}
	if req.JoinMode != nil {
		partner.JoinMode = *req.JoinMode
	}
	if req.AutoClose != nil {
		partner.AutoClose = *req.AutoClose
	}
	if req.AllowShare != nil {
		partner.AllowShare = *req.AllowShare
	}
	if req.AllowCollect != nil {
		partner.AllowCollect = *req.AllowCollect
	}
	if req.IsDraft != nil {
		partner.IsDraft = *req.IsDraft
	}
	if req.IsPublic != nil {
		partner.IsPublic = *req.IsPublic
	}
	// 草稿转正式发布：重置为招募中（防止草稿期被定时任务置为已过期后发布仍显示过期）
	if req.IsDraft != nil && *req.IsDraft == 0 && partner.Status == 3 {
		partner.Status = 0
	}
	if req.TripID != nil && *req.TripID != "" {
		partner.TripID = *req.TripID
	}

	if err := repository.UpdatePartner(partner); err != nil {
		response.Fail(c, 500, "更新失败")
		return
	}

	// 行程安排：全量替换搭子行程日列表（关联表 partner_days，编辑后详情页展示最新行程）
	if len(req.Days) > 0 {
		if err := repository.UpdatePartnerWithDays(partner.ID, nil, toPartnerDays(req.Days)); err != nil {
			log.Printf("搭子行程日列表更新失败: %v", err)
		}
	}
	response.Success(c, nil)
}

// GetPartnerDetail 获取搭子详情
// @Summary 搭子详情
// @Security BearerAuth
// @Tags 小程序-搭子
// @Param id path string true "搭子ID"
// @Success 200 {object} response.Response{data=object{id=string,userId=string,type=int,category=string,title=string,cover=string,images=string,destination=string,longitude=float64,latitude=float64,address=string,locationType=int,onlineLink=string,startDate=string,endDate=string,days=int,travelTags=string,tags=string,desc=string,richDesc=string,requirement=string,maxMembers=int,minMembers=int,currentMembers=int,genderLimit=int,maleCount=int,femaleCount=int,minAge=int,maxAge=int,feeMode=int,budgetPerPerson=int,officialPrice=float64,feeInclude=string,feeExclude=string,estTotal=int,visibility=int,joinMode=int,autoClose=int,allowShare=int,allowCollect=int,isDraft=int,status=int,isPublic=int,viewCount=int,likeCount=int,favoriteCount=int,commentCount=int,sortWeight=int,createdAt=string,updatedAt=string,authorName=string,authorAvatar=string,isFollowed=bool,isSelf=bool,isApplied=bool,application=object{id=string,status=int,remark=string,reason=string},isLiked=bool,isFavorited=bool,trip=model.Trip}}
// @Router /api/v1/partner/{id} [get]
func GetPartnerDetail(c *gin.Context) {
	id := c.Param("id")
	partner, err := repository.GetPartnerByID(id)
	if err != nil {
		response.Fail(c, 500, "搭子不存在")
		return
	}
	userID := c.MustGet("userID").(string)

	// 后台已下架的内容不可访问（作者本人可见）
	if partner.Status == 4 && partner.UserID != userID {
		response.Fail(c, 404, "搭子已下架")
		return
	}

	// 增加浏览量（非作者）
	if partner.UserID != userID {
		_ = repository.IncrementPartnerViewCount(id)
	}

	// 作者信息
	author, _ := repository.GetUserByID(partner.UserID)
	authorName := ""
	authorAvatar := ""
	if author != nil {
		authorName = author.Nickname
		authorAvatar = author.AvatarURL
	}

	// 关注状态
	followStatus, _ := repository.GetFollowStatus(userID, partner.UserID)
	isFollowed := followStatus == 1 || followStatus == 2

	// 当前用户是否已申请
	appliedMap, _ := repository.GetUserAppliedPartnerIDs(userID, []string{id})
	isApplied := appliedMap[id]

	// 当前用户对该搭子的申请记录（含状态和拒绝理由，未申请为null）
	myApp, _ := repository.GetMyApplication(userID, id)
	var application interface{}
	if myApp != nil {
		application = gin.H{
			"id":     myApp.ID,
			"status": myApp.Status,
			"remark": myApp.Remark,
			"reason": myApp.Reason,
		}
	}

	// 点赞/收藏状态（点赞记录在 partner_likes 表，收藏记录在 Favorite 表，两者独立）
	isLiked := repository.IsPartnerLiked(userID, id)
	isFavorited := repository.IsFavorited(userID, id, "partner")

	// 行程安排：搭子自身行程日列表（关联表 partner_days，由 GetPartnerByID 预加载）

	// 实时查询评论数
	cm := repository.GetPartnerCommentCounts([]string{id})
	if cnt, ok := cm[id]; ok {
		partner.CommentCount = int(cnt)
	}

	// 多图：保证返回数组（无图时返回空数组而非 null）
	images := partner.Images
	if images == nil {
		images = []string{}
	}

	response.Success(c, gin.H{
		"id":              partner.ID,
		"userId":          partner.UserID,
		"type":            partner.Type,
		"category":        partner.Category,
		"title":           partner.Title,
		"cover":           partner.Cover,
		"images":          images,
		"destination":     partner.Destination,
		"longitude":       partner.Longitude,
		"latitude":        partner.Latitude,
		"address":         partner.Address,
		"locationType":    partner.LocationType,
		"onlineLink":      partner.OnlineLink,
		"startDate":       partner.StartDate,
		"endDate":         partner.EndDate,
		"travelTags":      partner.TravelTags,
		"tags":            partner.Tags,
		"desc":            partner.Desc,
		"richDesc":        partner.RichDesc,
		"requirement":     partner.Requirement,
		"maxMembers":      partner.MaxMembers,
		"minMembers":      partner.MinMembers,
		"currentMembers":  partner.CurrentMembers,
		"genderLimit":     partner.GenderLimit,
		"maleCount":       partner.MaleCount,
		"femaleCount":     partner.FemaleCount,
		"minAge":          partner.MinAge,
		"maxAge":          partner.MaxAge,
		"feeMode":         partner.FeeMode,
		"budgetPerPerson": partner.BudgetPerPerson,
		"officialPrice":   partner.OfficialPrice,
		"feeInclude":      partner.FeeInclude,
		"feeExclude":      partner.FeeExclude,
		"estTotal":        partner.EstTotal,
		"visibility":      partner.Visibility,
		"joinMode":        partner.JoinMode,
		"autoClose":       partner.AutoClose,
		"allowShare":      partner.AllowShare,
		"allowCollect":    partner.AllowCollect,
		"isDraft":         partner.IsDraft,
		"status":          partner.Status,
		"cancelReason":    partner.CancelReason,
		"isPublic":        partner.IsPublic,
		"viewCount":       partner.ViewCount,
		"likeCount":       partner.LikeCount,
		"favoriteCount":   partner.FavoriteCount,
		"commentCount":    partner.CommentCount,
		"sortWeight":      partner.SortWeight,
		"createdAt":       partner.CreatedAt,
		"updatedAt":       partner.UpdatedAt,
		"authorName":      authorName,
		"authorAvatar":    authorAvatar,
		"isFollowed":      isFollowed,
		"isSelf":          userID == partner.UserID,
		"isApplied":       isApplied,
		"application":     application,
		"isLiked":         isLiked,
		"isFavorited":     isFavorited,
		"days":            partner.Days,
		"tripId":          partner.TripID,
	})
}

// ApplyPartner 申请加入某个搭子
// @Summary 申请加入搭子
// @Security BearerAuth
// @Tags 小程序-搭子
// @Param id path string true "搭子ID"
// @Param body body object{remark=string} true "申请留言"
// @Success 200 {object} response.Response
// @Router /api/v1/partner/{id}/apply [post]
func ApplyPartner(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		Remark string `json:"remark"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 400, "参数错误")
		return
	}
	// 校验搭子状态：仅招募中可申请（满员仍可申请作为候补）
	partner, err := repository.GetPartnerByID(id)
	if err != nil {
		response.Fail(c, 404, "搭子不存在")
		return
	}
	if partner.Status != 0 {
		response.Fail(c, 400, "搭子已结束招募，无法申请")
		return
	}
	app := model.PartnerApplication{
		PartnerID: id,
		UserID:    c.MustGet("userID").(string),
		Remark:    req.Remark,
	}
	if err := repository.CreateApplication(&app); err != nil {
		response.Fail(c, 500, "申请失败")
		return
	}
	// 通知搭子发起人
	if partner != nil && partner.UserID != app.UserID {
		if err := repository.CreateNotification(&model.Notification{
			UserID:     partner.UserID,
			FromUserID: app.UserID,
			Type:       1,
			RelatedID:  app.ID,
			Content:    "您的搭子收到一条新申请",
		}); err != nil {
			// 通知创建失败不影响主流程
			log.Printf("创建搭子申请通知失败: %v", err)
		} else {
			// 实时推送：作者在线时立即刷新消息中心
			ws.WsHub.PushToUser(partner.UserID, map[string]interface{}{
				"action":  "new_notification",
				"type":    1, // 搭子申请
				"title":   "新的搭子申请",
				"content": "您的搭子收到一条新申请",
			})
		}
	}
	response.Success(c, nil)
}

// HandleApplication 处理搭子申请（同意/拒绝）
// @Summary 处理搭子申请
// @Security BearerAuth
// @Tags 小程序-搭子
// @Param id path string true "搭子ID"
// @Param body body object{applicationId=string,status=int,reason=string} true "申请ID和状态(1同意 2拒绝)，reason拒绝理由（拒绝时填写）"
// @Success 200 {object} response.Response
// @Router /api/v1/partner/{id}/application [put]
func HandleApplication(c *gin.Context) {
	partnerID := c.Param("id")
	var req struct {
		ApplicationID string `json:"applicationId"`
		Status        int    `json:"status"` // 1同意 2拒绝
		Reason        string `json:"reason"` // 拒绝理由（拒绝时填写）
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 400, "参数错误")
		return
	}
	if req.Status != 1 && req.Status != 2 {
		response.Fail(c, 400, "状态无效")
		return
	}

	// 验证是否为搭子发起人
	partner, err := repository.GetPartnerByID(partnerID)
	if err != nil {
		response.Fail(c, 404, "搭子不存在")
		return
	}
	if partner.UserID != c.MustGet("userID").(string) {
		response.Fail(c, 403, "无权限处理该搭子的申请")
		return
	}

	// 处理申请
	if req.Status == 1 {
		if err := service.ApproveApplication(req.ApplicationID); err != nil {
			response.Fail(c, 500, err.Error())
			return
		}
	} else {
		if err := service.RejectApplication(req.ApplicationID, req.Reason); err != nil {
			response.Fail(c, 500, err.Error())
			return
		}
	}
	response.Success(c, nil)
}

// CancelPartner 发起人解散搭子（可选填写解散原因，自动通知所有已加入成员）
// @Summary 解散搭子
// @Security BearerAuth
// @Tags 小程序-搭子
// @Param id path string true "搭子ID"
// @Param body body object{reason=string} false "解散原因（可选）"
// @Success 200 {object} response.Response
// @Router /api/v1/partner/{id}/cancel [put]
func CancelPartner(c *gin.Context) {
	id := c.Param("id")
	userID := c.MustGet("userID").(string)

	var req struct {
		Reason string `json:"reason"` // 解散原因（可选）
	}
	_ = c.ShouldBindJSON(&req)

	// 验证是否为发起人
	partner, err := repository.GetPartnerByID(id)
	if err != nil {
		response.Fail(c, 404, "搭子不存在")
		return
	}
	if partner.UserID != userID {
		response.Fail(c, 403, "仅发起人可解散搭子")
		return
	}
	if partner.Status != 0 && partner.Status != 1 {
		response.Fail(c, 400, "当前状态不可解散")
		return
	}

	if err := repository.CancelPartner(id, userID, req.Reason); err != nil {
		response.Fail(c, 500, "解散失败")
		return
	}

	// 通知所有已加入成员（解散原因单独存 cancelReason 字段返回）
	content := fmt.Sprintf("您加入的搭子「%s」已解散", partner.Title)
	memberIDs, _ := repository.GetPartnerMemberIDs(id)
	for _, mid := range memberIDs {
		if mid == userID {
			continue
		}
		if err := repository.CreateNotification(&model.Notification{
			UserID:       mid,
			FromUserID:   userID,
			Type:         6,
			RelatedID:    id,
			Content:      content,
			CancelReason: req.Reason,
		}); err != nil {
			log.Printf("解散搭子通知失败: %v", err)
		}
	}
	response.Success(c, nil)
}

// LeavePartner 成员退出搭子（发起人不可退出，只能解散；满员时名额自动补位候补）
// @Summary 退出搭子
// @Security BearerAuth
// @Tags 小程序-搭子
// @Param id path string true "搭子ID"
// @Success 200 {object} response.Response
// @Router /api/v1/partner/{id}/leave [put]
func LeavePartner(c *gin.Context) {
	id := c.Param("id")
	userID := c.MustGet("userID").(string)

	partner, err := repository.GetPartnerByID(id)
	if err != nil {
		response.Fail(c, 404, "搭子不存在")
		return
	}
	if partner.UserID == userID {
		response.Fail(c, 400, "发起人不能退出，只能解散搭子")
		return
	}
	if partner.Status != 0 && partner.Status != 1 {
		response.Fail(c, 400, "当前状态不可退出")
		return
	}
	// 已开始（出发日期已过）不可退出
	if partner.StartDate != nil && partner.StartDate.Before(time.Now()) {
		response.Fail(c, 400, "行程已开始，无法退出")
		return
	}
	// 校验已加入
	if _, err := repository.GetApplicationByPartnerAndUser(id, userID); err != nil {
		response.Fail(c, 400, "您尚未加入该搭子")
		return
	}

	// 满员（状态满员或人数已达上限）时退出 → 名额释放，自动补位最早的候补（补位成功会通知候补用户）
	wasFull := partner.Status == 1 || (partner.MaxMembers > 0 && partner.CurrentMembers >= partner.MaxMembers)
	if err := repository.LeavePartner(id, userID); err != nil {
		response.Fail(c, 500, "退出失败")
		return
	}

	// 满员时退出 → 名额释放，自动补位最早的候补（补位成功会通知候补用户）
	if wasFull {
		if app, err := repository.GetEarliestPendingApplication(id); err == nil {
			if err := service.ApproveApplication(app.ID); err != nil {
				log.Printf("退出后自动补位失败: %v", err)
			}
		}
		// 补位后若再次满员，恢复满员状态
		if after, err := repository.GetPartnerByID(id); err == nil && after.CurrentMembers >= after.MaxMembers {
			_ = repository.UpdatePartnerStatus(id, 1)
		}
	}

	// 通知发起人
	user, _ := repository.GetUserByID(userID)
	nickname := "该用户"
	if user != nil && user.Nickname != "" {
		nickname = user.Nickname
	}
	if err := repository.CreateNotification(&model.Notification{
		UserID:     partner.UserID,
		FromUserID: userID,
		Type:       6,
		RelatedID:  id,
		Content:    fmt.Sprintf("成员「%s」退出了您创建的搭子「%s」", nickname, partner.Title),
	}); err != nil {
		log.Printf("退出搭子通知失败: %v", err)
	}
	response.Success(c, nil)
}

// AIGeneratePartner AI生成搭子招募信息
// @Summary AI生成搭子
// @Description 根据目的地和天数调用AI生成搭子招募文案并自动发布（每日基础1次，邀请好友成功1人可额外+1次，超出返回400）
// @Security BearerAuth
// @Tags 小程序-搭子
// @Param body body object{destination=string,days=int,category=string} true "生成参数：destination=目的地, days=天数, category=分类(可选)"
// @Success 200 {object} response.Response{data=model.Partner}
// @Router /api/v1/partner/ai-generate [post]
func AIGeneratePartner(c *gin.Context) {
	var req struct {
		Destination string `json:"destination" binding:"required"`
		Days        int    `json:"days" binding:"required"`
		Category    string `json:"category"`
		StartDate   string `json:"startDate"` // 用户选择的出发日期（可选，指定后强制采用）
		EndDate     string `json:"endDate"`   // 用户选择的结束日期（可选）
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 400, "参数错误")
		return
	}
	userID := c.MustGet("userID").(string)

	// 统计：记录一次AI生成点击
	aiLog, _ := repository.CreateAiGenerateLog(userID, "partner")

	// 额度校验：管理员不限次数，其他用户今日基础1次 + 邀请成功奖励，超出拒绝
	user, _ := repository.GetUserByID(userID)
	if user == nil || user.Role != 2 {
		inviteCount, _ := repository.CountTodayInviteSuccess(userID)
		partnerUsed, _ := repository.CountTodayAIPartners(userID)
		if int(partnerUsed) >= 1+int(inviteCount) {
			response.Fail(c, 400, "今日AI生成次数已用完，邀请好友可额外获得次数")
			return
		}
	}

	// 出发日期说明（未指定时提示从今天开始）
	startDesc := "未指定（从今天开始）"
	if req.StartDate != "" {
		startDesc = req.StartDate
	}
	prompt := fmt.Sprintf(ai.PartnerPrompt, req.Destination, req.Days, startDesc, req.Days)
	result, err := ai.Chat(prompt)
	if err != nil {
		_ = repository.UpdateAiGenerateLogStatus(aiLog.ID, 2)
		response.Fail(c, 500, "AI生成失败")
		return
	}

	// 解析 AI 返回的搭子信息
	var aiResult struct {
		Title           string   `json:"title"`
		Category        string   `json:"category"`
		Destination     string   `json:"destination"`
		Days            int      `json:"days"`
		Desc            string   `json:"desc"`
		RichDesc        string   `json:"richDesc"`
		Requirement     string   `json:"requirement"`
		Address         string   `json:"address"`
		StartDate       string   `json:"startDate"`
		EndDate         string   `json:"endDate"`
		MaxMembers      int      `json:"maxMembers"`
		MinMembers      int      `json:"minMembers"`
		GenderLimit     int      `json:"genderLimit"`
		MaleCount       int      `json:"maleCount"`
		FemaleCount     int      `json:"femaleCount"`
		MinAge          int      `json:"minAge"`
		MaxAge          int      `json:"maxAge"`
		FeeMode         int      `json:"feeMode"`
		BudgetPerPerson int      `json:"budgetPerPerson"`
		FeeInclude      string   `json:"feeInclude"`
		FeeExclude      string   `json:"feeExclude"`
		EstTotal        int      `json:"estTotal"`
		Tags            []string `json:"tags"`
		Schedule        []struct {
			DayNumber int    `json:"dayNumber"`
			Date      string `json:"date"`
			Title     string `json:"title"`
			Items     []struct {
				SectionType   string `json:"sectionType"`
				Title         string `json:"title"`
				Description   string `json:"description"`
				StartTime     string `json:"startTime"`
				EndTime       string `json:"endTime"`
				Address       string `json:"address"`
				StartPoint    string `json:"startPoint"`
				EndPoint      string `json:"endPoint"`
				TransportMode string `json:"transportMode"`
			} `json:"items"`
		} `json:"schedule"`
	}
	if err := json.Unmarshal([]byte(result), &aiResult); err != nil {
		_ = repository.UpdateAiGenerateLogStatus(aiLog.ID, 2)
		response.Fail(c, 500, "AI返回格式异常")
		return
	}
	// 兜底：AI 未返回的字段用默认值填充
	if aiResult.Title == "" {
		aiResult.Title = fmt.Sprintf("%s%d天组队！", req.Destination, req.Days)
	}
	if aiResult.Destination == "" {
		aiResult.Destination = req.Destination
	}
	if aiResult.Days <= 0 {
		aiResult.Days = req.Days
	}
	if aiResult.Category == "" {
		aiResult.Category = req.Category
		if aiResult.Category == "" {
			aiResult.Category = "旅游"
		}
	}
	if aiResult.MaxMembers <= 0 {
		aiResult.MaxMembers = 4
	}
	if aiResult.MinMembers <= 0 {
		aiResult.MinMembers = 2
	}
	if aiResult.MaxAge <= 0 {
		aiResult.MaxAge = 99
	}
	if aiResult.MinAge <= 0 {
		aiResult.MinAge = 18
	}
	tagsJSON := ""
	if len(aiResult.Tags) > 0 {
		if b, err := json.Marshal(aiResult.Tags); err == nil {
			tagsJSON = string(b)
		}
	}

	// AI 生成不落库（不保存草稿），仅返回生成数据；用户确认发布后通过发布接口新建
	p := model.Partner{
		ID:              fmt.Sprintf("ai_%d", time.Now().UnixNano()), // 临时标识，仅用于前端流程串联，非库中记录ID
		UserID:          userID,
		Title:           aiResult.Title,
		Category:        aiResult.Category,
		Destination:     aiResult.Destination,
		Address:         aiResult.Address,
		StartDate:       parseDate(aiResult.StartDate),
		EndDate:         parseDate(aiResult.EndDate),
		TravelTags:      strings.Join(aiResult.Tags, ","),
		Tags:            tagsJSON,
		Desc:            aiResult.Desc,
		RichDesc:        aiResult.RichDesc,
		Requirement:     aiResult.Requirement,
		MaxMembers:      aiResult.MaxMembers,
		MinMembers:      aiResult.MinMembers,
		GenderLimit:     aiResult.GenderLimit,
		MaleCount:       aiResult.MaleCount,
		FemaleCount:     aiResult.FemaleCount,
		MinAge:          aiResult.MinAge,
		MaxAge:          aiResult.MaxAge,
		FeeMode:         aiResult.FeeMode,
		BudgetPerPerson: aiResult.BudgetPerPerson,
		FeeInclude:      aiResult.FeeInclude,
		FeeExclude:      aiResult.FeeExclude,
		EstTotal:        aiResult.EstTotal,
		IsDraft:         0, // 非草稿（AI生成不保存草稿）
		IsPublic:        1,
		IsAI:            1,
	}
	p.Status = 0         // 默认招募中
	p.CurrentMembers = 1 // 发起人计入
	// 用户指定出发日期时强制采用（AI 返回日期不可靠，防止乱编过去/错误日期）
	if req.StartDate != "" {
		p.StartDate = parseDate(req.StartDate)
		if req.EndDate != "" {
			p.EndDate = parseDate(req.EndDate)
		} else {
			p.EndDate = parseDate(addDaysStr(req.StartDate, aiResult.Days-1))
		}
	}

	// 行程安排：AI 生成的 schedule 转为响应用行程日列表（不落库）
	var tripDays []model.TripDay
	if len(aiResult.Schedule) > 0 {
		dayList := make([]model.TripDay, 0, len(aiResult.Schedule))
		for i, d := range aiResult.Schedule {
			dayNumber := d.DayNumber
			if dayNumber <= 0 {
				dayNumber = i + 1
			}
			items := make([]model.TripItem, 0, len(d.Items))
			for _, it := range d.Items {
				items = append(items, model.TripItem{
					SectionType:   it.SectionType,
					Title:         it.Title,
					Description:   it.Description,
					StartTime:     it.StartTime,
					EndTime:       it.EndTime,
					Address:       it.Address,
					StartPoint:    it.StartPoint,
					EndPoint:      it.EndPoint,
					TransportMode: it.TransportMode,
				})
			}
			// 用户指定出发日期时，每天日期按出发日期顺延（AI 返回的日期不采用）
			dayDate := d.Date
			if req.StartDate != "" {
				dayDate = addDaysStr(req.StartDate, dayNumber-1)
			}
			dayList = append(dayList, model.TripDay{
				DayNumber: dayNumber,
				Date:      dayDate,
				Title:     d.Title,
				Items:     items,
			})
		}
		tripDays = dayList
	}

	// 响应：搭子完整字段 + 行程安排数组（前端 AI 流程依赖 days 数组结构）
	respData := make(map[string]interface{})
	if b, err := json.Marshal(p); err == nil {
		json.Unmarshal(b, &respData)
	}
	respData["coverImage"] = p.Cover
	respData["cities"] = []string{p.Destination}
	respData["dayCount"] = len(tripDays)
	if respData["dayCount"] == 0 {
		respData["dayCount"] = aiResult.Days
	}
	if tripDays == nil {
		tripDays = []model.TripDay{}
	}
	respData["days"] = tripDays
	_ = repository.UpdateAiGenerateLogStatus(aiLog.ID, 1)
	response.Success(c, respData)
}

// LikePartner 点赞搭子
// @Summary 点赞搭子
// @Security BearerAuth
// @Tags 小程序-搭子
// @Param id path string true "搭子ID"
// @Success 200 {object} response.Response
// @Router /api/v1/partner/{id}/like [post]
func LikePartner(c *gin.Context) {
	id := c.Param("id")
	userID := c.MustGet("userID").(string)
	if err := repository.LikePartner(userID, id); err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	// 通知搭子发起人（非本人点赞才通知）+ 实时推送
	partner, _ := repository.GetPartnerByID(id)
	if partner != nil && partner.UserID != userID {
		notification := model.Notification{
			UserID:     partner.UserID,
			FromUserID: userID,
			Type:       2,
			RelatedID:  id,
			Content:    "您的搭子收到一个赞",
		}
		if err := repository.CreateNotification(&notification); err == nil {
			ws.WsHub.PushToUser(partner.UserID, map[string]interface{}{
				"action":  "new_notification",
				"type":    2, // 点赞
				"content": notification.Content,
			})
		}
	}
	response.Success(c, nil)
}

// UnlikePartner 取消点赞搭子
// @Summary 取消点赞搭子
// @Security BearerAuth
// @Tags 小程序-搭子
// @Param id path string true "搭子ID"
// @Success 200 {object} response.Response
// @Router /api/v1/partner/{id}/like [delete]
func UnlikePartner(c *gin.Context) {
	id := c.Param("id")
	userID := c.MustGet("userID").(string)
	if err := repository.UnlikePartner(userID, id); err != nil {
		response.Fail(c, 400, err.Error())
		return
	}
	response.Success(c, nil)
}
