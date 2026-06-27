package model

import (
	"fmt"
	"strings"
)

// CreateGuideReq 创建攻略请求 — 包含攻略基本信息 + 板块列表
type CreateGuideReq struct {
	Title           string             `json:"title" binding:"required"`
	CoverImage      string             `json:"coverImage" binding:"required"`
	Destination     string             `json:"destination" binding:"required"`
	Summary         string             `json:"summary" binding:"required"`
	BudgetMin       *float64           `json:"budgetMin"`
	BudgetMax       *float64           `json:"budgetMax"`
	BestSeason      string             `json:"bestSeason"`
	RecommendedDays *int               `json:"recommendedDays"`
	Tags            string             `json:"tags"`
	Difficulty      string             `json:"difficulty"`
	CrowdType       string             `json:"crowdType"`
	VideoURL        string             `json:"videoUrl"`
	Images          string             `json:"images"`
	IsOriginal      int                `json:"isOriginal"`
	Status          int                `json:"status"` // 0草稿 / 1已发布
	Sections        []CreateSectionReq `json:"sections" binding:"required,min=1"`
}

// CreateSectionReq 创建板块请求
type CreateSectionReq struct {
	SectionType string `json:"sectionType" binding:"required"`
	Title       string `json:"title" binding:"required"`
	Content     string `json:"content" binding:"required"`
}

// UpdateGuideReq 更新攻略请求（所有字段可选）
type UpdateGuideReq struct {
	Title           *string  `json:"title"`
	CoverImage      *string  `json:"coverImage"`
	Destination     *string  `json:"destination"`
	Summary         *string  `json:"summary"`
	BudgetMin       *float64 `json:"budgetMin"`
	BudgetMax       *float64 `json:"budgetMax"`
	BestSeason      *string  `json:"bestSeason"`
	RecommendedDays *int     `json:"recommendedDays"`
	Tags            *string  `json:"tags"`
	Difficulty      *string  `json:"difficulty"`
	CrowdType       *string  `json:"crowdType"`
	VideoURL        *string  `json:"videoUrl"`
	Images          *string  `json:"images"`
	IsOriginal      *int     `json:"isOriginal"`
	Status          *int     `json:"status"`
}

// ==================== 校验逻辑 ====================

// ValidateCreateGuideReq 校验创建攻略请求，返回 nil 表示通过
func ValidateCreateGuideReq(req *CreateGuideReq) *ValidationError {
	// title: 不能为空，长度 5~200
	if len(strings.TrimSpace(req.Title)) < 5 || len(req.Title) > 200 {
		return &ValidationError{Field: "title", Msg: "标题长度需在5~200字符之间"}
	}
	// cover_image: 不能为空
	if strings.TrimSpace(req.CoverImage) == "" {
		return &ValidationError{Field: "coverImage", Msg: "封面图不能为空"}
	}
	// destination: 不能为空
	if strings.TrimSpace(req.Destination) == "" {
		return &ValidationError{Field: "destination", Msg: "目的地不能为空"}
	}
	// summary: 不能为空，长度 10~500
	s := strings.TrimSpace(req.Summary)
	if len(s) < 10 || len(s) > 500 {
		return &ValidationError{Field: "summary", Msg: "摘要长度需在10~500字符之间"}
	}
	// budget_min / budget_max: 若填写，必须为正数，且 min ≤ max
	if req.BudgetMin != nil && req.BudgetMax != nil {
		if *req.BudgetMin <= 0 || *req.BudgetMax <= 0 {
			return &ValidationError{Field: "budgetMin", Msg: "预算必须为正数"}
		}
		if *req.BudgetMin > *req.BudgetMax {
			return &ValidationError{Field: "budgetMin", Msg: "预算下限不能大于上限"}
		}
	}
	// recommended_days: 若填写，必须 ≥ 1
	if req.RecommendedDays != nil && *req.RecommendedDays < 1 {
		return &ValidationError{Field: "recommendedDays", Msg: "建议天数至少为1"}
	}
	// status: 只能是 0 或 1
	if req.Status != 0 && req.Status != 1 {
		return &ValidationError{Field: "status", Msg: "状态值无效"}
	}
	// sections: 至少 1 个板块，且每个板块的 type/title/content 不能为空
	if len(req.Sections) == 0 {
		return &ValidationError{Field: "sections", Msg: "至少需要1个内容板块"}
	}
	for i, sec := range req.Sections {
		if !ValidSectionTypes[sec.SectionType] {
			return &ValidationError{Field: fmt.Sprintf("sections[%d].sectionType", i),
				Msg: fmt.Sprintf("无效的板块类型: %s", sec.SectionType)}
		}
		if strings.TrimSpace(sec.Title) == "" {
			return &ValidationError{Field: fmt.Sprintf("sections[%d].title", i), Msg: "板块标题不能为空"}
		}
		if strings.TrimSpace(sec.Content) == "" {
			return &ValidationError{Field: fmt.Sprintf("sections[%d].content", i), Msg: "板块内容不能为空"}
		}
	}
	return nil
}

// ==================== 类型 ====================

// ValidationError 字段级校验错误
type ValidationError struct {
	Field string `json:"field"`
	Msg   string `json:"msg"`
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Msg)
}
