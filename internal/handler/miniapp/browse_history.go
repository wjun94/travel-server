package miniapp

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"travel-server/internal/model"
	"travel-server/internal/repository"
	"travel-server/pkg/response"
)

// AddBrowseHistory 添加浏览记录
// @Summary 添加浏览记录
// @Security BearerAuth
// @Tags 小程序-浏览历史
// @Param body body object{targetId=string,targetType=string,title=string,coverImage=string} true "浏览记录"
// @Success 200 {object} response.Response
// @Router /api/v1/browse/history [post]
func AddBrowseHistory(c *gin.Context) {
	uid := c.MustGet("userID").(string)
	var req struct {
		TargetID   string `json:"targetId" binding:"required"`
		TargetType string `json:"targetType"` // 默认guide
		Title      string `json:"title"`
		CoverImage string `json:"coverImage"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 400, "参数错误")
		return
	}
	if req.TargetType == "" {
		req.TargetType = "guide"
	}
	bh := model.BrowseHistory{
		UserID:     uid,
		TargetID:   req.TargetID,
		TargetType: req.TargetType,
		Title:      req.Title,
		CoverImage: req.CoverImage,
	}
	if err := repository.AddBrowseHistory(&bh); err != nil {
		response.Fail(c, 500, "添加失败")
		return
	}
	response.Success(c, nil)
}

// GetBrowseHistory 获取浏览历史
// @Summary 获取浏览历史
// @Security BearerAuth
// @Tags 小程序-浏览历史
// @Param page query int false "页码"
// @Param pageSize query int false "每页数量"
// @Success 200 {object} response.Response{data=object{list=[]model.BrowseHistory,total=int}}
// @Router /api/v1/browse/history [get]
func GetBrowseHistory(c *gin.Context) {
	uid := c.MustGet("userID").(string)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	list, total, err := repository.GetBrowseHistory(uid, page, pageSize)
	if err != nil {
		response.Fail(c, 500, "获取失败")
		return
	}
	response.Success(c, gin.H{"list": list, "total": total})
}
