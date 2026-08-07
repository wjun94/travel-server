package miniapp

import (
	"log"

	"travel-server/internal/repository"
)

// BackfillGuideIsOverseas 存量攻略国内外标记回填（幂等，服务启动时调用）
// 判定规则：目的地命中境外国家列表 → 境外；命中省市区树 → 国内；均未命中默认国内
func BackfillGuideIsOverseas() {
	guides, err := repository.ListAllGuides()
	if err != nil {
		log.Printf("回填攻略国内外标记失败: %v", err)
		return
	}
	count := 0
	for i := range guides {
		overseas := 0
		if !IsDomesticDestination(guides[i].Destination) {
			overseas = 1
		}
		if guides[i].IsOverseas != overseas {
			if err := repository.UpdateGuide(guides[i].ID, map[string]interface{}{"is_overseas": overseas}); err != nil {
				log.Printf("回填攻略 %s 失败: %v", guides[i].ID, err)
				continue
			}
			count++
		}
	}
	if count > 0 {
		log.Printf("✅ 存量攻略国内外标记回填完成，共更新 %d 条", count)
	}
}
