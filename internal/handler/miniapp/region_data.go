package miniapp

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"sync"
)

// ✅ 修正1：添加Children字段匹配原始JSON结构
type rawRegion struct {
	Code     string      `json:"code"`
	Name     string      `json:"name"`
	Children []rawRegion `json:"children,omitempty"` // 关键修复：添加嵌套结构
}

// RegionNode 树形节点 (保持输出结构不变)
type RegionNode struct {
	ID       string       `json:"id"`
	Name     string       `json:"name"`
	Children []RegionNode `json:"children,omitempty"`
}

var (
	regionTree []RegionNode
	once       sync.Once
)

// loadRegionTree 从项目根目录的 data/pca-code.json 加载
func loadRegionTree() []RegionNode {
	once.Do(func() {
		log.Println("🔍 开始加载省市区数据...")

		// 获取当前工作目录
		wd, err := os.Getwd()
		if err != nil {
			log.Fatalf("❌ 获取工作目录失败: %v", err)
		}
		log.Printf("📂 当前工作目录: %s", wd)

		// 构建文件路径
		filePath := filepath.Join(wd, "data", "pca-code.json")
		log.Printf("📄 尝试读取文件: %s", filePath)

		// 读取文件
		data, err := os.ReadFile(filePath)
		if err != nil {
			log.Fatalf("❌ 读取 pca-code.json 失败: %v", err)
		}

		// ✅ 修正2：直接解析为树形结构 (不再需要手动组装)
		var rawTree []rawRegion
		if err := json.Unmarshal(data, &rawTree); err != nil {
			log.Fatalf("❌ 解析 pca-code.json 失败: %v", err)
		}
		log.Printf("✅ 成功加载 %d 个省级节点 (含子节点)", len(rawTree))

		// ✅ 修正3：递归转换 rawRegion -> RegionNode
		regionTree = convertToRegionNode(rawTree)
		log.Printf("✅ 转换完成，共 %d 个省份", len(regionTree))
	})
	return regionTree
}

// 递归转换函数：rawRegion -> RegionNode
func convertToRegionNode(raw []rawRegion) []RegionNode {
	var result []RegionNode
	for _, r := range raw {
		// 转换当前节点
		id, _ := strconv.ParseInt(r.Code, 10, 64) // 忽略错误（已在日志中处理）

		node := RegionNode{
			ID:   strconv.FormatInt(id, 10),
			Name: r.Name,
		}

		// 递归转换子节点
		if len(r.Children) > 0 {
			node.Children = convertToRegionNode(r.Children)
		}

		result = append(result, node)
	}
	return result
}

