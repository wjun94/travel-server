package miniapp

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"travel-server/internal/model"
	"travel-server/internal/repository"
	"travel-server/pkg/response"
)

// validChecklistTargetTypes 备忘清单可关联的目标类型（空=不关联）
var validChecklistTargetTypes = map[string]bool{"": true, "trip": true, "guide": true, "partner": true}

// GetChecklists 获取用户的备忘清单
// @Summary 获取备忘清单
// @Security BearerAuth
// @Tags 小程序-备忘
// @Param page query int false "页码"
// @Param pageSize query int false "每页数量"
// @Success 200 {object} response.Response{data=object{list=[]model.Checklist,total=int}}
// @Router /api/v1/checklist [get]
func GetChecklists(c *gin.Context) {
	uid := c.MustGet("userID").(string)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	lists, total, err := repository.GetChecklistsByUser(uid, page, pageSize)
	if err != nil {
		response.Fail(c, 500, "获取失败")
		return
	}
	response.Success(c, gin.H{"list": lists, "total": total})
}

// CreateChecklist 创建新的备忘清单（可关联行程/攻略/搭子）
// @Summary 创建备忘清单
// @Security BearerAuth
// @Tags 小程序-备忘
// @Param body body object{name=string,targetType=string,targetId=string,tripId=string,items=[]model.ChecklistItem} true "清单信息"
// @Success 200 {object} response.Response
// @Router /api/v1/checklist [post]
func CreateChecklist(c *gin.Context) {
	var req struct {
		Name       string                `json:"name"`
		TargetType string                `json:"targetType"` // trip行程 guide攻略 partner搭子（空=不关联）
		TargetID   string                `json:"targetId"`
		TripID     string                `json:"tripId"` // 兼容旧字段
		Items      []model.ChecklistItem `json:"items"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 400, "参数错误")
		return
	}
	if !validChecklistTargetTypes[req.TargetType] {
		response.Fail(c, 400, "参数错误")
		return
	}
	// 旧字段兼容：只传 tripId 时按行程关联处理
	if req.TargetType == "" && req.TripID != "" {
		req.TargetType, req.TargetID = "trip", req.TripID
	}
	// 有关联时目标 ID 必填
	if req.TargetType != "" && req.TargetID == "" {
		response.Fail(c, 400, "参数错误")
		return
	}
	cl := model.Checklist{
		UserID:     c.MustGet("userID").(string),
		Name:       req.Name,
		TargetType: req.TargetType,
		TargetID:   req.TargetID,
		Items:      req.Items,
	}
	// trip 类型同步写兼容字段 trip_id
	if req.TargetType == "trip" {
		cl.TripID = req.TargetID
	}
	if err := repository.CreateChecklist(&cl); err != nil {
		response.Fail(c, 500, "创建失败")
		return
	}
	response.Success(c, cl)
}

// UpdateChecklistItem 更新清单条目的勾选状态
// @Summary 更新清单条目勾选状态
// @Security BearerAuth
// @Tags 小程序-备忘
// @Param id path string true "清单条目ID"
// @Param body body object{checked=int} true "勾选状态(0/1)"
// @Success 200 {object} response.Response
// @Router /api/v1/checklist/{id}/item [put]
func UpdateChecklistItem(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		Checked int `json:"checked"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 400, "参数错误")
		return
	}
	if err := repository.UpdateChecklistItem(id, req.Checked); err != nil {
		response.Fail(c, 500, "更新失败")
		return
	}
	response.Success(c, nil)
}

// GetChecklistDetail 获取备忘清单详情
// @Summary 获取清单详情
// @Security BearerAuth
// @Tags 小程序-备忘
// @Param id path string true "清单ID"
// @Success 200 {object} response.Response{data=model.Checklist}
// @Router /api/v1/checklist/{id} [get]
func GetChecklistDetail(c *gin.Context) {
	id := c.Param("id")
	uid := c.MustGet("userID").(string)
	cl, err := repository.GetChecklistDetail(id, uid)
	if err != nil {
		response.Fail(c, 500, "获取失败")
		return
	}
	response.Success(c, cl)
}

// UpdateChecklist 更新备忘清单（名称 + 关联 + 条目）
// @Summary 更新备忘清单
// @Security BearerAuth
// @Tags 小程序-备忘
// @Param id path string true "清单ID"
// @Param body body object{name=string,targetType=string,targetId=string,items=[]model.ChecklistItem} true "名称、关联和条目"
// @Success 200 {object} response.Response
// @Router /api/v1/checklist/{id} [put]
func UpdateChecklist(c *gin.Context) {
	id := c.Param("id")
	uid := c.MustGet("userID").(string)
	// 指针接收关联字段：nil=不修改，非nil=明确设置（空串即取消关联）
	var req struct {
		Name       string                `json:"name"`
		TargetType *string               `json:"targetType"` // trip行程 guide攻略 partner搭子
		TargetID   *string               `json:"targetId"`
		Items      []model.ChecklistItem `json:"items"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 400, "参数错误")
		return
	}
	// 关联字段必须成对出现
	if (req.TargetType == nil) != (req.TargetID == nil) {
		response.Fail(c, 400, "参数错误")
		return
	}
	hasTarget := req.TargetType != nil
	targetType, targetID := "", ""
	if hasTarget {
		targetType = *req.TargetType
		if req.TargetID != nil {
			targetID = *req.TargetID
		}
		if !validChecklistTargetTypes[targetType] {
			response.Fail(c, 400, "参数错误")
			return
		}
		// 设置了类型但未给目标 ID（取消关联除外）
		if targetType != "" && targetID == "" {
			response.Fail(c, 400, "参数错误")
			return
		}
	}
	if err := repository.UpdateChecklist(id, uid, req.Name, targetType, targetID, hasTarget, req.Items); err != nil {
		response.Fail(c, 500, "更新失败")
		return
	}
	response.Success(c, nil)
}

// GetChecklistCategories 获取系统预置的备忘清单分类
// @Summary 获取系统预置分类
// @Tags 小程序-备忘
// @Success 200 {object} response.Response{data=[]model.ChecklistCategory}
// @Router /api/v1/checklist/categories [get]
func GetChecklistCategories(c *gin.Context) {
	cats, err := repository.GetChecklistCategories()
	if err != nil {
		response.Fail(c, 500, "获取失败")
		return
	}
	response.Success(c, cats)
}

// DeleteChecklist 删除备忘清单
// @Summary 删除备忘清单
// @Security BearerAuth
// @Tags 小程序-备忘
// @Param id path string true "清单ID"
// @Success 200 {object} response.Response
// @Router /api/v1/checklist/{id} [delete]
func DeleteChecklist(c *gin.Context) {
	id := c.Param("id")
	uid := c.MustGet("userID").(string)
	if err := repository.DeleteChecklist(id, uid); err != nil {
		response.Fail(c, 500, "删除失败")
		return
	}
	response.Success(c, nil)
}
