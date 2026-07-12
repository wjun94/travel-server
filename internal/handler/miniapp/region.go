package miniapp

import (
	"strings"

	"travel-server/pkg/response"

	"github.com/gin-gonic/gin"
)

// ProvinceCity 省份及城市列表
type ProvinceCity struct {
	Province string   `json:"province"` // 省/自治区/直辖市
	Cities   []string `json:"cities"`   // 城市列表
}

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
	Type     string `json:"type"`               // 类型：province/city/country
	Name     string `json:"name"`               // 名称
	Province string `json:"province,omitempty"` // 所属省份（city类型时）
	Code     string `json:"code,omitempty"`     // 国家代码（country类型时）
	Emoji    string `json:"emoji,omitempty"`    // 国旗 emoji（country类型时）
}

// GetDomesticRegions 获取国内省/市列表
// @Summary 国内省/市列表
// @Tags 公开-地区
// @Success 200 {object} response.Response{data=[]ProvinceCity}
// @Router /api/v1/regions/domestic [get]
func GetDomesticRegions(c *gin.Context) {
	response.Success(c, provincialData)
}

// GetCountries 获取境外国家列表
// @Summary 境外国家列表
// @Tags 公开-地区
// @Success 200 {object} response.Response{data=[]CountryItem}
// @Router /api/v1/regions/countries [get]
func GetCountries(c *gin.Context) {
	response.Success(c, countryData)
}

// SearchDestination 搜索目的地（城市/国家匹配）
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

	// 1. 匹配国家（中文名、英文名、代码）
	for _, country := range countryData {
		if contains(kwLower, country.Name) || contains(kwLower, country.NameEn) || contains(kwLower, country.Code) {
			results = append(results, DestinationMatch{
				Type:  "country",
				Name:  country.Name,
				Code:  country.Code,
				Emoji: country.Emoji,
			})
		}
	}

	// 2. 匹配省份
	for _, pc := range provincialData {
		if contains(kwLower, pc.Province) {
			results = append(results, DestinationMatch{
				Type: "province",
				Name: pc.Province,
			})
		}

		// 3. 匹配城市
		for _, city := range pc.Cities {
			if contains(kwLower, city) {
				results = append(results, DestinationMatch{
					Type:     "city",
					Name:     city,
					Province: pc.Province,
				})
			}
		}
	}

	response.Success(c, results)
}

// contains 判断 keyword 是否在 s 中（忽略大小写）
func contains(kwLower, s string) bool {
	return strings.Contains(strings.ToLower(s), kwLower)
}
