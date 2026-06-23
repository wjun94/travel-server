package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

	"travel-server/internal/middleware"
	"travel-server/internal/service"
	ws "travel-server/internal/ws"
	"travel-server/pkg/config"
	"travel-server/pkg/response"
)

// GetWeather 查询城市天气（高德 API）
// @Summary 获取城市天气
// @Tags 公共
// @Param city query string true "城市名称"
// @Success 200 {object} response.Response
// @Router /api/v1/weather [get]
func GetWeather(c *gin.Context) {
	city := c.Query("city")
	if city == "" {
		response.Fail(c, 400, "缺少city参数")
		return
	}
	key := config.AppConfig.AmapKey
	url := fmt.Sprintf("https://restapi.amap.com/v3/weather/weatherInfo?key=%s&city=%s&extensions=all", key, city)
	resp, err := http.Get(url)
	if err != nil {
		response.Fail(c, 500, "天气查询失败")
		return
	}
	defer resp.Body.Close()
	var weatherData interface{}
	json.NewDecoder(resp.Body).Decode(&weatherData)
	response.Success(c, weatherData)
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true }, // 允许所有来源
}

// WebSocketHandler 处理 WebSocket 连接（协同编辑 + 消息）
func WebSocketHandler(c *gin.Context) {
	token := c.Query("token")
	claims, err := middleware.ParseMiniAppToken(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"msg": "token无效"})
		return
	}
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("ws upgrade error:", err)
		return
	}
	defer conn.Close()

	userID := claims.UserID
	for {
		_, msgBytes, err := conn.ReadMessage()
		if err != nil {
			break
		}
		var msg map[string]interface{}
		if err := json.Unmarshal(msgBytes, &msg); err != nil {
			continue
		}
		action, _ := msg["action"].(string)
		switch action {
		case "join_trip":
			tripID, _ := msg["trip_id"].(string)
			ws.WsHub.Join("trip:"+tripID, conn)
		case "edit_trip":
			tripID, _ := msg["trip_id"].(string)
			tripIDUint := uint(0)
			if f, ok := msg["trip_id"].(float64); ok {
				tripIDUint = uint(f)
			}
			// 持久化编辑
			service.ApplyTripEdit(tripIDUint, msg)
			// 广播给同房间其他连接
			ws.WsHub.Broadcast("trip:"+tripID, msg, conn)
		}
		_ = userID // 实际业务可记录操作人
	}
}
