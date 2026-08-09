package admin

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"travel-server/internal/model"
	"travel-server/internal/repository"
	"travel-server/pkg/response"
)

// CreatePartner 创建官方搭子团（后台发布）
// @Summary 创建官方搭子团
// @Security BearerAuth
// @Tags 后台-搭子
// @Param body body model.Partner true "官方搭子信息"
// @Success 200 {object} response.Response
// @Router /api/v1/admin/partner [post]
func CreatePartner(c *gin.Context) {
	var p model.Partner
	if err := c.ShouldBindJSON(&p); err != nil {
		response.Fail(c, 400, "参数错误")
		return
	}
	p.Type = 1    // 官方活动
	p.UserID = "" // 系统发布
	if err := repository.CreatePartner(&p); err != nil {
		response.Fail(c, 500, "创建失败")
		return
	}
	response.Success(c, p)
}

// ListPartners 搭子列表（分页，支持目的地/状态/类型筛选）
// @Summary 搭子列表
// @Description 获取搭子列表，支持分页及目的地、状态、类型筛选（type -1全部 0用户 1官方）
// @Security BearerAuth
// @Tags 后台-搭子
// @Param page query int false "页码" default(1)
// @Param pageSize query int false "每页数量" default(10)
// @Param destination query string false "目的地关键词"
// @Param status query int false "状态(0招募中 1满员 2已取消 3已过期 4已下架，-1或不传为全部)" default(-1)
// @Param type query int false "类型(-1全部 0用户 1官方，默认1兼容官方搭子页)" default(1)
// @Success 200 {object} response.Response{data=object{list=[]object{id=string,userId=string,type=int,title=string,cover=string,destination=string,status=int,createdAt=string,authorName=string,authorAvatar=string},total=int}}
// @Router /api/v1/admin/partners [get]
func ListPartners(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	destination := c.Query("destination")
	status, _ := strconv.Atoi(c.DefaultQuery("status", "-1"))
	ptype, _ := strconv.Atoi(c.DefaultQuery("type", "1"))

	partners, total, err := repository.GetPartners(page, pageSize, destination, status, ptype)
	if err != nil {
		response.Fail(c, 500, "获取搭子列表失败")
		return
	}
	// 批量注入作者信息
	userIDs := make([]string, 0, len(partners))
	for _, p := range partners {
		userIDs = append(userIDs, p.UserID)
	}
	users := repository.GetUsersByIDs(userIDs)
	list := make([]gin.H, 0, len(partners))
	for _, p := range partners {
		item := gin.H{
			"id":              p.ID,
			"userId":          p.UserID,
			"type":            p.Type,
			"title":           p.Title,
			"cover":           p.Cover,
			"category":        p.Category,
			"destination":     p.Destination,
			"locationType":    p.LocationType,
			"address":         p.Address,
			"onlineLink":      p.OnlineLink,
			"startDate":       p.StartDate,
			"endDate":         p.EndDate,
			"days":            p.Days,
			"travelTags":      p.TravelTags,
			"desc":            p.Desc,
			"requirement":     p.Requirement,
			"maxMembers":      p.MaxMembers,
			"minMembers":      p.MinMembers,
			"currentMembers":  p.CurrentMembers,
			"genderLimit":     p.GenderLimit,
			"minAge":          p.MinAge,
			"maxAge":          p.MaxAge,
			"feeMode":         p.FeeMode,
			"budgetPerPerson": p.BudgetPerPerson,
			"officialPrice":   p.OfficialPrice,
			"feeInclude":      p.FeeInclude,
			"feeExclude":      p.FeeExclude,
			"estTotal":        p.EstTotal,
			"visibility":      p.Visibility,
			"joinMode":        p.JoinMode,
			"isPublic":        p.IsPublic,
			"status":          p.Status,
			"viewCount":       p.ViewCount,
			"likeCount":       p.LikeCount,
			"favoriteCount":   p.FavoriteCount,
			"commentCount":    p.CommentCount,
			"createdAt":       p.CreatedAt,
			"updatedAt":       p.UpdatedAt,
		}
		if u, ok := users[p.UserID]; ok && u != nil {
			item["authorName"] = u.Nickname
			item["authorAvatar"] = u.AvatarURL
		}
		list = append(list, item)
	}
	response.Success(c, gin.H{
		"list":  list,
		"total": total,
	})
}

// UpdatePartnerStatus 审核搭子（下架/恢复招募）
// @Summary 审核搭子
// @Description 修改搭子状态：0恢复招募 1满员 2已取消 3已过期 4已下架（后台审核下架用4）
// @Security BearerAuth
// @Tags 后台-搭子
// @Param id path string true "搭子ID"
// @Param body body object{status=int} true "状态(0恢复招募 4下架)"
// @Success 200 {object} response.Response
// @Router /api/v1/admin/partner/{id}/status [put]
func UpdatePartnerStatus(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		Status int `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 400, "参数错误")
		return
	}
	if req.Status < 0 || req.Status > 4 {
		response.Fail(c, 400, "状态值无效")
		return
	}
	if err := repository.UpdatePartnerStatus(id, req.Status); err != nil {
		response.Fail(c, 500, "更新失败")
		return
	}
	response.Success(c, nil)
}

// GetPartnerDetail 搭子详情（含关联行程安排）
// @Summary 搭子详情
// @Security BearerAuth
// @Tags 后台-搭子
// @Param id path string true "搭子ID"
// @Success 200 {object} response.Response{data=object{partner=model.Partner,days=[]model.TripDay}}
// @Router /api/v1/admin/partner/{id} [get]
func GetPartnerDetail(c *gin.Context) {
	id := c.Param("id")
	partner, err := repository.GetPartnerByID(id)
	if err != nil {
		response.Fail(c, 404, "搭子不存在")
		return
	}
	// 关联行程安排（搭子创建时自动生成草稿行程）
	var days []model.TripDay
	if partner.TripID != "" {
		if trip, err := repository.GetTripByID(partner.TripID); err == nil {
			days = trip.Days
		}
	}
	if days == nil {
		days = []model.TripDay{}
	}
	// 作者信息
	authorName, authorAvatar := "", ""
	if author, err := repository.GetUserByID(partner.UserID); err == nil && author != nil {
		authorName = author.Nickname
		authorAvatar = author.AvatarURL
	}
	response.Success(c, gin.H{
		"partner":      partner,
		"days":         days,
		"authorName":   authorName,
		"authorAvatar": authorAvatar,
	})
}
