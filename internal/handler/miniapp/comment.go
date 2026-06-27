package miniapp

import (
	"strconv"
	"travel-server/internal/model"
	"travel-server/internal/repository"
	"travel-server/pkg/response"

	"github.com/gin-gonic/gin"
)

// CreateComment 发表评论
// @Summary 发表评论
// @Security BearerAuth
// @Tags 小程序-评论
// @Param body body model.Comment true "评论内容"
// @Success 200 {object} response.Response
// @Router /api/v1/comment [post]
func CreateComment(c *gin.Context) {
	var comment model.Comment
	if err := c.ShouldBindJSON(&comment); err != nil {
		response.Fail(c, 400, "参数错误")
		return
	}
	comment.UserID = c.MustGet("userID").(string)
	if err := repository.CreateComment(&comment); err != nil {
		response.Fail(c, 500, "评论失败")
		return
	}
	response.Success(c, comment)
}

// GetComments 获取评论列表
// @Summary 评论列表
// @Tags 小程序-评论
// @Param target_type query string true "目标类型(guide/trip)"
// @Param target_id query string true "目标ID"
// @Param page query int false "页码"
// @Param pageSize query int false "每页数量"
// @Success 200 {object} response.Response{data=object{list=[]model.Comment,total=int}}
// @Router /api/v1/comments [get]
func GetComments(c *gin.Context) {
	targetType := c.Query("target_type")
	targetID := c.Query("target_id")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	if targetType == "" || targetID == "" {
		response.Fail(c, 400, "参数错误")
		return
	}
	comments, total, err := repository.GetCommentsByTarget(targetType, targetID, page, pageSize)
	if err != nil {
		response.Fail(c, 500, "获取失败")
		return
	}
	response.Success(c, gin.H{"list": comments, "total": total})
}

// GetReplies 获取子回复
// @Summary 子回复列表
// @Tags 小程序-评论
// @Param parent_id query string true "父评论ID"
// @Success 200 {object} response.Response{data=[]model.Comment}
// @Router /api/v1/comment/replies [get]
func GetReplies(c *gin.Context) {
	parentID := c.Query("parent_id")
	replies, err := repository.GetRepliesByParentID(parentID)
	if err != nil {
		response.Fail(c, 500, "获取失败")
		return
	}
	response.Success(c, replies)
}

// LikeComment 点赞评论
// @Summary 点赞评论
// @Security BearerAuth
// @Tags 小程序-评论
// @Param id path string true "评论ID"
// @Success 200 {object} response.Response
// @Router /api/v1/comment/{id}/like [post]
func LikeComment(c *gin.Context) {
	id := c.Param("id")
	if err := repository.IncrementCommentLikeCount(id); err != nil {
		response.Fail(c, 500, "点赞失败")
		return
	}
	response.Success(c, nil)
}
