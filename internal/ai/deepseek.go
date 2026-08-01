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
返回JSON需包含以下字段：
- title: 行程标题（如"东京五日深度游"）
- countries: 国家列表数组（境外如["日本"]，国内为[]）
- provinces: 省份列表数组（国内如["四川"]，境外为[]）
- cities: 城市列表数组（如["东京"]）
- isOverseas: 是否境外（1境外 0国内）
- totalBudget: 预估总预算（元，数字）
- summary: 行程概述（一句话简介）
- days: 每天行程数组
每天需包含上午、下午、晚上安排，每个行程项需包含以下字段：
- time: 开始时间（如09:00）
- name: 景点/餐厅/活动名称
- type: 类型（attraction景点/food美食/hotel住宿/transport交通/shopping购物/tips避坑）
- duration: 游玩时长（如2h）
- address: 地址
- description: 简短描述
只返回严格的JSON格式，不要解释：{"title":"...","countries":[],"provinces":[],"cities":[],"isOverseas":0,"totalBudget":0,"summary":"...","days":[{"day":1,"items":[{"time":"09:00","name":"...","type":"attraction","duration":"2h","address":"...","description":"..."}]}]}`

// PartnerPrompt 搭子招募提示词模板
const PartnerPrompt = `你是一个旅行搭子招募文案专家。请为目的地%s、行程%d天的旅行搭子招募生成一份文案。
返回JSON需包含以下字段：
- title: 招募标题（吸引人，如"成都5天4夜组队！美食徒步走起"）
- category: 活动分类（旅游/美食/运动/学习/探店/看展/桌游）
- destination: 目的地
- days: 出行天数
- desc: 行程简述（一两句话概括行程亮点）
- requirement: 人员要求（如"不矫情、能早起、AA制"）
- maxMembers: 招募人数上限（数字）
- genderLimit: 0不限 1仅男生 2仅女生
- feeMode: 费用模式（0免费 1AA 2组织者全包 3人均固定预算）
- budgetPerPerson: 人均预算（元，数字）
- tags: 标签JSON数组（如["徒步","拍照","美食"]，最多5个）
只返回严格的JSON格式，不要解释：{"title":"...","category":"旅游","destination":"...","days":5,"desc":"...","requirement":"...","maxMembers":4,"genderLimit":0,"feeMode":1,"budgetPerPerson":500,"tags":["徒步","拍照"]}`

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
