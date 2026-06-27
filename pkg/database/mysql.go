// Package database 负责数据库连接初始化
package database

import (
	"fmt"
	"log"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"travel-server/internal/model"
	"travel-server/pkg/config"
	"travel-server/pkg/snowflake"
)

var DB *gorm.DB

// InitMySQL 初始化 MySQL 连接并自动迁移表结构
func InitMySQL() {
	// 初始化雪花算法 ID 生成器
	if err := snowflake.Init(); err != nil {
		log.Fatalf("雪花算法初始化失败: %v", err)
	}

	cfg := config.AppConfig
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.DBUser, cfg.DBPassword, cfg.DBHost, cfg.DBPort, cfg.DBName)

	var err error
	// 重试连接，最多尝试 30 次，每次间隔 2 秒
	for i := 0; i < 30; i++ {
		DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
		if err == nil {
			break
		}
		log.Printf("数据库连接失败(第%d次重试): %v", i+1, err)
		time.Sleep(2 * time.Second)
	}
	if err != nil {
		log.Fatalf("数据库连接失败，已超出重试次数: %v", err)
	}

	// 自动创建/更新表结构
	err = DB.AutoMigrate(
		&model.User{},
		&model.Guide{},
		&model.GuideSection{},
		&model.Trip{},
		&model.TripDay{},
		&model.TripItem{},
		&model.TripMember{},
		&model.Partner{},
		&model.PartnerApplication{},
		&model.Message{},
		&model.Accounting{},
		&model.Checklist{},
		&model.ChecklistItem{},
		&model.Footprint{},
		&model.Recommendation{},
		&model.Favorite{},
		&model.Comment{},
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
			Permissions: `["dashboard","guides_manage","partners_manage"]`,
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
