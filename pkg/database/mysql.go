// Package database 负责数据库连接初始化
package database

import (
	"fmt"
	"log"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"travel-server/internal/model"
	"travel-server/pkg/config"
)

var DB *gorm.DB

// InitMySQL 初始化 MySQL 连接并自动迁移表结构
func InitMySQL() {
	cfg := config.AppConfig
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.DBUser, cfg.DBPassword, cfg.DBHost, cfg.DBPort, cfg.DBName)

	var err error
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("数据库连接失败: %v", err)
	}

	// 自动创建/更新表结构
	err = DB.AutoMigrate(
		&model.User{},
		&model.Post{},
		&model.Trip{},
		&model.TripCollaborator{},
		&model.Partner{},
		&model.PartnerApplication{},
		&model.Message{},
		&model.Accounting{},
		&model.Checklist{},
		&model.ChecklistItem{},
		&model.Footprint{},
		&model.Recommendation{},
		&model.AdminUser{},
		&model.Role{},
	)
	if err != nil {
		log.Fatalf("数据库迁移失败: %v", err)
	}
	// 初始化默认角色
	var roleCount int64
	DB.Model(&model.Role{}).Count(&roleCount)
	if roleCount == 0 {
		adminRole := model.Role{
			Name:        "超级管理员",
			Description: "拥有所有权限",
			Permissions: `["*"]`,
		}
		editorRole := model.Role{
			Name:        "内容编辑",
			Description: "可管理攻略和搭子",
			Permissions: `["dashboard","posts_manage","partners_manage"]`,
		}
		DB.Create(&adminRole)
		DB.Create(&editorRole)

		// 创建默认管理员账号 (密码: admin123)
		hash, _ := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
		DB.Create(&model.AdminUser{
			Username:     "admin",
			PasswordHash: string(hash),
			RoleID:       adminRole.ID,
			Status:       1,
		})
	}
}
