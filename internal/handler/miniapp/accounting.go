// Package miniapp 提供小程序端记账相关接口
package miniapp

import (
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"travel-server/internal/model"
	"travel-server/internal/repository"
	"travel-server/pkg/response"
	"travel-server/pkg/snowflake"
)

// validTargetTypes 记账可绑定的目标类型（custom 为自主账本）
var validTargetTypes = map[string]bool{"trip": true, "guide": true, "partner": true, "custom": true}

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

// AddAccount 添加一条记账条目（可绑定行程/攻略/搭子，或记入自主账本）
// @Summary 添加记账
// @Security BearerAuth
// @Tags 小程序-记账
// @Param body body object{targetType=string,targetId=string,targetName=string,category=string,amount=number,note=string,consumedAt=string} true "记账信息"
// @Success 200 {object} response.Response
// @Router /api/v1/account [post]
func AddAccount(c *gin.Context) {
	var req struct {
		TargetType string  `json:"targetType"`
		TargetID   string  `json:"targetId"`
		TargetName string  `json:"targetName"` // 自主账本名（custom 时必填）
		Category   string  `json:"category"`
		Amount     float64 `json:"amount"`
		Note       string  `json:"note"`
		ConsumedAt string  `json:"consumedAt"` // 消费时间 ISO8601，可选
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 400, "参数错误")
		return
	}
	if !validTargetTypes[req.TargetType] || req.Amount <= 0 {
		response.Fail(c, 400, "参数错误")
		return
	}
	// 绑定目标必须指定 targetId；自主账本自动生成账本ID并校验账本名
	if req.TargetType == "custom" {
		if req.TargetName == "" {
			req.TargetName = "我的账本"
		}
		if req.TargetID == "" {
			req.TargetID = snowflake.GenerateID()
		}
	} else if req.TargetID == "" {
		response.Fail(c, 400, "参数错误")
		return
	}
	if req.Category == "" {
		req.Category = "其他"
	}
	acc := model.Accounting{
		TargetType: req.TargetType,
		TargetID:   req.TargetID,
		TargetName: req.TargetName,
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

// CreateAccountBook 创建自主账本（返回账本ID与名称，记第一笔后出现在账本列表）
// @Summary 创建自主账本
// @Security BearerAuth
// @Tags 小程序-记账
// @Param body body object{name=string} true "账本名称"
// @Success 200 {object} response.Response{data=object{targetId=string,targetName=string}}
// @Router /api/v1/account/book [post]
func CreateAccountBook(c *gin.Context) {
	var req struct {
		Name string `json:"name"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 400, "参数错误")
		return
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		name = "我的账本"
	}
	if len([]rune(name)) > 30 {
		response.Fail(c, 400, "账本名称过长")
		return
	}
	response.Success(c, gin.H{
		"targetId":   snowflake.GenerateID(),
		"targetName": name,
	})
}

// DeleteAccountBook 删除整本账本（该账本下的所有记账条目）
// @Summary 删除账本
// @Security BearerAuth
// @Tags 小程序-记账
// @Param targetType query string true "绑定类型：trip行程 guide攻略 partner搭子 custom自主账本"
// @Param targetId query string true "目标/账本ID"
// @Success 200 {object} response.Response
// @Router /api/v1/account/book [delete]
func DeleteAccountBook(c *gin.Context) {
	targetType := c.Query("targetType")
	targetID := c.Query("targetId")
	if !validTargetTypes[targetType] || targetID == "" {
		response.Fail(c, 400, "参数错误")
		return
	}
	userID := c.MustGet("userID").(string)
	if err := repository.DeleteAccountBook(targetType, targetID, userID); err != nil {
		response.Fail(c, 404, "账本不存在")
		return
	}
	response.Success(c, nil)
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

// UpdateAccount 编辑一条记账记录（分类/金额/备注/消费时间，仅本人）
// @Summary 编辑记账
// @Security BearerAuth
// @Tags 小程序-记账
// @Param id path string true "记账ID"
// @Param body body object{category=string,amount=number,note=string,consumedAt=string} true "记账信息"
// @Success 200 {object} response.Response
// @Router /api/v1/account/{id} [put]
func UpdateAccount(c *gin.Context) {
	var req struct {
		Category   string  `json:"category"`
		Amount     float64 `json:"amount"`
		Note       string  `json:"note"`
		ConsumedAt string  `json:"consumedAt"` // 消费时间 ISO8601，可选
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 400, "参数错误")
		return
	}
	if req.Amount <= 0 {
		response.Fail(c, 400, "金额必须大于0")
		return
	}
	if req.Category == "" {
		req.Category = "其他"
	}
	updates := map[string]interface{}{
		"category": req.Category,
		"amount":   req.Amount,
		"note":     req.Note,
	}
	if req.ConsumedAt != "" {
		if t, err := time.Parse(time.RFC3339, req.ConsumedAt); err == nil {
			updates["consumed_at"] = t
		}
	}
	userID := c.MustGet("userID").(string)
	if err := repository.UpdateAccount(c.Param("id"), userID, updates); err != nil {
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
