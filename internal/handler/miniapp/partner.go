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
// @Success 200 {object} response.Response
// @Router /api/v1/partner/list [get]
func GetPartnerList(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	list, total, err := repository.GetPartnerList(page, pageSize)
	if err != nil {
		response.Fail(c, 500, "获取失败")
		return
	}
	response.Success(c, gin.H{"list": list, "total": total})
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
