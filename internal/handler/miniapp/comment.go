package miniapp

import (
	"strconv"

	"travel-server/internal/model"
	"travel-server/internal/repository"
	"travel-server/pkg/database"
	"travel-server/pkg/response"

	"github.com/gin-gonic/gin"
)

// CommentVO 评论视图（含用户信息）
type CommentVO struct {
	model.Comment
	Nickname        string `json:"nickname"`
	AvatarURL       string `json:"avatarUrl"`
	ReplyCount      int64  `json:"replyCount"`      // 子回复数量
	ReplyToNickname string `json:"replyToNickname"` // 被回复人昵称（仅回复列表有效）
	IsAuthor        bool   `json:"isAuthor"`        // 是否是发帖人
}

// toCommentVO 将 Comment 转为 CommentVO
func toCommentVO(c model.Comment, replyCount int64, replyToNickname string) CommentVO {
	return CommentVO{
		Comment:         c,
		Nickname:        c.User.Nickname,
		AvatarURL:       c.User.AvatarURL,
		ReplyCount:      replyCount,
		ReplyToNickname: replyToNickname,
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

	// 批量查询子回复数
	ids := make([]string, len(comments))
	for i, c := range comments {
		ids[i] = c.ID
	}
	replyCounts := repository.GetReplyCounts(ids)

	list := make([]CommentVO, len(comments))
	for i, c := range comments {
		list[i] = toCommentVO(c, replyCounts[c.ID], "")
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

	// 查发帖人（评论所属的攻略/行程作者）
	var targetAuthorID string
	if len(replies) > 0 {
		r := replies[0]
		switch r.TargetType {
		case "guide":
			var guide model.Guide
			database.DB.Select("user_id").Where("id = ?", r.TargetID).First(&guide)
			targetAuthorID = guide.UserID
		case "trip":
			var trip model.Trip
			database.DB.Select("user_id").Where("id = ?", r.TargetID).First(&trip)
			targetAuthorID = trip.UserID
		}
	}

	// 批量查被回复人的昵称
	parentIDs := make([]string, 0, len(replies))
	for _, r := range replies {
		if r.ParentID != nil {
			parentIDs = append(parentIDs, *r.ParentID)
		}
	}
	replyToMap := make(map[string]string)
	if len(parentIDs) > 0 {
		type parentInfo struct {
			ID     string
			UserID string
		}
		var parents []parentInfo
		database.DB.Model(&model.Comment{}).
			Select("id, user_id").
			Where("id IN ?", parentIDs).
			Find(&parents)

		userIDs := make([]string, len(parents))
		for i, p := range parents {
			userIDs[i] = p.UserID
		}
		var users []model.User
		database.DB.Select("id, nickname").Where("id IN ?", userIDs).Find(&users)

		userNicknames := make(map[string]string)
		for _, u := range users {
			userNicknames[u.ID] = u.Nickname
		}
		for _, p := range parents {
			replyToMap[p.ID] = userNicknames[p.UserID]
		}
	}

	list := make([]CommentVO, len(replies))
	for i, r := range replies {
		nickname := ""
		if r.ParentID != nil {
			nickname = replyToMap[*r.ParentID]
		}
		isAuthor := targetAuthorID != "" && r.UserID == targetAuthorID
		list[i] = toCommentVO(r, 0, nickname)
		list[i].IsAuthor = isAuthor
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
	// 通知评论作者（通过直接查库获取评论所属用户）
	var comment model.Comment
	database.DB.Select("user_id").First(&comment, "id = ?", id)
	userID := c.MustGet("userID").(string)
	if comment.UserID != "" && comment.UserID != userID {
		repository.CreateNotification(&model.Notification{
			UserID:    comment.UserID,
			Type:      2,
			RelatedID: id,
			Content:   "您的评论收到一个赞",
		})
	}
	response.Success(c, nil)
}
