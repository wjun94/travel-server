package miniapp

import (
	"strconv"

	"travel-server/internal/model"
	_ "travel-server/internal/model" // swagger 类型引用需要
	"travel-server/internal/repository"
	"travel-server/pkg/response"

	"github.com/gin-gonic/gin"
)

// GetFeed 获取公开的攻略瀑布流
// @Summary 攻略瀑布流
// @Tags 小程序-攻略
// @Param page query int false "页码"
// @Param pageSize query int false "每页数量"
// @Success 200 {object} response.Response{data=object{list=[]model.Post,total=int}}
// @Router /api/v1/feed [get]
func GetFeed(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	posts, total, err := repository.GetFeedPosts(page, pageSize)
	if err != nil {
		response.Fail(c, 500, "获取失败")
		return
	}
	response.Success(c, gin.H{"list": posts, "total": total})
}

// CreatePost 发布一篇图文攻略
// @Summary 发布攻略
// @Security BearerAuth
// @Tags 小程序-攻略
// @Param body body model.Post true "攻略内容"
// @Success 200 {object} response.Response
// @Router /api/v1/post [post]
func CreatePost(c *gin.Context) {
	var post model.Post
	if err := c.ShouldBindJSON(&post); err != nil {
		response.Fail(c, 400, "参数错误")
		return
	}
	post.UserID = c.GetUint("userID")
	post.Status = 1 // 直接发布，也可设为0待审核
	if err := repository.CreatePost(&post); err != nil {
		response.Fail(c, 500, "发布失败")
		return
	}
	response.Success(c, post)
}

// GetPostDetail 获取攻略详情
// @Summary 攻略详情
// @Security BearerAuth
// @Tags 小程序-攻略
// @Param id path int true "攻略ID"
// @Success 200 {object} response.Response{data=model.Post}
// @Router /api/v1/post/{id} [get]
func GetPostDetail(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.Fail(c, 400, "参数错误")
		return
	}
	post, err := repository.GetPostByID(uint(id))
	if err != nil {
		response.Fail(c, 404, "攻略不存在")
		return
	}
	response.Success(c, post)
}
