// Package miniapp 提供小程序端记账相关接口
package miniapp

import (
	"time"

	"github.com/gin-gonic/gin"

	"travel-server/internal/model"
	"travel-server/internal/repository"
	"travel-server/pkg/response"
)

// validTargetTypes 记账可绑定的目标类型
var validTargetTypes = map[string]bool{"trip": true, "guide": true, "partner": true}

// GetAccountList 获取指定目标的记账记录
// @Summary 获取账本明细
// @Security BearerAuth
// @Tags 小程序-记账
// @Param targetType query string true "绑定类型：trip行程 guide攻略 partner搭子"
// @Param targetId query string true "绑定目标ID"
// @Success 200 {object} response.Response{data=[]model.Accounting}
// @Router /api/v1/account/list [get]
func GetAccountList(c *gin.Context) {
	targetType := c.Query("targetType")
	targetID := c.Query("targetId")
	if !validTargetTypes[targetType] || targetID == "" {
		response.Fail(c, 400, "参数错误")
		return
	}
	userID := c.MustGet("userID").(string)
	accounts, err := repository.GetAccountsByTarget(targetType, targetID, userID)
	if err != nil {
		response.Fail(c, 500, "获取记账失败")
		return
	}
	response.Success(c, accounts)
}

// AddAccount 添加一条记账条目（可绑定行程/攻略/搭子）
// @Summary 添加记账
// @Security BearerAuth
// @Tags 小程序-记账
// @Param body body object{targetType=string,targetId=string,category=string,amount=number,note=string,consumedAt=string} true "记账信息"
// @Success 200 {object} response.Response
// @Router /api/v1/account [post]
func AddAccount(c *gin.Context) {
	var req struct {
		TargetType string  `json:"targetType"`
		TargetID   string  `json:"targetId"`
		Category   string  `json:"category"`
		Amount     float64 `json:"amount"`
		Note       string  `json:"note"`
		ConsumedAt string  `json:"consumedAt"` // 消费时间 ISO8601，可选
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 400, "参数错误")
		return
	}
	if !validTargetTypes[req.TargetType] || req.TargetID == "" || req.Amount <= 0 {
		response.Fail(c, 400, "参数错误")
		return
	}
	if req.Category == "" {
		req.Category = "其他"
	}
	acc := model.Accounting{
		TargetType: req.TargetType,
		TargetID:   req.TargetID,
		UserID:     c.MustGet("userID").(string),
		Category:   req.Category,
		Amount:     req.Amount,
		Note:       req.Note,
		ConsumedAt: time.Now(),
	}
	if req.ConsumedAt != "" {
		if t, err := time.Parse(time.RFC3339, req.ConsumedAt); err == nil {
			acc.ConsumedAt = t
		}
	}
	if err := repository.CreateAccount(&acc); err != nil {
		response.Fail(c, 500, "添加失败")
		return
	}
	response.Success(c, acc)
}

// DeleteAccount 删除一条记账记录
// @Summary 删除记账
// @Security BearerAuth
// @Tags 小程序-记账
// @Param id path string true "记账ID"
// @Success 200 {object} response.Response
// @Router /api/v1/account/{id} [delete]
func DeleteAccount(c *gin.Context) {
	userID := c.MustGet("userID").(string)
	if err := repository.DeleteAccount(c.Param("id"), userID); err != nil {
		response.Fail(c, 404, "记账条目不存在")
		return
	}
	response.Success(c, nil)
}

// GetAccountSummary 获取指定目标的账本汇总
// @Summary 账本汇总
// @Security BearerAuth
// @Tags 小程序-记账
// @Param targetType query string true "绑定类型：trip行程 guide攻略 partner搭子"
// @Param targetId query string true "绑定目标ID"
// @Success 200 {object} response.Response{data=repository.AccountSummary}
// @Router /api/v1/account/summary [get]
func GetAccountSummary(c *gin.Context) {
	targetType := c.Query("targetType")
	targetID := c.Query("targetId")
	if !validTargetTypes[targetType] || targetID == "" {
		response.Fail(c, 400, "参数错误")
		return
	}
	userID := c.MustGet("userID").(string)
	summary, err := repository.GetAccountSummary(targetType, targetID, userID)
	if err != nil {
		response.Fail(c, 500, "获取汇总失败")
		return
	}
	response.Success(c, summary)
}

// GetAccountOverview 获取我的账本总览（按目标聚合）
// @Summary 我的账本总览
// @Security BearerAuth
// @Tags 小程序-记账
// @Success 200 {object} response.Response{data=[]repository.AccountOverviewItem}
// @Router /api/v1/account/overview [get]
func GetAccountOverview(c *gin.Context) {
	userID := c.MustGet("userID").(string)
	items, err := repository.GetAccountOverview(userID)
	if err != nil {
		response.Fail(c, 500, "获取账本失败")
		return
	}
	if items == nil {
		items = []repository.AccountOverviewItem{}
	}
	response.Success(c, items)
}

// ImportWechatPay 批量导入微信支付账单
// @Summary 导入微信支付账单
// @Security BearerAuth
// @Tags 小程序-记账
// @Param body body object{target_type=string,target_id=string,transactions=[]string} true "账单数据"
// @Success 200 {object} response.Response
// @Router /api/v1/account/import [post]
func ImportWechatPay(c *gin.Context) {
	var req struct {
		TargetType   string   `json:"targetType"`
		TargetID     string   `json:"targetId"`
		Transactions []string `json:"transactions"` // 微信支付单号（实际应传入金额）
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 400, "参数错误")
		return
	}
	uid := c.MustGet("userID").(string)
	for _, t := range req.Transactions {
		acc := model.Accounting{
			TargetType:    req.TargetType,
			TargetID:      req.TargetID,
			UserID:        uid,
			TransactionID: t,
			Amount:        0, // 实际金额应由前端传入
		}
		repository.CreateAccount(&acc)
	}
	response.Success(c, nil)
}
