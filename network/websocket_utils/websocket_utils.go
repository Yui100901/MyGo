package websocket_utils

import (
	"context"
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
var WSServer = &websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// 生产环境应替换为实际源检查逻辑
		return true
	},
}

var DefaultDialer = &websocket.Dialer{
	Proxy:            http.ProxyFromEnvironment,
	HandshakeTimeout: 45 * time.Second,
}

const (
	writeTimeout = 3 * time.Second
	closeTimeout = 3 * time.Second
)

type WebSocket struct {
	conn *websocket.Conn //底层连接

	//生命周期
	closeOnce sync.Once
	ctx       context.Context
	cancel    context.CancelFunc

	writeMu     sync.Mutex    //写锁
	readTimeout time.Duration // 可配置的读取超时
	logger      *log.Logger   // 日志记录器

	// 心跳相关字段
	heartbeatInterval time.Duration
	heartbeatTicker   *time.Ticker
	heartbeatMutex    sync.Mutex
}

// NewWebSocketByDial 主动建立连接
func NewWebSocketByDial(dialer *websocket.Dialer, url string, requestHeader http.Header) (*WebSocket, error) {
	if dialer == nil {
		dialer = DefaultDialer
	}
	conn, _, err := dialer.Dial(url, requestHeader)
	if err != nil {
		return nil, fmt.Errorf("ws dial failed: %w", err)
	}
	return NewWebSocket(conn), nil
}

// NewWebSocketByUpgrade 升级HTTP连接
func NewWebSocketByUpgrade(upGrader *websocket.Upgrader, w http.ResponseWriter, r *http.Request, responseHeader http.Header) (*WebSocket, error) {
	if upGrader == nil {
		upGrader = WSServer
	}
	conn, err := upGrader.Upgrade(w, r, responseHeader)
	if err != nil {
		return nil, fmt.Errorf("ws upgrade failed: %w", err)
	}
	return NewWebSocket(conn), nil
}

func NewWebSocket(conn *websocket.Conn) *WebSocket {
	ctx, cancel := context.WithCancel(context.Background())
	return &WebSocket{
		conn:        conn,
		ctx:         ctx,
		cancel:      cancel,
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

// StartHeartbeat 启动心跳机制
func (ws *WebSocket) StartHeartbeat(interval time.Duration) {
	ws.heartbeatMutex.Lock()
	defer ws.heartbeatMutex.Unlock()

	// 如果已经存在心跳定时器，先停止
	if ws.heartbeatTicker != nil {
		ws.heartbeatTicker.Stop()
	}

	ws.heartbeatInterval = interval
	if interval <= 0 {
		ws.logger.Println("心跳已禁用")
		return
	}

	ws.heartbeatTicker = time.NewTicker(interval)
	ws.logger.Printf("启动心跳，间隔: %v (Remote: %s)", interval, ws.RemoteAddr())

	// 设置Pong处理器
	ws.conn.SetPongHandler(func(appData string) error {
		ws.logger.Printf("收到Pong (Remote: %s)", ws.RemoteAddr())
		return nil
	})

	go ws.heartbeatLoop()
}

// StopHeartbeat 停止心跳机制
func (ws *WebSocket) StopHeartbeat() {
	ws.heartbeatMutex.Lock()
	defer ws.heartbeatMutex.Unlock()

	if ws.heartbeatTicker != nil {
		ws.heartbeatTicker.Stop()
		ws.heartbeatTicker = nil
		ws.logger.Printf("心跳已停止 (Remote: %s)", ws.RemoteAddr())
	}
}

// heartbeatLoop 心跳循环
func (ws *WebSocket) heartbeatLoop() {
	lastResponse := time.Now()

	for {
		select {
		case <-ws.ctx.Done():
			ws.logger.Printf("心跳协程退出 (Remote: %s)", ws.RemoteAddr())
			return
		case t := <-ws.heartbeatTicker.C:
			// 检查上次响应时间
			if time.Since(lastResponse) > ws.heartbeatInterval*2 {
				ws.logger.Printf("心跳超时，未收到响应 (Remote: %s)", ws.RemoteAddr())
				ws.safeClose()
				return
			}

			// 发送Ping
			ws.writeMu.Lock()
			err := ws.conn.WriteControl(websocket.PingMessage, []byte{}, t.Add(writeTimeout))
			ws.writeMu.Unlock()

			if err != nil {
				ws.logger.Printf("发送心跳失败: %v (Remote: %s)", err, ws.RemoteAddr())
				ws.safeClose()
				return
			}

			ws.logger.Printf("发送心跳Ping (Remote: %s)", ws.RemoteAddr())
			lastResponse = t
		}
	}
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
		case <-ws.ctx.Done():
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
	case <-ws.ctx.Done():
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
		// 先停止心跳
		ws.StopHeartbeat()

		ws.cancel() // 停止上下文

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
	case <-ws.ctx.Done():
		return true
	default:
		return false
	}
}
