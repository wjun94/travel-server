package miniapp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"travel-server/internal/repository"
	"travel-server/pkg/config"
	"travel-server/pkg/response"
)

// wxAccessTokenCache access_token 内存缓存（有效期2小时，提前5分钟过期）
var (
	wxAccessTokenCache    string
	wxAccessTokenExpireAt time.Time
)

// getWxAccessToken 获取微信小程序全局 access_token（带缓存）
func getWxAccessToken() (string, error) {
	if wxAccessTokenCache != "" && time.Now().Before(wxAccessTokenExpireAt) {
		return wxAccessTokenCache, nil
	}
	cfg := config.AppConfig
	url := fmt.Sprintf("https://api.weixin.qq.com/cgi-bin/token?grant_type=client_credential&appid=%s&secret=%s",
		cfg.AppId, cfg.AppSecret)
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	var tokenResp struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
		ErrCode     int    `json:"errcode"`
		ErrMsg      string `json:"errmsg"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return "", err
	}
	if tokenResp.AccessToken == "" {
		return "", fmt.Errorf("获取access_token失败: %s", tokenResp.ErrMsg)
	}
	wxAccessTokenCache = tokenResp.AccessToken
	wxAccessTokenExpireAt = time.Now().Add(time.Duration(tokenResp.ExpiresIn-300) * time.Second)
	return tokenResp.AccessToken, nil
}

// getWxPhoneNumber 通过前端 getPhoneNumber 返回的 code 换取手机号
func getWxPhoneNumber(code string) (string, error) {
	token, err := getWxAccessToken()
	if err != nil {
		return "", err
	}
	url := fmt.Sprintf("https://api.weixin.qq.com/wxa/business/getuserphonenumber?access_token=%s", token)
	payload, _ := json.Marshal(map[string]string{"code": code})
	resp, err := http.Post(url, "application/json", bytes.NewReader(payload))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	var phoneResp struct {
		ErrCode   int    `json:"errcode"`
		ErrMsg    string `json:"errmsg"`
		PhoneInfo struct {
			PhoneNumber     string `json:"phoneNumber"`
			PurePhoneNumber string `json:"purePhoneNumber"`
		} `json:"phone_info"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&phoneResp); err != nil {
		return "", err
	}
	if phoneResp.ErrCode != 0 {
		return "", fmt.Errorf("获取手机号失败: %s", phoneResp.ErrMsg)
	}
	if phoneResp.PhoneInfo.PhoneNumber == "" {
		return "", fmt.Errorf("获取手机号失败: 手机号为空")
	}
	return phoneResp.PhoneInfo.PhoneNumber, nil
}

// BindPhone 绑定微信手机号
// @Summary 绑定微信手机号
// @Description 前端通过 button open-type=getPhoneNumber 获取 code，后端调用微信接口解码并直接更新用户手机号
// @Security BearerAuth
// @Tags 小程序-用户
// @Param body body object{code=string} true "微信手机号授权code"
// @Success 200 {object} response.Response{data=object{phone=string}}
// @Router /api/v1/bind/phone [post]
func BindPhone(c *gin.Context) {
	var req struct {
		Code string `json:"code" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 400, "参数错误")
		return
	}
	uid := c.MustGet("userID").(string)

	phone, err := getWxPhoneNumber(req.Code)
	if err != nil {
		response.Fail(c, 500, err.Error())
		return
	}
	if err := repository.UpdateUserPhone(uid, phone); err != nil {
		response.Fail(c, 500, "保存失败")
		return
	}
	response.Success(c, gin.H{"phone": phone})
}
