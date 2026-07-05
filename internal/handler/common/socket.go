package common

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

	"travel-server/internal/middleware"
	"travel-server/internal/service"
	ws "travel-server/internal/ws"
)

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
			tripID, _ := msg["tripId"].(string)
			ws.WsHub.Join("trip:"+tripID, conn)
		case "edit_trip":
			tripID, _ := msg["tripId"].(string)
			// 持久化编辑
			service.ApplyTripEdit(tripID, msg)
			// 广播给同房间其他连接
			ws.WsHub.Broadcast("trip:"+tripID, msg, conn)
		}
		_ = userID // 实际业务可记录操作人
	}
}
