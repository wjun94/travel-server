package miniapp

import (
	"strings"

	"travel-server/pkg/response"

	"github.com/gin-gonic/gin"
)

// CountryItem 国家/地区
type CountryItem struct {
	Code   string `json:"code"`   // 国家代码（ISO 3166-1 alpha-2）
	Name   string `json:"name"`   // 中文名称
	NameEn string `json:"nameEn"` // 英文名称
	Emoji  string `json:"emoji"`  // 国旗 emoji
	Phone  string `json:"phone"`  // 电话区号
}

// DestinationMatch 目的地匹配结果
type DestinationMatch struct {
	Type     string `json:"type"`               // 类型：国家/省/市/区/镇
	TypeEn   string `json:"typeEn"`             // 类型英文：country/province/city/district/town
	Name     string `json:"name"`               // 名称
	Province string `json:"province,omitempty"` // 所属省份
	City     string `json:"city,omitempty"`     // 所属城市/区（district/town类型时）
	District string `json:"district,omitempty"` // 所属区（town类型时）
	Code     string `json:"code,omitempty"`     // 国家代码（country类型时）
	Emoji    string `json:"emoji,omitempty"`    // 国旗 emoji（country类型时）
}

// GetAllRegions 获取全部省市区数据（树形结构）
// @Summary 获取全国省市区数据
// @Description 一次性返回所有省市区，按省份->城市->区县树形结构
// @Tags 公共接口
// @Produce json
// @Success 200 {object} response.Response{data=[]RegionNode}
// @Router /api/v1/regions/all [get]
func GetAllRegions(c *gin.Context) {
	tree := loadRegionTree()
	response.Success(c, tree)
}

// GetCountries 获取境外国家列表
// @Summary 境外国家列表
// @Tags 公开-地区
// @Success 200 {object} response.Response{data=[]CountryItem}
// @Router /api/v1/regions/countries [get]
func GetCountries(c *gin.Context) {
	response.Success(c, countryData)
}

// SearchDestination 搜索目的地（省、市、区、镇、国家匹配）
// @Summary 搜索目的地
// @Tags 公开-地区
// @Param keyword query string true "搜索关键词"
// @Success 200 {object} response.Response{data=[]DestinationMatch}
// @Router /api/v1/destinations/search [get]
func SearchDestination(c *gin.Context) {
	keyword := strings.TrimSpace(c.Query("keyword"))
	if keyword == "" {
		response.Fail(c, 400, "请输入关键词")
		return
	}

	kwLower := strings.ToLower(keyword)
	var results []DestinationMatch
	seen := make(map[string]bool)

	// 1. 匹配国家（中文名、英文名、代码）
	for _, country := range countryData {
		if contains(kwLower, country.Name) || contains(kwLower, country.NameEn) || contains(kwLower, country.Code) {
			key := "country_" + country.Name
			if !seen[key] {
				results = append(results, DestinationMatch{
					Type:   "国家",
					TypeEn: "country",
					Name:   country.Name,
					Code:   country.Code,
					Emoji:  country.Emoji,
				})
				seen[key] = true
			}
		}
	}

	// 2. 匹配国内省、市、区、镇（使用树形数据）
	tree := loadRegionTree()
	for _, province := range tree {
		searchNode(&results, seen, kwLower, province, 0, "", "", "")
	}

	response.Success(c, results)
}

// matchesNode 判断 keyword 是否匹配节点（中文名、全拼、首字母缩写）
func matchesNode(kwLower string, node *RegionNode) bool {
	if contains(kwLower, node.Name) {
		return true
	}
	// 如果关键词不含中文，尝试匹配拼音
	if !hasChinese(kwLower) && node.Pinyin != "" {
		if contains(kwLower, node.Pinyin) || contains(kwLower, node.FirstLetters) {
			return true
		}
	}
	return false
}

// searchNode 递归搜索树形节点中的匹配项
func searchNode(results *[]DestinationMatch, seen map[string]bool, kwLower string, node RegionNode, depth int, province, city, district string) {
	switch depth {
	case 0:
		// 省份级别
		if matchesNode(kwLower, &node) {
			key := "province_" + node.Name
			if !seen[key] {
				*results = append(*results, DestinationMatch{
					Type:   "省",
					TypeEn: "province",
					Name:   node.Name,
				})
				seen[key] = true
			}
		}
		province = node.Name
	case 1:
		// 城市级别
		if matchesNode(kwLower, &node) {
			key := "city_" + node.Name
			if !seen[key] {
				*results = append(*results, DestinationMatch{
					Type:     "市",
					TypeEn:   "city",
					Name:     node.Name,
					Province: province,
				})
				seen[key] = true
			}
		}
		city = node.Name
	case 2:
		// 区/县级别
		if matchesNode(kwLower, &node) {
			key := "district_" + node.Name
			if !seen[key] {
				*results = append(*results, DestinationMatch{
					Type:     "区",
					TypeEn:   "district",
					Name:     node.Name,
					Province: province,
					City:     city,
				})
				seen[key] = true
			}
		}
		district = node.Name
	default:
		// 镇/街道级别（depth >= 3）
		if matchesNode(kwLower, &node) {
			key := "town_" + node.Name
			if !seen[key] {
				*results = append(*results, DestinationMatch{
					Type:     "镇",
					TypeEn:   "town",
					Name:     node.Name,
					Province: province,
					City:     city,
					District: district,
				})
				seen[key] = true
			}
		}
	}

	// 递归处理子节点
	for i := range node.Children {
		child := node.Children[i]
		searchNode(results, seen, kwLower, child, depth+1, province, city, district)
	}
}

// contains 判断 keyword 是否在 s 中（忽略大小写）
func contains(kwLower, s string) bool {
	return strings.Contains(strings.ToLower(s), kwLower)
}