// countryData 境外国家/地区列表
var countryData = []CountryItem{
	{Code: "JP", Name: "日本", NameEn: "Japan", Emoji: "🇯🇵", Phone: "+81"},
	{Code: "KR", Name: "韩国", NameEn: "South Korea", Emoji: "🇰🇷", Phone: "+82"},
	{Code: "TH", Name: "泰国", NameEn: "Thailand", Emoji: "🇹🇭", Phone: "+66"},
	{Code: "SG", Name: "新加坡", NameEn: "Singapore", Emoji: "🇸🇬", Phone: "+65"},
	{Code: "MY", Name: "马来西亚", NameEn: "Malaysia", Emoji: "🇲🇾", Phone: "+60"},
	{Code: "VN", Name: "越南", NameEn: "Vietnam", Emoji: "🇻🇳", Phone: "+84"},
	{Code: "PH", Name: "菲律宾", NameEn: "Philippines", Emoji: "🇵🇭", Phone: "+63"},
	{Code: "ID", Name: "印度尼西亚", NameEn: "Indonesia", Emoji: "🇮🇩", Phone: "+62"},
	{Code: "KH", Name: "柬埔寨", NameEn: "Cambodia", Emoji: "🇰🇭", Phone: "+855"},
	{Code: "LA", Name: "老挝", NameEn: "Laos", Emoji: "🇱🇦", Phone: "+856"},
	{Code: "MM", Name: "缅甸", NameEn: "Myanmar", Emoji: "🇲🇲", Phone: "+95"},
	{Code: "BN", Name: "文莱", NameEn: "Brunei", Emoji: "🇧🇳", Phone: "+673"},
	{Code: "TL", Name: "东帝汶", NameEn: "Timor-Leste", Emoji: "🇹🇱", Phone: "+670"},
	{Code: "NP", Name: "尼泊尔", NameEn: "Nepal", Emoji: "🇳🇵", Phone: "+977"},
	{Code: "BT", Name: "不丹", NameEn: "Bhutan", Emoji: "🇧🇹", Phone: "+975"},
	{Code: "BD", Name: "孟加拉国", NameEn: "Bangladesh", Emoji: "🇧🇩", Phone: "+880"},
	{Code: "LK", Name: "斯里兰卡", NameEn: "Sri Lanka", Emoji: "🇱🇰", Phone: "+94"},
	{Code: "MV", Name: "马尔代夫", NameEn: "Maldives", Emoji: "🇲🇻", Phone: "+960"},
	{Code: "IN", Name: "印度", NameEn: "India", Emoji: "🇮🇳", Phone: "+91"},
	{Code: "PK", Name: "巴基斯坦", NameEn: "Pakistan", Emoji: "🇵🇰", Phone: "+92"},
	{Code: "KZ", Name: "哈萨克斯坦", NameEn: "Kazakhstan", Emoji: "🇰🇿", Phone: "+7"},
	{Code: "UZ", Name: "乌兹别克斯坦", NameEn: "Uzbekistan", Emoji: "🇺🇿", Phone: "+998"},
	{Code: "KG", Name: "吉尔吉斯斯坦", NameEn: "Kyrgyzstan", Emoji: "🇰🇬", Phone: "+996"},
	{Code: "TJ", Name: "塔吉克斯坦", NameEn: "Tajikistan", Emoji: "🇹🇯", Phone: "+992"},
	{Code: "TM", Name: "土库曼斯坦", NameEn: "Turkmenistan", Emoji: "🇹🇲", Phone: "+993"},
	{Code: "MN", Name: "蒙古", NameEn: "Mongolia", Emoji: "🇲🇳", Phone: "+976"},
	{Code: "AE", Name: "阿联酋", NameEn: "United Arab Emirates", Emoji: "🇦🇪", Phone: "+971"},
	{Code: "SA", Name: "沙特阿拉伯", NameEn: "Saudi Arabia", Emoji: "🇸🇦", Phone: "+966"},
	{Code: "QA", Name: "卡塔尔", NameEn: "Qatar", Emoji: "🇶🇦", Phone: "+974"},
	{Code: "KW", Name: "科威特", NameEn: "Kuwait", Emoji: "🇰🇼", Phone: "+965"},
	{Code: "OM", Name: "阿曼", NameEn: "Oman", Emoji: "🇴🇲", Phone: "+968"},
	{Code: "BH", Name: "巴林", NameEn: "Bahrain", Emoji: "🇧🇭", Phone: "+973"},
	{Code: "TR", Name: "土耳其", NameEn: "Turkey", Emoji: "🇹🇷", Phone: "+90"},
	{Code: "IL", Name: "以色列", NameEn: "Israel", Emoji: "🇮🇱", Phone: "+972"},
	{Code: "JO", Name: "约旦", NameEn: "Jordan", Emoji: "🇯🇴", Phone: "+962"},
	{Code: "IR", Name: "伊朗", NameEn: "Iran", Emoji: "🇮🇷", Phone: "+98"},
	{Code: "GB", Name: "英国", NameEn: "United Kingdom", Emoji: "🇬🇧", Phone: "+44"},
	{Code: "FR", Name: "法国", NameEn: "France", Emoji: "🇫🇷", Phone: "+33"},
	{Code: "DE", Name: "德国", NameEn: "Germany", Emoji: "🇩🇪", Phone: "+49"},
	{Code: "IT", Name: "意大利", NameEn: "Italy", Emoji: "🇮🇹", Phone: "+39"},
	{Code: "ES", Name: "西班牙", NameEn: "Spain", Emoji: "🇪🇸", Phone: "+34"},
	{Code: "PT", Name: "葡萄牙", NameEn: "Portugal", Emoji: "🇵🇹", Phone: "+351"},
	{Code: "NL", Name: "荷兰", NameEn: "Netherlands", Emoji: "🇳🇱", Phone: "+31"},
	{Code: "BE", Name: "比利时", NameEn: "Belgium", Emoji: "🇧🇪", Phone: "+32"},
	{Code: "CH", Name: "瑞士", NameEn: "Switzerland", Emoji: "🇨🇭", Phone: "+41"},
	{Code: "AT", Name: "奥地利", NameEn: "Austria", Emoji: "🇦🇹", Phone: "+43"},
	{Code: "SE", Name: "瑞典", NameEn: "Sweden", Emoji: "🇸🇪", Phone: "+46"},
	{Code: "NO", Name: "挪威", NameEn: "Norway", Emoji: "🇳🇴", Phone: "+47"},
	{Code: "DK", Name: "丹麦", NameEn: "Denmark", Emoji: "🇩🇰", Phone: "+45"},
	{Code: "FI", Name: "芬兰", NameEn: "Finland", Emoji: "🇫🇮", Phone: "+358"},
	{Code: "IS", Name: "冰岛", NameEn: "Iceland", Emoji: "🇮🇸", Phone: "+354"},
	{Code: "IE", Name: "爱尔兰", NameEn: "Ireland", Emoji: "🇮🇪", Phone: "+353"},
	{Code: "GR", Name: "希腊", NameEn: "Greece", Emoji: "🇬🇷", Phone: "+30"},
	{Code: "PL", Name: "波兰", NameEn: "Poland", Emoji: "🇵🇱", Phone: "+48"},
	{Code: "CZ", Name: "捷克", NameEn: "Czech Republic", Emoji: "🇨🇿", Phone: "+420"},
	{Code: "HU", Name: "匈牙利", NameEn: "Hungary", Emoji: "🇭🇺", Phone: "+36"},
	{Code: "RO", Name: "罗马尼亚", NameEn: "Romania", Emoji: "🇷🇴", Phone: "+40"},
	{Code: "BG", Name: "保加利亚", NameEn: "Bulgaria", Emoji: "🇧🇬", Phone: "+359"},
	{Code: "RS", Name: "塞尔维亚", NameEn: "Serbia", Emoji: "🇷🇸", Phone: "+381"},
	{Code: "HR", Name: "克罗地亚", NameEn: "Croatia", Emoji: "🇭🇷", Phone: "+385"},
	{Code: "RU", Name: "俄罗斯", NameEn: "Russia", Emoji: "🇷🇺", Phone: "+7"},
	{Code: "UA", Name: "乌克兰", NameEn: "Ukraine", Emoji: "🇺🇦", Phone: "+380"},
	{Code: "US", Name: "美国", NameEn: "United States", Emoji: "🇺🇸", Phone: "+1"},
	{Code: "CA", Name: "加拿大", NameEn: "Canada", Emoji: "🇨🇦", Phone: "+1"},
	{Code: "MX", Name: "墨西哥", NameEn: "Mexico", Emoji: "🇲🇽", Phone: "+52"},
	{Code: "AU", Name: "澳大利亚", NameEn: "Australia", Emoji: "🇦🇺", Phone: "+61"},
	{Code: "NZ", Name: "新西兰", NameEn: "New Zealand", Emoji: "🇳🇿", Phone: "+64"},
	{Code: "FJ", Name: "斐济", NameEn: "Fiji", Emoji: "🇫🇯", Phone: "+679"},
	{Code: "PG", Name: "巴布亚新几内亚", NameEn: "Papua New Guinea", Emoji: "🇵🇬", Phone: "+675"},
	{Code: "EG", Name: "埃及", NameEn: "Egypt", Emoji: "🇪🇬", Phone: "+20"},
	{Code: "ZA", Name: "南非", NameEn: "South Africa", Emoji: "🇿🇦", Phone: "+27"},
	{Code: "KE", Name: "肯尼亚", NameEn: "Kenya", Emoji: "🇰🇪", Phone: "+254"},
	{Code: "MA", Name: "摩洛哥", NameEn: "Morocco", Emoji: "🇲🇦", Phone: "+212"},
	{Code: "MU", Name: "毛里求斯", NameEn: "Mauritius", Emoji: "🇲🇺", Phone: "+230"},
	{Code: "SC", Name: "塞舌尔", NameEn: "Seychelles", Emoji: "🇸🇨", Phone: "+248"},
	{Code: "AR", Name: "阿根廷", NameEn: "Argentina", Emoji: "🇦🇷", Phone: "+54"},
	{Code: "BR", Name: "巴西", NameEn: "Brazil", Emoji: "🇧🇷", Phone: "+55"},
	{Code: "CL", Name: "智利", NameEn: "Chile", Emoji: "🇨🇱", Phone: "+56"},
	{Code: "PE", Name: "秘鲁", NameEn: "Peru", Emoji: "🇵🇪", Phone: "+51"},
	{Code: "CO", Name: "哥伦比亚", NameEn: "Colombia", Emoji: "🇨🇴", Phone: "+57"},
	{Code: "CU", Name: "古巴", NameEn: "Cuba", Emoji: "🇨🇺", Phone: "+53"},
	{Code: "CR", Name: "哥斯达黎加", NameEn: "Costa Rica", Emoji: "🇨🇷", Phone: "+506"},
}
