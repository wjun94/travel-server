package miniapp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"travel-server/internal/model"
	"travel-server/pkg/database"
	"travel-server/pkg/response"
)

// 微信内容安全检测场景枚举（security.msgSecCheck）
const (
	secSceneProfile = 1 // 资料（昵称等）
	secSceneComment = 2 // 评论
	secSceneForum   = 3 // 论坛（攻略/行程/搭子/清单等）
	secSceneSocial  = 4 // 社交日志（私聊/群聊消息）
)

// checkContentSec 调用微信内容安全 API（security.msgSecCheck v2）校验用户发布文本。
// 返回 false 表示内容含违规信息；返回 err 表示检测过程失败（上层按违规拦截处理，
// 保证内容安全策略在小程序内任意发布场景生效）。
func checkContentSec(userID string, scene int, content string) (bool, error) {
	content = strings.TrimSpace(content)
	if content == "" {
		return true, nil
	}
	// 微信接口单次最多 2500 字，超长截断
	if runes := []rune(content); len(runes) > 2500 {
		content = string(runes[:2500])
	}
	// 获取 openid（用于微信侧识别违规内容来源，降低投诉/举报风险）
	var user model.User
	database.DB.Select("open_id").Where("id = ?", userID).First(&user)

	token, err := getWxAccessToken()
	if err != nil {
		return false, err
	}
	url := "https://api.weixin.qq.com/wxa/msg_sec_check?access_token=" + token
	payload, _ := json.Marshal(map[string]interface{}{
		"version": 2,
		"openid":  user.OpenID,
		"scene":   scene,
		"content": content,
	})
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Post(url, "application/json", bytes.NewReader(payload))
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	var checkResp struct {
		ErrCode int    `json:"errcode"`
		ErrMsg  string `json:"errmsg"`
		Result  struct {
			Suggest string `json:"suggest"`
		} `json:"result"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&checkResp); err != nil {
		return false, err
	}
	if checkResp.ErrCode == 87014 { // 内容含违规信息
		return false, nil
	}
	if checkResp.ErrCode != 0 {
		return false, fmt.Errorf("内容安全检测失败: %s", checkResp.ErrMsg)
	}
	// v2 suggest: pass 通过 / review 建议人工复核 / risky 违规
	return checkResp.Result.Suggest != "risky", nil
}

// secGuard 内容安全统一拦截：检测不通过时写入错误响应并返回 false。
// 违规内容仅提示"含违规信息"，检测失败提示稍后重试。
func secGuard(c *gin.Context, userID string, scene int, content string) bool {
	safe, err := checkContentSec(userID, scene, content)
	if err != nil {
		response.Fail(c, 500, "内容安全检查失败，请稍后重试")
		return false
	}
	if !safe {
		response.Fail(c, 400, "发布内容含违规信息，请修改后重试")
		return false
	}
	return true
}

// secText 拼接多个文本字段用于统一检测（自动忽略空值）
func secText(parts ...string) string {
	texts := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			texts = append(texts, p)
		}
	}
	return strings.Join(texts, " ")
}
