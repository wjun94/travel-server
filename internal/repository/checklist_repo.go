package repository

import (
	"travel-server/internal/model"
	"travel-server/pkg/database"

	"gorm.io/gorm"
)

// GetChecklistsByUser 获取用户的备忘清单（分页，填充关联名称）
func GetChecklistsByUser(userID string, page, pageSize int) ([]model.Checklist, int64, error) {
	var lists []model.Checklist
	var total int64
	offset := (page - 1) * pageSize
	database.DB.Model(&model.Checklist{}).Where("user_id = ?", userID).Count(&total)
	err := database.DB.Where("user_id = ?", userID).Order("created_at desc").Offset(offset).Limit(pageSize).Preload("Items").Find(&lists).Error
	if err == nil {
		ptrLists := make([]*model.Checklist, len(lists))
		for i := range lists {
			ptrLists[i] = &lists[i]
		}
		fillChecklistTargetNames(ptrLists)
	}
	return lists, total, err
}

// CreateChecklist 创建清单
func CreateChecklist(cl *model.Checklist) error {
	return database.DB.Create(cl).Error
}

// UpdateChecklistItem 更新清单条目勾选状态
func UpdateChecklistItem(id string, checked int) error {
	return database.DB.Model(&model.ChecklistItem{}).Where("id = ?", id).Update("checked", checked).Error
}

// GetChecklistDetail 获取单个备忘清单详情（含条目与关联名称）
func GetChecklistDetail(id, userID string) (*model.Checklist, error) {
	var cl model.Checklist
	err := database.DB.Where("id = ? AND user_id = ?", id, userID).Preload("Items").First(&cl).Error
	if err == nil {
		fillChecklistTargetNames([]*model.Checklist{&cl})
	}
	return &cl, err
}

// UpdateChecklist 更新备忘清单（名称 + 关联 + 替换条目）
// hasTarget=true 表示明确设置关联（targetType/targetID 可为空串，即取消关联）
func UpdateChecklist(id, userID, name, targetType, targetID string, hasTarget bool, items []model.ChecklistItem) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		// 校验归属
		var cl model.Checklist
		if err := tx.Where("id = ? AND user_id = ?", id, userID).First(&cl).Error; err != nil {
			return err
		}
		// 更新名称与关联（trip 类型同步写兼容字段 trip_id，其他类型清空）
		updates := map[string]interface{}{}
		if name != "" {
			updates["name"] = name
		}
		if hasTarget {
			updates["target_type"] = targetType
			updates["target_id"] = targetID
			if targetType == "trip" {
				updates["trip_id"] = targetID
			} else {
				updates["trip_id"] = ""
			}
		}
		if len(updates) > 0 {
			if err := tx.Model(&cl).Updates(updates).Error; err != nil {
				return err
			}
		}
		// 删除旧条目
		if err := tx.Where("checklist_id = ?", id).Delete(&model.ChecklistItem{}).Error; err != nil {
			return err
		}
		// 插入新条目
		for i := range items {
			items[i].ID = ""
			items[i].ChecklistID = id
		}
		if len(items) > 0 {
			if err := tx.Create(&items).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// fillChecklistTargetNames 批量填充清单关联名称（行程/攻略/搭子标题）
// 兼容旧数据：target_type 为空但 trip_id 有值时按行程处理
func fillChecklistTargetNames(lists []*model.Checklist) {
	tripIDs, guideIDs, partnerIDs := []string{}, []string{}, []string{}
	for _, cl := range lists {
		tt, tid := cl.TargetType, cl.TargetID
		if tt == "" && cl.TripID != "" { // 旧数据兼容
			tt, tid = "trip", cl.TripID
			cl.TargetType, cl.TargetID = tt, tid
		}
		switch tt {
		case "trip":
			tripIDs = append(tripIDs, tid)
		case "guide":
			guideIDs = append(guideIDs, tid)
		case "partner":
			partnerIDs = append(partnerIDs, tid)
		}
	}
	nameMap := make(map[string]string)
	if len(tripIDs) > 0 {
		var rows []model.Trip
		database.DB.Select("id, title").Where("id IN ?", tripIDs).Find(&rows)
		for _, r := range rows {
			nameMap["trip:"+r.ID] = r.Title
		}
	}
	if len(guideIDs) > 0 {
		var rows []model.Guide
		database.DB.Select("id, title").Where("id IN ?", guideIDs).Find(&rows)
		for _, r := range rows {
			nameMap["guide:"+r.ID] = r.Title
		}
	}
	if len(partnerIDs) > 0 {
		var rows []model.Partner
		database.DB.Select("id, title").Where("id IN ?", partnerIDs).Find(&rows)
		for _, r := range rows {
			nameMap["partner:"+r.ID] = r.Title
		}
	}
	for _, cl := range lists {
		if cl.TargetType == "" {
			continue
		}
		cl.TargetName = nameMap[cl.TargetType+":"+cl.TargetID]
	}
}

// GetChecklistCategories 获取系统预置的备忘清单分类（含条目）
func GetChecklistCategories() ([]model.ChecklistCategory, error) {
	var cats []model.ChecklistCategory
	err := database.DB.Order("sort_order asc").Preload("Items").Find(&cats).Error
	return cats, err
}

// DeleteChecklist 删除备忘清单（级联删除条目）
func DeleteChecklist(id, userID string) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("checklist_id = ?", id).Delete(&model.ChecklistItem{}).Error; err != nil {
			return err
		}
		return tx.Where("id = ? AND user_id = ?", id, userID).Delete(&model.Checklist{}).Error
	})
}
