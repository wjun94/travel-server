package cron

import (
	"log"
	"time"

	"travel-server/internal/repository"
)

// Start 启动所有定时任务
func Start() {
	go cleanupBrowseHistory()
	go checkExpiredPartners()
}

// cleanupBrowseHistory 每天凌晨3点清理30天前的浏览记录
func cleanupBrowseHistory() {
	for {
		now := time.Now()
		next := time.Date(now.Year(), now.Month(), now.Day(), 3, 0, 0, 0, now.Location())
		if next.Before(now) {
			next = next.Add(24 * time.Hour)
		}
		time.Sleep(next.Sub(now))
		repository.CleanupBrowseHistory(30)
		log.Println("浏览历史清理完成")
	}
}

// checkExpiredPartners 每分钟检查过期未满员搭子，自动关闭
func checkExpiredPartners() {
	for {
		affected := repository.AutoCloseExpiredPartners()
		if affected > 0 {
			log.Printf("自动关闭 %d 个过期未满员搭子", affected)
		}
		time.Sleep(1 * time.Minute)
	}
}
