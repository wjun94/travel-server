package miniapp

import (
	"strconv"

	"travel-server/internal/model"
	"travel-server/internal/repository"
	"travel-server/pkg/response"

	"github.com/gin-gonic/gin"
)

// CommentVO 评论视图（含用户信息）
type CommentVO struct {
	model.Comment
	Nickname  string `json:"nickname"`
	AvatarURL string `json:"avatarUrl"`
}

// toCommentVO 将 Comment 转为 CommentVO
func toCommentVO(c model.Comment) CommentVO {
	return CommentVO{
		Comment:   c,
		Nickname:  c.User.Nickname,
		AvatarURL: c.User.AvatarURL,
	}
}

// CreateComment 发表评论
// @Summary 发表评论
// @Security BearerAuth
// @Tags 小程序-评论
// @Param body body object{targetType=string,targetId=string,parentId=string,content=string} true "评论内容"
// @Success 200 {object} response.Response
// @Router /api/v1/comment [post]
func CreateComment(c *gin.Context) {
	var req struct {
		TargetType string  `json:"targetType" binding:"required"`
		TargetID   string  `json:"targetId" binding:"required"`
		ParentID   *string `json:"parentId"`
		Content    string  `json:"content" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 400, "参数错误")
		return
	}
	comment := model.Comment{
		UserID:     c.MustGet("userID").(string),
		TargetType: req.TargetType,
		TargetID:   req.TargetID,
		ParentID:   req.ParentID,
		Content:    req.Content,
	}
	if err := repository.CreateComment(&comment); err != nil {
		response.Fail(c, 500, "评论失败")
		return
	}
	response.Success(c, comment)
}

// DeleteComment 删除自己的评论
// @Summary 删除评论
// @Security BearerAuth
// @Tags 小程序-评论
// @Param id path string true "评论ID"
// @Success 200 {object} response.Response
// @Router /api/v1/comment/{id} [delete]
func DeleteComment(c *gin.Context) {
	id := c.Param("id")
	uid := c.MustGet("userID").(string)
	if err := repository.DeleteComment(id, uid); err != nil {
		response.Fail(c, 500, "删除失败")
		return
	}
	response.Success(c, nil)
}

// GetComments 获取评论列表
// @Summary 评论列表
// @Tags 小程序-评论
// @Param target_type query string true "目标类型(guide/trip)"
// @Param target_id query string true "目标ID"
// @Param page query int false "页码"
// @Param pageSize query int false "每页数量"
// @Success 200 {object} response.Response{data=object{list=[]CommentVO,total=int}}
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
	list := make([]CommentVO, len(comments))
	for i, c := range comments {
		list[i] = toCommentVO(c)
	}
	response.Success(c, gin.H{"list": list, "total": total})
}

// GetReplies 获取子回复
// @Summary 子回复列表
// @Tags 小程序-评论
// @Param parent_id query string true "父评论ID"
// @Success 200 {object} response.Response{data=[]CommentVO}
// @Router /api/v1/comment/replies [get]
func GetReplies(c *gin.Context) {
	parentID := c.Query("parent_id")
	replies, err := repository.GetRepliesByParentID(parentID)
	if err != nil {
		response.Fail(c, 500, "获取失败")
		return
	}
	list := make([]CommentVO, len(replies))
	for i, r := range replies {
		list[i] = toCommentVO(r)
	}
	response.Success(c, list)
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
