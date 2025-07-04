package websocket_utils

import (
	"errors"
	"fmt"
	"github.com/gorilla/websocket"
	"net/http"
	"sync"
	"time"
)

// WSServer 生产环境应配置严格的源检查
var WSServer = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

const (
	writeTimeout = 3 * time.Second
	closeTimeout = 3 * time.Second
)

type WebSocket struct {
	conn        *websocket.Conn
	closeOnce   sync.Once
	done        chan struct{}
	writeMu     sync.Mutex
	readTimeout time.Duration // 可配置的读取超时
}

// NewWebSocketByUpgrade 升级HTTP连接
func NewWebSocketByUpgrade(w http.ResponseWriter, r *http.Request, responseHeader http.Header) (*WebSocket, error) {
	conn, err := WSServer.Upgrade(w, r, responseHeader)
	if err != nil {
		return nil, fmt.Errorf("ws upgrade failed: %w", err)
	}
	return newWebSocket(conn), nil
}

// NewWebSocketByDial 主动建立连接
func NewWebSocketByDial(url string, requestHeader http.Header) (*WebSocket, error) {
	dialer := websocket.DefaultDialer
	dialer.HandshakeTimeout = 10 * time.Second // 添加握手超时
	conn, _, err := dialer.Dial(url, requestHeader)
	if err != nil {
		return nil, fmt.Errorf("ws dial failed: %w", err)
	}
	return newWebSocket(conn), nil
}

func newWebSocket(conn *websocket.Conn) *WebSocket {
	return &WebSocket{
		conn:        conn,
		done:        make(chan struct{}),
		readTimeout: 0, // 默认无超时
	}
}

// SetReadTimeout 设置读取超时(可选)
func (ws *WebSocket) SetReadTimeout(timeout time.Duration) {
	ws.readTimeout = timeout
}

// OnMessage 处理消息(支持优雅退出)
func (ws *WebSocket) OnMessage(handleFunc func(messageType int, payload []byte)) error {
	if handleFunc == nil {
		return errors.New("handle function is nil")
	}

	defer ws.safeClose() // 确保退出时清理资源

	for {
		select {
		case <-ws.done:
			return nil
		default:
			if ws.readTimeout > 0 {
				_ = ws.conn.SetReadDeadline(time.Now().Add(ws.readTimeout))
			}

			msgType, message, err := ws.conn.ReadMessage()
			if err != nil {
				if websocket.IsCloseError(err,
					websocket.CloseNormalClosure,
					websocket.CloseGoingAway) {
					return nil
				}
				return fmt.Errorf("read error: %w", err)
			}
			handleFunc(msgType, message)
		}
	}
}

// SendMessage 线程安全的消息发送
func (ws *WebSocket) SendMessage(messageType int, payload []byte) error {
	ws.writeMu.Lock()
	defer ws.writeMu.Unlock()

	select {
	case <-ws.done:
		return websocket.ErrCloseSent
	default:
		_ = ws.conn.SetWriteDeadline(time.Now().Add(writeTimeout))
		return ws.conn.WriteMessage(messageType, payload)
	}
}

// Close 安全关闭连接
func (ws *WebSocket) Close() {
	ws.safeClose()
}

func (ws *WebSocket) safeClose() {
	ws.closeOnce.Do(func() {
		close(ws.done) // 通知所有消费者

		// 发送关闭帧(忽略错误)
		_ = ws.conn.WriteControl(
			websocket.CloseMessage,
			websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""),
			time.Now().Add(closeTimeout),
		)

		// 安全关闭底层连接
		_ = ws.conn.Close()
	})
}

// Done 获取关闭信号
func (ws *WebSocket) Done() <-chan struct{} {
	return ws.done
}

// RemoteAddr 获取对端地址
func (ws *WebSocket) RemoteAddr() string {
	return ws.conn.RemoteAddr().String()
}
