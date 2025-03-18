package websocket_utils

import (
	"github.com/Yui100901/MyGo/log_utils"
	"github.com/gorilla/websocket"
	"net/http"
	"sync"
)

//
// @Author yfy2001
// @Date 2025/1/16 09 56
//

// WSServer 是一个全局的 WebSocket 升级器
var WSServer = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// 验证
		return true
	},
}

// WebSocket 表示一个 WebSocket 连接
type WebSocket struct {
	Conn *websocket.Conn // 底层 WebSocket 连接对象
	Done chan struct{}
	mu   sync.Mutex
}

// NewWebSocketByUpgrade 通过 HTTP 请求升级为 WebSocket
func NewWebSocketByUpgrade(w http.ResponseWriter, r *http.Request, responseHeader http.Header) (*WebSocket, error) {
	conn, err := WSServer.Upgrade(w, r, responseHeader)
	if err != nil {
		return nil, err
	}
	return &WebSocket{
		Conn: conn,
		Done: make(chan struct{}),
		mu:   sync.Mutex{},
	}, nil
}

// NewWebSocketByDial 通过 WebSocket 地址进行连接
func NewWebSocketByDial(url string, requestHeader http.Header) (*WebSocket, error) {
	conn, _, err := websocket.DefaultDialer.Dial(url, requestHeader)
	if err != nil {
		return nil, err
	}
	return &WebSocket{
		Conn: conn,
		Done: make(chan struct{}),
		mu:   sync.Mutex{},
	}, nil
}

// OnMessage 注册一个处理消息的回调函数
func (ws *WebSocket) OnMessage(handleFunc func([]byte)) {
	defer close(ws.Done)
	for {
		_, message, err := ws.Conn.ReadMessage()
		if err != nil {
			log_utils.Error.Println("WebSocket Read ERROR:", err)
			return
		}
		if handleFunc != nil {
			handleFunc(message)
		}
		log_utils.Info.Printf("WebSocket Receive: %s", message)
	}
}

// SendMessage 发送一条消息
func (ws *WebSocket) SendMessage(messageType int, data func() []byte) error {
	ws.mu.Lock()
	defer ws.mu.Unlock()
	return ws.Conn.WriteMessage(messageType, data())
}

// Close 关闭 WebSocket 连接
func (ws *WebSocket) Close() {
	ws.mu.Lock()
	defer ws.mu.Unlock()
	ws.Conn.Close()
}
