package miniapp

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"travel-server/internal/model"
	"travel-server/internal/repository"
	"travel-server/internal/service"
	"travel-server/pkg/response"
)

// CreatePartner 发布搭子信息
// @Summary 发布搭子
// @Security BearerAuth
// @Tags 小程序-搭子
// @Param body body model.Partner true "搭子信息"
// @Success 200 {object} response.Response
// @Router /api/v1/partner [post]
func CreatePartner(c *gin.Context) {
	var p model.Partner
	if err := c.ShouldBindJSON(&p); err != nil {
		response.Fail(c, 400, "参数错误")
		return
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

// GetPartnerList 获取搭子列表
// @Summary 搭子列表
// @Security BearerAuth
// @Tags 小程序-搭子
// @Param page query int false "页码"
// @Param pageSize query int false "每页数量"
// @Success 200 {object} response.Response{data=object{list=[]object{id=string,userId=string,tripId=string,type=int,title=string,cover=string,destination=string,longitude=float64,latitude=float64,startDate=string,endDate=string,days=int,travelTags=string,desc=string,requirement=string,maxMembers=int,currentMembers=int,genderLimit=int,minAge=int,maxAge=int,budgetPerPerson=int,officialPrice=float64,status=int,isPublic=int,viewCount=int,sortWeight=int,createdAt=string,updatedAt=string,isApplied=bool,isSelf=bool},total=int64}}
// @Router /api/v1/partner/list [get]
func GetPartnerList(c *gin.Context) {
	userID := c.MustGet("userID").(string)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	list, total, err := repository.GetPartnerList(page, pageSize)
	if err != nil {
		response.Fail(c, 500, "获取失败")
		return
	}

	// 批量查询当前用户是否已申请
	partnerIDs := make([]string, len(list))
	for i, p := range list {
		partnerIDs[i] = p.ID
	}
	appliedMap, _ := repository.GetUserAppliedPartnerIDs(userID, partnerIDs)

	type partnerVO struct {
		model.Partner
		IsApplied bool `json:"isApplied"`
		IsSelf    bool `json:"isSelf"`
	}
	result := make([]partnerVO, len(list))
	for i, p := range list {
		result[i] = partnerVO{
			Partner:   p,
			IsApplied: appliedMap[p.ID],
			IsSelf:    userID == p.UserID,
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
		"title":           partner.Title,
		"cover":           partner.Cover,
		"destination":     partner.Destination,
		"longitude":       partner.Longitude,
		"latitude":        partner.Latitude,
		"startDate":       partner.StartDate,
		"endDate":         partner.EndDate,
		"days":            partner.Days,
		"travelTags":      partner.TravelTags,
		"desc":            partner.Desc,
		"requirement":     partner.Requirement,
		"maxMembers":      partner.MaxMembers,
		"currentMembers":  partner.CurrentMembers,
		"genderLimit":     partner.GenderLimit,
		"minAge":          partner.MinAge,
		"maxAge":          partner.MaxAge,
		"budgetPerPerson": partner.BudgetPerPerson,
		"officialPrice":   partner.OfficialPrice,
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
