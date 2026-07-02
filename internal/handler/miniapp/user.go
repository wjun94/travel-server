package miniapp

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"travel-server/internal/middleware"
	"travel-server/internal/model"
	"travel-server/internal/repository"
	"travel-server/internal/service"
	"travel-server/pkg/config"
	"travel-server/pkg/database"
	"travel-server/pkg/response"
)

// UserLogin 微信小程序登录
// @Summary 微信登录
// @Description 通过临时code换取openid，完成登录/注册，返回JWT Token
// @Tags 小程序-用户
// @Accept json
// @Produce json
// @Param code body string true "微信临时登录凭证"
// @Success 200 {object} response.Response{data=object{token=string,user=model.User}}
// @Router /api/v1/user/login [post]
func UserLogin(c *gin.Context) {
	var req struct {
		Code string `json:"code" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 400, "参数错误")
		return
	}

	wxResp, err := getWxSession(req.Code)
	if err != nil || wxResp.ErrCode != 0 {
		response.Fail(c, 500, "微信登录失败")
		return
	}

	user, err := service.GetOrCreateUser(wxResp.OpenID)
	if err != nil {
		response.Fail(c, 500, "服务器错误")
		return
	}

	token, _ := middleware.GenerateMiniAppToken(user.ID)
	response.Success(c, gin.H{"token": token, "user": user})
}

type WxSessionResp struct {
	OpenID     string `json:"openid"`
	SessionKey string `json:"session_key"`
	UnionID    string `json:"unionid"`
	ErrCode    int    `json:"errcode"`
	ErrMsg     string `json:"errmsg"`
}

func getWxSession(code string) (*WxSessionResp, error) {
	cfg := config.AppConfig
	url := fmt.Sprintf("https://api.weixin.qq.com/sns/jscode2session?appid=%s&secret=%s&js_code=%s&grant_type=authorization_code",
		cfg.AppId, cfg.AppSecret, code)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var wxResp WxSessionResp
	if err := json.NewDecoder(resp.Body).Decode(&wxResp); err != nil {
		return nil, err
	}
	return &wxResp, nil
}

// GetUserInfo 获取当前用户信息
// @Summary 获取个人信息
// @Security BearerAuth
// @Tags 小程序-用户
// @Success 200 {object} response.Response{data=model.User}
// @Router /api/v1/user/info [get]
func GetUserInfo(c *gin.Context) {
	uid := c.MustGet("userID").(string)
	user, err := repository.GetUserByID(uid)
	if err != nil {
		response.Fail(c, 401, "用户不存在")
		return
	}
	response.Success(c, user)
}

// UpdateProfile 更新用户头像昵称
// @Summary 更新个人资料
// @Security BearerAuth
// @Tags 小程序-用户
// @Param body body object{nickname=string,avatar_url=string} true "资料"
// @Success 200 {object} response.Response
// @Router /api/v1/user/profile [put]
func UpdateProfile(c *gin.Context) {
	var req struct {
		Nickname  string `json:"nickname"`
		AvatarURL string `json:"avatarUrl"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, 400, "参数错误")
		return
	}
	uid := c.MustGet("userID").(string)
	database.DB.Model(&model.User{}).Where("id = ?", uid).Updates(map[string]interface{}{
		"nickname":   req.Nickname,
		"avatar_url": req.AvatarURL,
	})
	response.Success(c, nil)
}
