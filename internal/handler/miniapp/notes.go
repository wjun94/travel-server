package miniapp

import (
	"strconv"

	"travel-server/internal/repository"
	"travel-server/pkg/response"

	"github.com/gin-gonic/gin"
)

// GetMyNotes 我的全部笔记（攻略+行程+搭子，合并按时间倒序）
// @Summary 我的全部笔记
// @Security BearerAuth
// @Tags 小程序-笔记
// @Param page query int false "页码"
// @Param pageSize query int false "每页数量"
// @Success 200 {object} response.Response{data=object{list=[]model.FeedItem,total=int}}
// @Router /api/v1/my/notes [get]
func GetMyNotes(c *gin.Context) {
	userID := c.MustGet("userID").(string)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))

	items, total, err := repository.ListMyNotes(userID, page, pageSize)
	if err != nil {
		response.Fail(c, 500, "获取失败")
		return
	}
	response.Success(c, gin.H{"list": items, "total": total})
}
