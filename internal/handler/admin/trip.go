package admin

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"travel-server/internal/repository"
	"travel-server/pkg/response"
)

// ListTrips 行程列表（全部用户，支持状态/关键词筛选）
// @Summary 行程列表
// @Description 获取所有用户行程，支持分页及状态、关键词筛选，含作者信息与行程项数
// @Security BearerAuth
// @Tags 后台-行程
// @Param page query int false "页码" default(1)
// @Param pageSize query int false "每页数量" default(10)
// @Param status query int false "状态(1草稿 2已发布 3已归档，-1或不传为全部)" default(-1)
// @Param keyword query string false "标题/目的地关键词"
// @Success 200 {object} response.Response{data=object{list=[]object,total=int}}
// @Router /api/v1/admin/trips [get]
func ListTrips(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	status, _ := strconv.Atoi(c.DefaultQuery("status", "-1"))
	keyword := c.Query("keyword")

	trips, total, err := repository.ListAllTrips(page, pageSize, status, keyword)
	if err != nil {
		response.Fail(c, 500, "获取行程列表失败")
		return
	}

	// 批量注入作者信息与行程项数
	userIDs := make([]string, 0, len(trips))
	tripIDs := make([]string, 0, len(trips))
	for _, t := range trips {
		userIDs = append(userIDs, t.UserID)
		tripIDs = append(tripIDs, t.ID)
	}
	users := repository.GetUsersByIDs(userIDs)
	itemCounts := repository.GetTripItemCounts(tripIDs)

	list := make([]gin.H, 0, len(trips))
	for _, t := range trips {
		authorName, authorAvatar := "", ""
		if u, ok := users[t.UserID]; ok && u != nil {
			authorName = u.Nickname
			authorAvatar = u.AvatarURL
		}
		list = append(list, gin.H{
			"id":           t.ID,
			"userId":       t.UserID,
			"title":        t.Title,
			"coverImage":   t.CoverImage,
			"destinations": t.Destinations,
			"totalBudget":  t.TotalBudget,
			"isOverseas":   t.IsOverseas,
			"viewCount":    t.ViewCount,
			"likeCount":    t.LikeCount,
			"status":       t.Status,
			"isPublic":     t.IsPublic,
			"isAI":         t.IsAI,
			"createdAt":    t.CreatedAt,
			"updatedAt":    t.UpdatedAt,
			"authorName":   authorName,
			"authorAvatar": authorAvatar,
			"itemCount":    itemCounts[t.ID],
		})
	}
	response.Success(c, gin.H{"list": list, "total": total})
}

// UpdateTripStatus 审核行程（发布/下架/归档）
// @Summary 审核行程
// @Description 修改行程状态：1草稿(下架) 2已发布 3已归档完结
// @Security BearerAuth
// @Tags 后台-行程
// @Param id path string true "行程ID"
// @Param body body object{status=int} true "状态(1草稿 2已发布 3已归档)"
// @Success 200 {object} response.Response
// @Router /api/v1/admin/trip/{id}/status [put]
func UpdateTripStatus(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		Status int `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 400, "参数错误")
		return
	}
	if req.Status != 1 && req.Status != 2 && req.Status != 3 {
		response.Fail(c, 400, "状态值无效")
		return
	}
	if err := repository.UpdateTripStatus(id, req.Status); err != nil {
		response.Fail(c, 500, "更新失败")
		return
	}
	response.Success(c, nil)
}

// GetTripDetail 行程详情（含行程日与行程项、同行者、作者信息）
// @Summary 行程详情
// @Security BearerAuth
// @Tags 后台-行程
// @Param id path string true "行程ID"
// @Success 200 {object} response.Response{data=object{id=string,userId=string,title=string,coverImage=string,destinations=[]string,totalBudget=float64,isOverseas=int,summary=string,viewCount=int,likeCount=int,favoriteCount=int,status=int,isPublic=int,isAI=int,createdAt=string,updatedAt=string,authorName=string,authorAvatar=string,itemCount=int64,days=[]model.TripDay,members=[]model.TripMember}}
// @Router /api/v1/admin/trip/{id} [get]
func GetTripDetail(c *gin.Context) {
	id := c.Param("id")
	trip, err := repository.GetTripByID(id)
	if err != nil {
		response.Fail(c, 404, "行程不存在")
		return
	}
	// 作者信息与统计（列表同款增强）
	authorName, authorAvatar := "", ""
	if author, err := repository.GetUserByID(trip.UserID); err == nil && author != nil {
		authorName = author.Nickname
		authorAvatar = author.AvatarURL
	}
	itemCount := int64(0)
	if counts := repository.GetTripItemCounts([]string{trip.ID}); counts != nil {
		itemCount = counts[trip.ID]
	}
	destinations := trip.Destinations
	if destinations == nil {
		destinations = []string{}
	}
	response.Success(c, gin.H{
		"id":            trip.ID,
		"userId":        trip.UserID,
		"guideId":       trip.GuideID,
		"title":         trip.Title,
		"coverImage":    trip.CoverImage,
		"destinations":  destinations,
		"totalBudget":   trip.TotalBudget,
		"isOverseas":    trip.IsOverseas,
		"summary":       trip.Summary,
		"viewCount":     trip.ViewCount,
		"likeCount":     trip.LikeCount,
		"favoriteCount": repository.GetTripFavoriteCount(id),
		"status":        trip.Status,
		"isPublic":      trip.IsPublic,
		"isAI":          trip.IsAI,
		"createdAt":     trip.CreatedAt,
		"updatedAt":     trip.UpdatedAt,
		"authorName":    authorName,
		"authorAvatar":  authorAvatar,
		"itemCount":     itemCount,
		"days":          trip.Days,
		"members":       trip.Members,
	})
}
