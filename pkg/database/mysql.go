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

	// 迁移旧索引：删除旧的 uk_user_target（不含 user_id），AutoMigrate 会创建新索引
	if err := DB.Migrator().DropIndex(&model.Favorite{}, "uk_user_target"); err != nil {
		log.Printf("uk_user_target 索引无需迁移或已删除: %v", err)
	}

	// 自动创建/更新表结构
	err = DB.AutoMigrate(
		&model.User{},
		&model.Guide{},
		&model.GuideSection{},
		&model.GuideDayItem{},
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
		&model.ChecklistCategory{},
		&model.ChecklistCategoryItem{},
		&model.Footprint{},
		&model.Recommendation{},
		&model.Favorite{},
		&model.Comment{},
		&model.BrowseHistory{},
		&model.AdminUser{},
		&model.Role{},
		&model.Follow{},
		&model.Notification{},
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
	// 初始化系统预置的备忘清单分类
	initChecklistCategories()
}

// initChecklistCategories 初始化系统预置备忘清单分类及条目
func initChecklistCategories() {
	var count int64
	DB.Model(&model.ChecklistCategory{}).Count(&count)
	if count > 0 {
		return
	}
	categories := []struct {
		name  string
		typ   int
		order int
		items []string
	}{
		// 系统内置基础分类
		{"证件&手续类", 0, 1, []string{"身份证", "护照", "签证", "驾驶证", "行驶证", "银行卡", "现金", "机票/车票/酒店订单", "保险单", "学生证", "军官证", "港澳通行证", "入台证", "接种凭证"}},
		{"电子数码类", 0, 2, []string{"手机", "充电器", "充电宝", "数据线", "耳机", "相机", "存储卡", "三脚架", "转换插头（出国）", "U盘", "手表", "剃须刀", "自拍杆", "排插"}},
		{"衣物穿搭类", 0, 3, []string{"内衣内裤", "袜子", "外套", "短袖", "长裤", "裙子", "睡衣", "帽子", "墨镜", "围巾", "拖鞋", "运动鞋", "凉鞋", "泳衣", "腰带"}},
		{"洗护化妆类", 0, 4, []string{"洗面奶", "牙膏牙刷", "毛巾", "洗发水沐浴露", "护肤品", "防晒霜", "面膜", "卸妆用品", "化妆品", "梳子", "发圈", "护手霜", "剃须刀套装"}},
		{"药品应急类", 0, 5, []string{"感冒药", "肠胃药", "创可贴", "碘伏", "晕车药", "驱蚊液", "过敏药", "退烧药", "止痛药", "棉签", "体温计", "止泻药", "纱布"}},
		{"随身便携小件", 0, 6, []string{"雨伞/雨衣", "纸巾湿巾", "垃圾袋", "口罩", "便携水杯", "钥匙", "小锁（行李箱）", "便携购物袋", "眼罩耳塞", "小剪刀"}},
		{"行李收纳类", 0, 7, []string{"行李箱", "收纳袋", "鞋袋", "压缩袋", "行李牌", "绑箱带", "防水袋"}},
		{"住宿居家用品", 0, 8, []string{"一次性床单被罩", "枕套", "马桶垫", "睡衣", "便携烧水壶", "一次性餐具"}},
		{"美食采购伴手礼", 0, 9, []string{"当地特产", "零食", "酒水", "伴手礼", "伴手礼分装袋"}},
		{"出行车辆自驾专属", 0, 10, []string{"备胎工具", "拖车绳", "玻璃水", "防滑链", "行车记录仪", "车载充电器", "停车牌"}},
		// 细分场景专用分类
		{"亲子出游专属", 1, 11, []string{"儿童换洗衣物", "奶粉奶瓶", "辅食", "纸尿裤", "湿巾", "儿童退烧药", "玩具", "推车", "防走失绳"}},
		{"海边沙滩专属", 1, 12, []string{"泳衣", "沙滩鞋", "浮潜装备", "防晒帽", "沙滩巾", "防水手机袋", "挖沙工具"}},
		{"登山徒步户外", 1, 13, []string{"登山杖", "护膝", "速干衣", "头灯", "保温水壶", "急救包", "防风外套", "防滑鞋"}},
		{"出国跨境旅行", 1, 14, []string{"外币", "电话卡/WiFi蛋", "翻译设备", "电源转化头", "入境申报单"}},
		{"商务出差", 1, 15, []string{"笔记本电脑", "文件合同", "名片", "正装", "便携打印机", "公文包"}},
	}
	for _, c := range categories {
		cat := model.ChecklistCategory{
			Name:      c.name,
			Type:      c.typ,
			SortOrder: c.order,
		}
		if err := DB.Create(&cat).Error; err != nil {
			log.Printf("创建备忘分类[%s]失败: %v", c.name, err)
			continue
		}
		for _, item := range c.items {
			DB.Create(&model.ChecklistCategoryItem{
				CategoryID: cat.ID,
				Text:       item,
			})
		}
	}
	log.Printf("初始化备忘清单分类完成，共 %d 个分类", len(categories))
}
