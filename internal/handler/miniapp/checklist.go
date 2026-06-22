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
// @Success 200 {object} response.Response{data=[]model.Checklist}
// @Router /api/v1/checklist [get]
func GetChecklists(c *gin.Context) {
	uid := c.GetUint("userID")
	lists, err := repository.GetChecklistsByUser(uid)
	if err != nil {
		response.Fail(c, 500, "获取失败")
		return
	}
	response.Success(c, lists)
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
	cl.UserID = c.GetUint("userID")
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
// @Param id path int true "清单条目ID"
// @Param body body object{checked=int} true "勾选状态(0/1)"
// @Success 200 {object} response.Response
// @Router /api/v1/checklist/{id}/item [put]
func UpdateChecklistItem(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.Fail(c, 400, "参数错误")
		return
	}
	var req struct {
		Checked int `json:"checked"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 400, "参数错误")
		return
	}
	if err := repository.UpdateChecklistItem(uint(id), req.Checked); err != nil {
		response.Fail(c, 500, "更新失败")
		return
	}
	response.Success(c, nil)
}
