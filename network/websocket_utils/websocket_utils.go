package websocket_utils

import (
	"github.com/Yui100901/MyGo/log_utils"
	"github.com/gorilla/websocket"
	"net/http"
	"sync"
	"time"
)

// WSServer 全局WebSocket升级器
var WSServer = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// 生产环境应替换为实际源检查逻辑
		return true
	},
}

// WebSocket 封装WebSocket连接
type WebSocket struct {
	conn      *websocket.Conn
	closeOnce sync.Once     // 确保关闭操作幂等性
	done      chan struct{} // 连接关闭通知通道
	writeMu   sync.Mutex    // 写操作互斥锁
}

// NewWebSocketByUpgrade 升级HTTP连接为WebSocket
func NewWebSocketByUpgrade(w http.ResponseWriter, r *http.Request, responseHeader http.Header) (*WebSocket, error) {
	conn, err := WSServer.Upgrade(w, r, responseHeader)
	if err != nil {
		return nil, err
	}
	return newWebSocket(conn), nil
}

// NewWebSocketByDial 主动建立WebSocket连接
func NewWebSocketByDial(url string, requestHeader http.Header) (*WebSocket, error) {
	conn, _, err := websocket.DefaultDialer.Dial(url, requestHeader)
	if err != nil {
		return nil, err
	}
	return newWebSocket(conn), nil
}

// newWebSocket 内部构造函数
func newWebSocket(conn *websocket.Conn) *WebSocket {
	return &WebSocket{
		conn: conn,
		done: make(chan struct{}),
	}
}

// OnMessage 持续处理消息（阻塞式）
func (ws *WebSocket) OnMessage(handleFunc func(messageType int, payload []byte)) {
	defer ws.safeClose() // 确保退出时关闭资源

	for {
		select {
		case <-ws.done:
			return // 已关闭连接
		default:
			msgType, message, err := ws.conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseNormalClosure) {
					log_utils.Error.Printf("WebSocket read error: %v", err)
				}
				return
			}
			log_utils.Info.Printf("WebSocket Receive (%d): %s", msgType, message)

			if handleFunc != nil {
				handleFunc(msgType, message)
			}
		}
	}
}

// SendMessage 安全发送消息
func (ws *WebSocket) SendMessage(messageType int, payload []byte) error {
	ws.writeMu.Lock()
	defer ws.writeMu.Unlock()

	select {
	case <-ws.done:
		return websocket.ErrCloseSent // 连接已关闭
	default:
		return ws.conn.WriteMessage(messageType, payload)
	}
}

// Close 安全关闭连接
func (ws *WebSocket) Close() {
	ws.safeClose()
}

// safeClose 内部关闭方法（幂等）
func (ws *WebSocket) safeClose() {
	ws.closeOnce.Do(func() {
		close(ws.done) // 通知所有协程

		// 发送标准关闭帧
		_ = ws.conn.WriteControl(
			websocket.CloseMessage,
			websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""),
			time.Now().Add(3*time.Second),
		)

		ws.conn.Close() // 关闭底层连接
	})
}

// Done 获取关闭通知通道
func (ws *WebSocket) Done() <-chan struct{} {
	return ws.done
}
