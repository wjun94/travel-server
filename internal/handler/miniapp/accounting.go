// Package miniapp 提供小程序端记账相关接口
package miniapp

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"travel-server/internal/model"
	"travel-server/internal/repository"
	"travel-server/pkg/response"
)

// GetAccounts 获取指定行程的记账记录
// @Summary 获取账本
// @Security BearerAuth
// @Tags 小程序-记账
// @Param tripId path int true "行程ID"
// @Success 200 {object} response.Response{data=[]model.Accounting}
// @Router /api/v1/account/{tripId} [get]
func GetAccounts(c *gin.Context) {
	tripID, err := strconv.Atoi(c.Param("tripId"))
	if err != nil {
		response.Fail(c, 400, "参数错误")
		return
	}
	userID := c.GetUint("userID")
	accounts, err := repository.GetAccountsByTrip(uint(tripID), userID)
	if err != nil {
		response.Fail(c, 500, "获取记账失败")
		return
	}
	response.Success(c, accounts)
}

// AddAccount 添加一条记账条目
// @Summary 添加记账
// @Security BearerAuth
// @Tags 小程序-记账
// @Param body body model.Accounting true "记账信息"
// @Success 200 {object} response.Response
// @Router /api/v1/account [post]
func AddAccount(c *gin.Context) {
	var acc model.Accounting
	if err := c.ShouldBindJSON(&acc); err != nil {
		response.Fail(c, 400, "参数错误")
		return
	}
	acc.UserID = c.GetUint("userID")
	if err := repository.CreateAccount(&acc); err != nil {
		response.Fail(c, 500, "添加失败")
		return
	}
	response.Success(c, acc)
}

// ImportWechatPay 批量导入微信支付账单
// @Summary 导入微信支付账单
// @Security BearerAuth
// @Tags 小程序-记账
// @Param body body object{trip_id=int,transactions=[]string} true "账单数据"
// @Success 200 {object} response.Response
// @Router /api/v1/account/import [post]
func ImportWechatPay(c *gin.Context) {
	var req struct {
		TripID       uint     `json:"tripId"`
		Transactions []string `json:"transactions"` // 微信支付单号（实际应传入金额）
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 400, "参数错误")
		return
	}
	uid := c.GetUint("userID")
	for _, t := range req.Transactions {
		acc := model.Accounting{
			TripID:        req.TripID,
			UserID:        uid,
			TransactionID: t,
			Amount:        0, // 实际金额应由前端传入
		}
		repository.CreateAccount(&acc)
	}
	response.Success(c, nil)
}
