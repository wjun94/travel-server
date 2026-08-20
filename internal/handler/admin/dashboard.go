package admin

import (
	"time"

	"travel-server/internal/model"
	"travel-server/pkg/database"
	"travel-server/pkg/response"

	"github.com/gin-gonic/gin"
)

// countInRange 统计指定模型记录数，since 为空表示全部
func countInRange(m interface{}, since *time.Time) int64 {
	var count int64
	q := database.DB.Model(m)
	if since != nil {
		q = q.Where("created_at >= ?", since)
	}
	q.Count(&count)
	return count
}

// Dashboard 管理端首页统计（总数 + 今日/本周/本月新增）
// @Summary 仪表盘统计
// @Description 返回用户/攻略/搭子/行程/评论/收藏/搭子申请/投诉/AI生成 9 个维度的总数与今日、本周（周一起）、本月新增数据
// @Security BearerAuth
// @Tags 后台-仪表盘
// @Success 200 {object} response.Response{data=object{user=object{total=int,today=int,week=int,month=int},guide=object{total=int,today=int,week=int,month=int},partner=object{total=int,today=int,week=int,month=int},trip=object{total=int,today=int,week=int,month=int},comment=object{total=int,today=int,week=int,month=int},favorite=object{total=int,today=int,week=int,month=int},application=object{total=int,today=int,week=int,month=int},complaint=object{total=int,today=int,week=int,month=int},aiGenerate=object{total=int,today=int,week=int,month=int}}}
// @Router /api/v1/admin/dashboard [get]
func Dashboard(c *gin.Context) {
	// 时间起点（Asia/Shanghai）：今日 0 点 / 本周一 0 点 / 本月 1 号 0 点
	loc, _ := time.LoadLocation("Asia/Shanghai")
	if loc == nil {
		loc = time.Local
	}
	now := time.Now().In(loc)
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
	weekdayOffset := (int(now.Weekday()) + 6) % 7 // 周一为每周第一天（周一=0）
	weekStart := todayStart.AddDate(0, 0, -weekdayOffset)
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, loc)

	// 统计指定维度
	stat := func(m interface{}) gin.H {
		return gin.H{
			"total": countInRange(m, nil),
			"today": countInRange(m, &todayStart),
			"week":  countInRange(m, &weekStart),
			"month": countInRange(m, &monthStart),
		}
	}

	// 收藏维度：仅统计收藏记录（action=2），点赞记录不纳入
	favStat := func(m interface{}) gin.H {
		count := func(since *time.Time) int64 {
			var c int64
			q := database.DB.Model(m).Where("action = 2")
			if since != nil {
				q = q.Where("created_at >= ?", since)
			}
			q.Count(&c)
			return c
		}
		return gin.H{
			"total": count(nil),
			"today": count(&todayStart),
			"week":  count(&weekStart),
			"month": count(&monthStart),
		}
	}

	response.Success(c, gin.H{
		"user":        stat(&model.User{}),
		"guide":       stat(&model.Guide{}),
		"partner":     stat(&model.Partner{}),
		"trip":        stat(&model.Trip{}),
		"comment":     stat(&model.Comment{}),
		"favorite":    favStat(&model.Favorite{}),
		"application": stat(&model.PartnerApplication{}),
		"complaint":   stat(&model.Complaint{}),
		"aiGenerate":  stat(&model.AiGenerateLog{}),
	})
}
