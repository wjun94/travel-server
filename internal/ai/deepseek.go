// Package ai 集成 DeepSeek Chat API 实现智能行程生成
package ai

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"

	"travel-server/pkg/config"
)

// TripPrompt 行程生成提示词模板（%s目的地 %d天数 %s出发日期说明）
const TripPrompt = `你是一个专业的旅行规划师。请为%s的%d天旅行规划一个详细行程，出发日期：%s。
每天行程需标注具体日期date（格式YYYY-MM-DD，第N天 = 出发日期 + N-1天）。
返回JSON需包含以下字段：
- title: 行程标题（如"东京五日深度游"）
- countries: 国家列表数组（境外如["日本"]，国内为[]）
- provinces: 省份列表数组（国内如["四川"]，境外为[]）
- cities: 城市列表数组（如["东京"]）
- isOverseas: 是否境外（1境外 0国内）
- totalBudget: 预估总预算（元，数字）
- summary: 行程概述（一句话简介）
- days: 每天行程数组，每天包含：
  - day: 第几天（1开始）
  - date: 当天具体日期（格式YYYY-MM-DD）
  - items: 当天行程项（需包含上午、下午、晚上安排），每项包含：
    - time: 开始时间（如09:00）
    - name: 景点/餐厅/活动名称
    - type: 类型（attraction景点/food美食/hotel住宿/transport交通/shopping购物/tips避坑）
    - duration: 游玩时长（如2h）
    - address: 地址
    - description: 简短描述
只返回严格的JSON格式，不要解释：{"title":"...","countries":[],"provinces":[],"cities":[],"isOverseas":0,"totalBudget":0,"summary":"...","days":[{"day":1,"date":"2026-09-01","items":[{"time":"09:00","name":"...","type":"attraction","duration":"2h","address":"...","description":"..."}]}]}`

// PartnerPrompt 搭子招募提示词模板（%s目的地 %d天数 %d天数 %s出发日期说明）
const PartnerPrompt = `你是一个旅行搭子招募文案专家。请为目的地%s、行程%d天的旅行搭子招募生成一份完整的文案和行程安排，出发日期：%s。
返回JSON需包含以下字段：
- title: 招募标题（吸引人，如"成都5天4夜组队！美食徒步走起"）
- category: 活动分类（旅游/美食/运动/学习/探店/看展/桌游）
- destination: 目的地
- days: 出行天数（数字）
- desc: 行程简述（一两句话概括行程亮点）
- richDesc: 详细介绍（3-5句话，包含路线安排、行程亮点、特色体验）
- requirement: 人员要求（如"不矫情、能早起、AA制"）
- address: 集合地点（如"成都东站集合"）
- startDate: 出发日期（格式YYYY-MM-DD，必须等于用户指定出发日期，若未指定则从今天开始）
- endDate: 结束日期（格式YYYY-MM-DD，等于startDate+days-1天）
- maxMembers: 招募人数上限（数字，4-10）
- minMembers: 最小成团人数（数字，2-4，不超过maxMembers）
- genderLimit: 0不限 1仅男生 2仅女生
- maleCount: 男生需求数（genderLimit为1时>0，否则0）
- femaleCount: 女生需求数（genderLimit为2时>0，否则0）
- minAge: 年龄下限（数字）
- maxAge: 年龄上限（数字）
- feeMode: 费用模式（0免费 1AA 2组织者全包 3人均固定预算）
- budgetPerPerson: 人均预算（元，数字）
- feeInclude: 费用包含（如"住宿、门票、交通"）
- feeExclude: 费用不含（如"餐饮、个人消费"）
- estTotal: 预估总花费（元，数字）
- tags: 标签JSON数组（如["徒步","拍照","美食"]，最多5个）
- schedule: 每日行程安排数组（共%d天），每天包含：
  - dayNumber: 第几天（1开始）
  - date: 当天具体日期（格式YYYY-MM-DD，第N天 = 出发日期 + N-1天）
  - title: 当天行程主题（如"抵达成都，火锅初体验"）
  - items: 当天行程项数组（每天3-6项，涵盖景点、美食、住宿、交通等），每项包含：
    - sectionType: 类型（attraction景点/food美食/hotel住宿/transport交通/shopping购物/tips避坑）
    - title: 行程项名称（如"宽窄巷子"）
    - description: 简短描述
    - startTime: 开始时间（如09:00）
    - endTime: 结束时间（如11:00）
    - address: 地点地址
    - startPoint: 起点名称（仅transport类型填写，其他为空）
    - endPoint: 终点名称（仅transport类型填写，其他为空）
    - transportMode: 交通方式（train/bus/metro/taxi/walk/plane，仅transport类型填写，其他为空）
只返回严格的JSON格式，不要解释：{"title":"...","category":"旅游","destination":"...","days":5,"desc":"...","richDesc":"...","requirement":"...","address":"...","startDate":"2026-08-08","endDate":"2026-08-12","maxMembers":6,"minMembers":2,"genderLimit":0,"maleCount":0,"femaleCount":0,"minAge":18,"maxAge":40,"feeMode":1,"budgetPerPerson":500,"feeInclude":"住宿、门票","feeExclude":"餐饮、个人消费","estTotal":2000,"tags":["徒步","拍照"],"schedule":[{"dayNumber":1,"date":"2026-08-08","title":"抵达成都，火锅初体验","items":[{"sectionType":"transport","title":"前往成都","description":"高铁/飞机抵达","startTime":"09:00","endTime":"12:00","address":"成都东站","startPoint":"出发城市","endPoint":"成都东站","transportMode":"train"},{"sectionType":"food","title":"火锅晚餐","description":"体验地道成都火锅","startTime":"18:00","endTime":"20:00","address":"春熙路","startPoint":"","endPoint":"","transportMode":""}]}]}`

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
