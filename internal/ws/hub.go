// Package ws 管理 WebSocket 房间和广播
package ws

import (
	"sync"

	"github.com/gorilla/websocket"
)

// Hub 维护房间与连接的关系
type Hub struct {
	rooms map[string]map[*websocket.Conn]bool // roomID -> 连接集合
	mu    sync.Mutex
}

var WsHub = NewHub()

// NewHub 创建 Hub 实例
func NewHub() *Hub {
	return &Hub{rooms: make(map[string]map[*websocket.Conn]bool)}
}

// Join 将连接加入指定房间
func (h *Hub) Join(room string, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.rooms[room] == nil {
		h.rooms[room] = make(map[*websocket.Conn]bool)
	}
	h.rooms[room][conn] = true
}

// Leave 将连接从房间移除
func (h *Hub) Leave(room string, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if conns, ok := h.rooms[room]; ok {
		delete(conns, conn)
		if len(conns) == 0 {
			delete(h.rooms, room)
		}
	}
}

// Broadcast 向房间内所有连接广播消息（可排除发送者）
func (h *Hub) Broadcast(room string, msg interface{}, exclude *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	conns := h.rooms[room]
	for conn := range conns {
		if conn != exclude {
			conn.WriteJSON(msg)
		}
	}
}

// JoinUser 将连接加入用户专属房间（用于服务端实时通知推送）
func (h *Hub) JoinUser(userID string, conn *websocket.Conn) {
	h.Join("user:"+userID, conn)
}

// LeaveUser 将连接移出用户专属房间
func (h *Hub) LeaveUser(userID string, conn *websocket.Conn) {
	h.Leave("user:"+userID, conn)
}

// PushToUser 向用户专属房间推送消息（用户不在线时静默丢弃）
func (h *Hub) PushToUser(userID string, msg interface{}) {
	h.Broadcast("user:"+userID, msg, nil)
}
