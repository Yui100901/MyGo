package websocket_utils

import (
	"errors"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"os"
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
	logger      *log.Logger   // 日志记录器
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
		logger:      log.New(os.Stdout, "[WS] ", log.LstdFlags),
	}
}

// SetLogger 设置自定义日志记录器
func (ws *WebSocket) SetLogger(logger *log.Logger) {
	ws.logger = logger
}

// SetReadTimeout 设置读取超时
func (ws *WebSocket) SetReadTimeout(timeout time.Duration) {
	ws.readTimeout = timeout
}

// OnMessage 持续处理消息
func (ws *WebSocket) OnMessage(handleFunc func(messageType int, payload []byte)) error {
	if handleFunc == nil {
		return errors.New("handle function is nil")
	}

	defer ws.safeClose() // 确保退出时清理资源

	ws.logger.Printf("开始接收消息 (Remote: %s)", ws.RemoteAddr())

	for {
		select {
		case <-ws.done:
			ws.logger.Printf("连接已关闭，停止接收消息 (Remote: %s)", ws.RemoteAddr())
			return nil
		default:
			if ws.readTimeout > 0 {
				_ = ws.conn.SetReadDeadline(time.Now().Add(ws.readTimeout))
			}

			msgType, message, err := ws.conn.ReadMessage()
			if err != nil {
				if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
					ws.logger.Printf("连接正常关闭 (Remote: %s)", ws.RemoteAddr())
					return nil
				}
				ws.logger.Printf("读取消息错误: %v (Remote: %s)", err, ws.RemoteAddr())
				return fmt.Errorf("read error: %w", err)
			}

			ws.logger.Printf("收到消息 (类型: %d, 长度: %d字节, Remote: %s)",
				msgType, len(message), ws.RemoteAddr())

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
		ws.logger.Printf("尝试发送消息但连接已关闭 (Remote: %s)", ws.RemoteAddr())
		return websocket.ErrCloseSent
	default:
		_ = ws.conn.SetWriteDeadline(time.Now().Add(writeTimeout))
		err := ws.conn.WriteMessage(messageType, payload)
		if err != nil {
			ws.logger.Printf("发送消息失败: %v (Remote: %s)", err, ws.RemoteAddr())
			return fmt.Errorf("write error: %w", err)
		}

		ws.logger.Printf("消息发送成功 (类型: %d, 长度: %d字节, Remote: %s)",
			messageType, len(payload), ws.RemoteAddr())
		return nil
	}
}

// Close 安全关闭连接
func (ws *WebSocket) Close() {
	ws.logger.Printf("关闭连接 (Remote: %s)", ws.RemoteAddr())
	ws.safeClose()
}

func (ws *WebSocket) safeClose() {
	ws.closeOnce.Do(func() {
		close(ws.done) // 通知所有消费者

		// 发送关闭帧
		err := ws.conn.WriteControl(
			websocket.CloseMessage,
			websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""),
			time.Now().Add(closeTimeout),
		)
		if err != nil {
			ws.logger.Printf("发送关闭帧失败: %v (Remote: %s)", err, ws.RemoteAddr())
		} else {
			ws.logger.Printf("关闭帧已发送 (Remote: %s)", ws.RemoteAddr())
		}

		// 安全关闭底层连接
		err = ws.conn.Close()
		if err != nil {
			ws.logger.Printf("关闭底层连接失败: %v (Remote: %s)", err, ws.RemoteAddr())
		} else {
			ws.logger.Printf("连接已完全关闭 (Remote: %s)", ws.RemoteAddr())
		}
	})
}

// Done 获取关闭信号通道
func (ws *WebSocket) Done() <-chan struct{} {
	return ws.done
}

// RemoteAddr 获取对端地址
func (ws *WebSocket) RemoteAddr() string {
	return ws.conn.RemoteAddr().String()
}

// LocalAddr 获取本地地址
func (ws *WebSocket) LocalAddr() string {
	return ws.conn.LocalAddr().String()
}

// IsClosed 检查连接是否已关闭
func (ws *WebSocket) IsClosed() bool {
	select {
	case <-ws.done:
		return true
	default:
		return false
	}
}
