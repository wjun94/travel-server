package admin

import (
	"strconv"

	_ "travel-server/internal/model" // swagger 类型引用需要
	"travel-server/internal/repository"
	"travel-server/pkg/response"

	"github.com/gin-gonic/gin"
)

// ListGuides 后台攻略列表（含审核状态，支持状态筛选）
// @Summary 攻略列表
// @Security BearerAuth
// @Tags 后台-内容
// @Param page query int false "页码"
// @Param pageSize query int false "每页数量"
// @Param status query int false "状态(0草稿 1已发布 2下架，-1或不传为全部)" default(-1)
// @Success 200 {object} response.Response{data=object{list=[]model.Guide,total=int}}
// @Router /api/v1/admin/guides [get]
func ListGuides(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	status, _ := strconv.Atoi(c.DefaultQuery("status", "-1"))
	guides, total, err := repository.ListGuides(page, pageSize, status)
	if err != nil {
		response.Fail(c, 500, "获取攻略列表失败")
		return
	}
	response.Success(c, gin.H{"list": guides, "total": total})
}

// UpdateGuideStatus 审核攻略（发布/下架）
// @Summary 审核攻略
// @Security BearerAuth
// @Tags 后台-内容
// @Param id path string true "攻略ID"
// @Param body body object{status=int} true "状态(1已发布 2下架)"
// @Success 200 {object} response.Response
// @Router /api/v1/admin/guide/{id}/status [put]
func UpdateGuideStatus(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		Status int `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 400, "参数错误")
		return
	}
	if req.Status != 1 && req.Status != 2 {
		response.Fail(c, 400, "状态值无效")
		return
	}
	if err := repository.UpdateGuideStatus(id, req.Status); err != nil {
		response.Fail(c, 500, "更新失败")
		return
	}
	response.Success(c, nil)
}

// GetGuideDetail 攻略详情（含每日行程、作者信息与统计）
// @Summary 攻略详情
// @Security BearerAuth
// @Tags 后台-内容
// @Param id path string true "攻略ID"
// @Success 200 {object} response.Response{data=object{guide=model.Guide,days=[]model.GuideSection,favoriteCount=int,commentCount=int,authorName=string,authorAvatar=string}}
// @Router /api/v1/admin/guide/{id} [get]
func GetGuideDetail(c *gin.Context) {
	id := c.Param("id")
	guide, err := repository.GetGuideByID(id)
	if err != nil {
		response.Fail(c, 404, "攻略不存在")
		return
	}
	days, _ := repository.GetDaysByGuideID(id)
	// 作者信息与统计
	authorName, authorAvatar := "", ""
	if author, err := repository.GetUserByID(guide.UserID); err == nil && author != nil {
		authorName = author.Nickname
		authorAvatar = author.AvatarURL
	}
	response.Success(c, gin.H{
		"guide":         guide,
		"days":          days,
		"favoriteCount": repository.GetGuideFavoriteCount(id),
		"commentCount":  repository.GetGuideCommentCount(id),
		"authorName":    authorName,
		"authorAvatar":  authorAvatar,
	})
}
