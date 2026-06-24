package model

import (
	"database/sql/driver"
	"errors"
	"time"
)

// Trip 行程表
type Trip struct {
	ID            uint               `gorm:"primaryKey" json:"id"`
	UserID        uint               `json:"userId"`                       // 创建者 ID
	Title         string             `gorm:"size:200" json:"title"`         // 行程标题
	Destination   string             `gorm:"size:200" json:"destination"`   // 目的地
	Days          int                `json:"days"`                          // 天数
	StartDate     *time.Time         `json:"startDate"`                    // 开始日期
	DailyPlans    JSONString         `gorm:"type:text" json:"dailyPlans"`  // 每日计划 JSON
	WeatherData   JSONString         `gorm:"type:text" json:"weatherData"` // 天气数据 JSON
	Status        int                `gorm:"default:0" json:"status"`       // 0草稿 1已发布 2协同中
	Version       int                `gorm:"default:0" json:"version"`      // 乐观锁版本号
	CreatedAt     time.Time          `json:"createdAt"`
	Collaborators []TripCollaborator `gorm:"foreignKey:TripID" json:"collaborators"` // 协作者列表
}

// TripCollaborator 行程协作者
type TripCollaborator struct {
	TripID     uint `gorm:"primaryKey" json:"tripId"`
	UserID     uint `gorm:"primaryKey" json:"userId"`
	Permission int  `gorm:"default:1" json:"permission"` // 1编辑 2只读
}

// JSONString 自定义 JSON 字段类型，用于 GORM 存储
type JSONString string

func (j JSONString) Value() (driver.Value, error) {
	if j == "" {
		return nil, nil
	}
	return string(j), nil
}

func (j *JSONString) Scan(value interface{}) error {
	if value == nil {
		*j = ""
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("类型不匹配")
	}
	*j = JSONString(bytes)
	return nil
}

// DailyPlan 每天的计划
type DailyPlan struct {
	Day   int        `json:"day"`
	Items []PlanItem `json:"items"`
}

// PlanItem 单个行程点
type PlanItem struct {
	Time     string `json:"time"`
	Name     string `json:"name"`
	Type     string `json:"type"`     // attraction / restaurant / other
	Duration string `json:"duration"` // 游玩时长，如 "2h"
	Note     string `json:"note"`
}
