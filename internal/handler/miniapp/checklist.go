package miniapp

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"travel-server/internal/model"
	"travel-server/internal/repository"
	"travel-server/pkg/response"
)

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

// CreateChecklist 创建新的备忘清单
// @Summary 创建备忘清单
// @Security BearerAuth
// @Tags 小程序-备忘
// @Param body body model.Checklist true "清单信息"
// @Success 200 {object} response.Response
// @Router /api/v1/checklist [post]
func CreateChecklist(c *gin.Context) {
	var cl model.Checklist
	if err := c.ShouldBindJSON(&cl); err != nil {
		response.Fail(c, 400, "参数错误")
		return
	}
	cl.UserID = c.MustGet("userID").(string)
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

// UpdateChecklist 更新备忘清单
// @Summary 更新备忘清单
// @Security BearerAuth
// @Tags 小程序-备忘
// @Param id path string true "清单ID"
// @Param body body object{name=string,items=[]model.ChecklistItem} true "名称和条目"
// @Success 200 {object} response.Response
// @Router /api/v1/checklist/{id} [put]
func UpdateChecklist(c *gin.Context) {
	id := c.Param("id")
	uid := c.MustGet("userID").(string)
	var req struct {
		Name  string                `json:"name"`
		Items []model.ChecklistItem `json:"items"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 400, "参数错误")
		return
	}
	if err := repository.UpdateChecklist(id, uid, req.Name, req.Items); err != nil {
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
