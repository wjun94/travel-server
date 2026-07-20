// Package ai 集成 DeepSeek Chat API 实现智能行程生成
package ai

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"

	"travel-server/pkg/config"
)

// TripPrompt 行程生成提示词模板
const TripPrompt = `你是一个专业的旅行规划师。请为%s的%d天旅行规划一个详细行程。
每天需包含上午、下午、晚上安排，每个行程项需包含以下字段：
- time: 开始时间（如09:00）
- name: 景点/餐厅/活动名称
- type: 类型（attraction景点/food美食/hotel住宿 /transport交通/shopping购物）
- duration: 游玩时长（如2h）
- address: 地址
- description: 简短描述
只返回严格的JSON格式，不要解释：{"days":[{"day":1,"items":[{"time":"09:00","name":"...","type":"attraction","duration":"2h","address":"...","description":"..."}]}]}`

// Chat 发送对话请求到 DeepSeek，返回模型回复内容
func Chat(prompt string) (string, error) {
	reqBody := map[string]interface{}{
		"model": "deepseek-chat",
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
	}
	jsonBody, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "https://api.deepseek.com/v1/chat/completions", bytes.NewReader(jsonBody))
	req.Header.Set("Authorization", "Bearer "+config.AppConfig.DeepSeekApiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	data, _ := io.ReadAll(resp.Body)
	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	json.Unmarshal(data, &result)

	if len(result.Choices) > 0 {
		return result.Choices[0].Message.Content, nil
	}
	return "", nil
}
