package miniapp

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"travel-server/internal/model"
	"travel-server/internal/repository"
	"travel-server/internal/service"
	"travel-server/pkg/response"
)

// CreatePartnerReq 发布搭子请求
type CreatePartnerReq struct {
	Title           string  `json:"title"`           // 搭子标题
	Cover           string  `json:"cover"`           // 封面图
	Images          string  `json:"images"`          // 多图JSON数组
	Category        string  `json:"category"`        // 活动分类：旅游/美食/运动/学习/探店/看展/桌游
	Destination     string  `json:"destination"`     // 目的地
	Longitude       float64 `json:"longitude"`       // 经度
	Latitude        float64 `json:"latitude"`        // 纬度
	Address         string  `json:"address"`         // 详细地址
	LocationType    int     `json:"locationType"`    // 0线下 1线上
	OnlineLink      string  `json:"onlineLink"`      // 线上链接
	StartDate       string  `json:"startDate"`       // YYYY-MM-DD
	EndDate         string  `json:"endDate"`         // YYYY-MM-DD
	Days            int     `json:"days"`            // 出行天数
	TravelTags      string  `json:"travelTags"`      // 逗号分隔（兼容旧版）
	Tags            string  `json:"tags"`            // 多选标签JSON数组
	Desc            string  `json:"desc"`            // 行程简述
	RichDesc        string  `json:"richDesc"`        // 详细介绍（富文本）
	Requirement     string  `json:"requirement"`     // 人员要求
	MaxMembers      int     `json:"maxMembers"`      // 招募上限
	MinMembers      int     `json:"minMembers"`      // 最小成团人数
	GenderLimit     int     `json:"genderLimit"`     // 0不限 1仅男生 2仅女生
	MaleCount       int     `json:"maleCount"`       // 男生需求数
	FemaleCount     int     `json:"femaleCount"`     // 女生需求数
	MinAge          int     `json:"minAge"`          // 年龄下限
	MaxAge          int     `json:"maxAge"`          // 年龄上限
	FeeMode         int     `json:"feeMode"`         // 0免费 1AA 2组织者全包 3人均固定预算
	BudgetPerPerson int     `json:"budgetPerPerson"` // 人均预算
	OfficialPrice   float64 `json:"officialPrice"`   // 官方活动定价
	FeeInclude      string  `json:"feeInclude"`      // 费用包含
	FeeExclude      string  `json:"feeExclude"`      // 费用不含
	EstTotal        int     `json:"estTotal"`        // 预估总价
	Visibility      int     `json:"visibility"`      // 0全部可见 1同城可见 2好友可见
	JoinMode        int     `json:"joinMode"`        // 0需审核 1直接加入
	AutoClose       int     `json:"autoClose"`       // 满员自动关闭：0否 1是
	AllowShare      int     `json:"allowShare"`      // 允许转发：0否 1是
	AllowCollect    int     `json:"allowCollect"`    // 允许收藏：0否 1是
	IsDraft         int     `json:"isDraft"`         // 0已发布 1草稿
	IsPublic        int     `json:"isPublic"`        // 0仅自己可见 1公开招募
	TripID          string  `json:"tripId"`          // 关联行程ID
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
		Days:            req.Days,
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
		TripID:          req.TripID,
	}
	p.UserID = c.MustGet("userID").(string)
	p.Status = 0         // 默认招募中
	p.CurrentMembers = 1 // 发起人计入
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
// @Success 200 {object} response.Response{data=object{list=[]object{id=string,userId=string,tripId=string,type=int,title=string,cover=string,destination=string,longitude=float64,latitude=float64,startDate=string,endDate=string,days=int,travelTags=string,desc=string,requirement=string,maxMembers=int,currentMembers=int,genderLimit=int,minAge=int,maxAge=int,budgetPerPerson=int,officialPrice=float64,status=int,isPublic=int,viewCount=int,sortWeight=int,createdAt=string,updatedAt=string,authorId=string,authorName=string,authorAvatar=string,isApplied=bool,isSelf=bool,isFollowed=bool},total=int64}}
// @Router /api/v1/partner/list [get]
func GetPartnerList(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	list, total, err := repository.GetPartnerList(page, pageSize)
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
		AuthorID     string `json:"authorId"`
		AuthorName   string `json:"authorName"`
		AuthorAvatar string `json:"authorAvatar"`
		IsApplied    bool   `json:"isApplied"`
		IsSelf       bool   `json:"isSelf"`
		IsFollowed   bool   `json:"isFollowed"`
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
		result[i] = partnerVO{
			Partner:      p,
			AuthorID:     p.UserID,
			AuthorName:   authorName,
			AuthorAvatar: authorAvatar,
			IsApplied:    loggedIn && appliedMap[p.ID],
			IsSelf:       loggedIn && currentUserID == p.UserID,
			IsFollowed:   loggedIn && (followMap[p.UserID] == 1 || followMap[p.UserID] == 2),
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
// @Success 200 {object} response.Response{data=object{list=[]object{id=string,userId=string,tripId=string,type=int,title=string,cover=string,destination=string,longitude=float64,latitude=float64,startDate=string,endDate=string,days=int,travelTags=string,desc=string,requirement=string,maxMembers=int,currentMembers=int,genderLimit=int,minAge=int,maxAge=int,budgetPerPerson=int,officialPrice=float64,status=int,isPublic=int,viewCount=int,sortWeight=int,createdAt=string,updatedAt=string,authorId=string,authorName=string,authorAvatar=string,isApplied=bool,isSelf=bool,isFollowed=bool},total=int64}}
// @Router /api/v1/my/partners [get]
func GetMyPartners(c *gin.Context) {
	userID := c.MustGet("userID").(string)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	list, total, err := repository.GetMyPartners(userID, page, pageSize)
	if err != nil {
		response.Fail(c, 500, "获取失败")
		return
	}

	type partnerVO struct {
		model.Partner
		AuthorID     string `json:"authorId"`
		AuthorName   string `json:"authorName"`
		AuthorAvatar string `json:"authorAvatar"`
		IsApplied    bool   `json:"isApplied"`
		IsSelf       bool   `json:"isSelf"`
		IsFollowed   bool   `json:"isFollowed"`
	}
	result := make([]partnerVO, len(list))
	for i, p := range list {
		result[i] = partnerVO{
			Partner:      p,
			AuthorID:     p.UserID,
			AuthorName:   "",
			AuthorAvatar: "",
			IsApplied:    false,
			IsSelf:       true,
			IsFollowed:   false,
		}
	}

	response.Success(c, gin.H{"list": result, "total": total})
}

// GetPartnerDetail 获取搭子详情
// @Summary 搭子详情
// @Security BearerAuth
// @Tags 小程序-搭子
// @Param id path string true "搭子ID"
// @Success 200 {object} response.Response{data=object{id=string,userId=string,tripId=string,type=int,title=string,cover=string,destination=string,longitude=float64,latitude=float64,startDate=string,endDate=string,days=int,travelTags=string,desc=string,requirement=string,maxMembers=int,currentMembers=int,genderLimit=int,minAge=int,maxAge=int,budgetPerPerson=int,officialPrice=float64,status=int,isPublic=int,viewCount=int,sortWeight=int,createdAt=string,updatedAt=string,authorName=string,authorAvatar=string,isFollowed=bool,isSelf=bool,isApplied=bool}}
// @Router /api/v1/partner/{id} [get]
func GetPartnerDetail(c *gin.Context) {
	id := c.Param("id")
	partner, err := repository.GetPartnerByID(id)
	if err != nil {
		response.Fail(c, 404, "搭子不存在")
		return
	}
	userID := c.MustGet("userID").(string)

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

	response.Success(c, gin.H{
		"id":              partner.ID,
		"userId":          partner.UserID,
		"tripId":          partner.TripID,
		"type":            partner.Type,
		"category":        partner.Category,
		"title":           partner.Title,
		"cover":           partner.Cover,
		"images":          partner.Images,
		"destination":     partner.Destination,
		"longitude":       partner.Longitude,
		"latitude":        partner.Latitude,
		"address":         partner.Address,
		"locationType":    partner.LocationType,
		"onlineLink":      partner.OnlineLink,
		"startDate":       partner.StartDate,
		"endDate":         partner.EndDate,
		"days":            partner.Days,
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
		"isPublic":        partner.IsPublic,
		"viewCount":       partner.ViewCount,
		"sortWeight":      partner.SortWeight,
		"createdAt":       partner.CreatedAt,
		"updatedAt":       partner.UpdatedAt,
		"authorName":      authorName,
		"authorAvatar":    authorAvatar,
		"isFollowed":      isFollowed,
		"isSelf":          userID == partner.UserID,
		"isApplied":       isApplied,
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
	partner, _ := repository.GetPartnerByID(id)
	if partner != nil && partner.UserID != app.UserID {
		repository.CreateNotification(&model.Notification{
			UserID:     partner.UserID,
			FromUserID: app.UserID,
			Type:       1,
			RelatedID:  app.ID,
			Content:    "您的搭子收到一条新申请",
		})
	}
	response.Success(c, nil)
}

// HandleApplication 处理搭子申请（同意/拒绝）
// @Summary 处理搭子申请
// @Security BearerAuth
// @Tags 小程序-搭子
// @Param id path string true "搭子ID"
// @Param body body object{applicationId=string,status=int} true "申请ID和状态(1同意 2拒绝)"
// @Success 200 {object} response.Response
// @Router /api/v1/partner/{id}/application [put]
func HandleApplication(c *gin.Context) {
	partnerID := c.Param("id")
	var req struct {
		ApplicationID string `json:"applicationId"`
		Status        int    `json:"status"` // 1同意 2拒绝
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
		if err := repository.UpdateApplicationStatus(req.ApplicationID, 2); err != nil {
			response.Fail(c, 500, "处理失败")
			return
		}
	}
	response.Success(c, nil)
}

// CancelPartner 发起人取消搭子
// @Summary 取消搭子
// @Security BearerAuth
// @Tags 小程序-搭子
// @Param id path string true "搭子ID"
// @Success 200 {object} response.Response
// @Router /api/v1/partner/{id}/cancel [put]
func CancelPartner(c *gin.Context) {
	id := c.Param("id")
	userID := c.MustGet("userID").(string)

	// 验证是否为发起人
	partner, err := repository.GetPartnerByID(id)
	if err != nil {
		response.Fail(c, 404, "搭子不存在")
		return
	}
	if partner.UserID != userID {
		response.Fail(c, 403, "仅发起人可取消搭子")
		return
	}
	if partner.Status != 0 {
		response.Fail(c, 400, "当前状态不可取消")
		return
	}

	if err := repository.CancelPartner(id, userID); err != nil {
		response.Fail(c, 500, "取消失败")
		return
	}
	response.Success(c, nil)
}
