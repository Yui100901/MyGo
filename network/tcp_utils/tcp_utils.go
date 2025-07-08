package tcp_utils

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

//
// @Author yfy2001
// @Date 2025/7/8 13 59
//

const (
	defaultReadTimeout  = 30 * time.Second
	defaultWriteTimeout = 30 * time.Second
	defaultHeartbeat    = 15 * time.Second
	closeGracePeriod    = 5 * time.Second
)

// TCPConn TCP连接封装
type TCPConn struct {
	conn             net.Conn
	closeOnce        sync.Once
	done             chan struct{}
	readMu           sync.Mutex
	writeMu          sync.Mutex
	logger           *log.Logger
	readTimeout      time.Duration
	writeTimeout     time.Duration
	heartbeatTicker  *time.Ticker
	heartbeatMutex   sync.Mutex
	heartbeatEnabled uint32 // 原子操作
	reader           *bufio.Reader
	writer           *bufio.Writer
}

// NewTCPConn 创建新的TCP连接
func NewTCPConn(conn net.Conn) *TCPConn {
	tcp := &TCPConn{
		conn:         conn,
		done:         make(chan struct{}),
		logger:       log.New(os.Stdout, "[TCP] ", log.LstdFlags),
		readTimeout:  defaultReadTimeout,
		writeTimeout: defaultWriteTimeout,
		reader:       bufio.NewReader(conn),
		writer:       bufio.NewWriter(conn),
	}

	tcp.logger.Printf("创建新的TCP连接 (Local: %s, Remote: %s)",
		tcp.LocalAddr(), tcp.RemoteAddr())
	return tcp
}

// Dial 连接到TCP服务器
func Dial(addr string, timeout time.Duration) (*TCPConn, error) {
	dialer := net.Dialer{Timeout: timeout}
	conn, err := dialer.Dial("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("连接失败: %w", err)
	}
	return NewTCPConn(conn), nil
}

// ListenAndServe 监听TCP连接
func ListenAndServe(addr string, handler func(*TCPConn)) error {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("监听失败: %w", err)
	}
	defer listener.Close()

	log.Printf("TCP服务器监听于 %s", addr)

	for {
		conn, err := listener.Accept()
		if err != nil {
			var ne net.Error
			if errors.As(err, &ne) && ne.Temporary() {
				time.Sleep(100 * time.Millisecond)
				continue
			}
			return fmt.Errorf("接受连接失败: %w", err)
		}

		tcpConn := NewTCPConn(conn)
		go handler(tcpConn)
	}
}

// SetLogger 设置自定义日志记录器
func (t *TCPConn) SetLogger(logger *log.Logger) {
	t.logger = logger
}

// SetTimeouts 设置读写超时
func (t *TCPConn) SetTimeouts(readTimeout, writeTimeout time.Duration) {
	t.readTimeout = readTimeout
	t.writeTimeout = writeTimeout
}

// StartHeartbeat 启动心跳机制
func (t *TCPConn) StartHeartbeat(interval time.Duration, heartbeatMsg []byte) {
	t.heartbeatMutex.Lock()
	defer t.heartbeatMutex.Unlock()

	if t.IsClosed() {
		t.logger.Println("尝试启动心跳但连接已关闭")
		return
	}

	// 停止现有心跳
	if t.heartbeatTicker != nil {
		t.heartbeatTicker.Stop()
	}

	if interval <= 0 {
		atomic.StoreUint32(&t.heartbeatEnabled, 0)
		t.logger.Println("心跳已禁用")
		return
	}

	t.heartbeatTicker = time.NewTicker(interval)
	atomic.StoreUint32(&t.heartbeatEnabled, 1)
	t.logger.Printf("启动心跳，间隔: %v (Remote: %s)", interval, t.RemoteAddr())

	go t.heartbeatLoop(heartbeatMsg)
}

// StopHeartbeat 停止心跳机制
func (t *TCPConn) StopHeartbeat() {
	t.heartbeatMutex.Lock()
	defer t.heartbeatMutex.Unlock()

	if t.heartbeatTicker != nil {
		t.heartbeatTicker.Stop()
		t.heartbeatTicker = nil
		t.logger.Printf("心跳已停止 (Remote: %s)", t.RemoteAddr())
	}
	atomic.StoreUint32(&t.heartbeatEnabled, 0)
}

// heartbeatLoop 心跳循环
func (t *TCPConn) heartbeatLoop(heartbeatMsg []byte) {
	t.logger.Printf("心跳协程启动 (Remote: %s)", t.RemoteAddr())
	defer t.logger.Printf("心跳协程退出 (Remote: %s)", t.RemoteAddr())

	for {
		select {
		case <-t.done:
			return
		case <-t.heartbeatTicker.C:
			if atomic.LoadUint32(&t.heartbeatEnabled) == 0 {
				return
			}

			if err := t.Write(heartbeatMsg); err != nil {
				t.logger.Printf("发送心跳失败: %v (Remote: %s)", err, t.RemoteAddr())
				t.safeClose()
				return
			}
			t.logger.Printf("心跳已发送 (Remote: %s)", t.RemoteAddr())
		}
	}
}

// Read 从连接读取数据
func (t *TCPConn) Read() ([]byte, error) {
	t.readMu.Lock()
	defer t.readMu.Unlock()

	select {
	case <-t.done:
		return nil, io.EOF
	default:
		if t.readTimeout > 0 {
			_ = t.conn.SetReadDeadline(time.Now().Add(t.readTimeout))
		}

		// 读取消息长度前缀 (4字节)
		lenBuf := make([]byte, 4)
		if _, err := io.ReadFull(t.reader, lenBuf); err != nil {
			if err == io.EOF {
				t.logger.Printf("对端关闭连接 (Remote: %s)", t.RemoteAddr())
			} else {
				t.logger.Printf("读取长度前缀失败: %v (Remote: %s)", err, t.RemoteAddr())
			}
			return nil, err
		}

		// 解析消息长度
		length := uint32(lenBuf[0])<<24 | uint32(lenBuf[1])<<16 | uint32(lenBuf[2])<<8 | uint32(lenBuf[3])
		if length > 10*1024*1024 { // 限制10MB
			return nil, fmt.Errorf("消息过大: %d bytes", length)
		}

		// 读取消息体
		data := make([]byte, length)
		if _, err := io.ReadFull(t.reader, data); err != nil {
			t.logger.Printf("读取消息体失败: %v (Remote: %s)", err, t.RemoteAddr())
			return nil, err
		}

		t.logger.Printf("收到消息 (长度: %d字节, Remote: %s)", length, t.RemoteAddr())
		return data, nil
	}
}

// Write 向连接写入数据
func (t *TCPConn) Write(data []byte) error {
	t.writeMu.Lock()
	defer t.writeMu.Unlock()

	select {
	case <-t.done:
		return io.ErrClosedPipe
	default:
		if t.writeTimeout > 0 {
			_ = t.conn.SetWriteDeadline(time.Now().Add(t.writeTimeout))
		}

		// 添加长度前缀
		length := len(data)
		lenBuf := []byte{
			byte(length >> 24),
			byte(length >> 16),
			byte(length >> 8),
			byte(length),
		}

		// 写入长度前缀
		if _, err := t.writer.Write(lenBuf); err != nil {
			t.logger.Printf("写入长度前缀失败: %v (Remote: %s)", err, t.RemoteAddr())
			return err
		}

		// 写入消息体
		if _, err := t.writer.Write(data); err != nil {
			t.logger.Printf("写入消息体失败: %v (Remote: %s)", err, t.RemoteAddr())
			return err
		}

		// 刷新缓冲区
		if err := t.writer.Flush(); err != nil {
			t.logger.Printf("刷新缓冲区失败: %v (Remote: %s)", err, t.RemoteAddr())
			return err
		}

		t.logger.Printf("消息发送成功 (长度: %d字节, Remote: %s)", length, t.RemoteAddr())
		return nil
	}
}

// OnMessage 持续处理消息
func (t *TCPConn) OnMessage(handleFunc func([]byte)) error {
	if handleFunc == nil {
		return errors.New("处理函数不能为空")
	}

	defer t.safeClose() // 确保退出时清理资源

	t.logger.Printf("开始接收消息 (Remote: %s)", t.RemoteAddr())

	for {
		select {
		case <-t.done:
			t.logger.Printf("连接已关闭，停止接收消息 (Remote: %s)", t.RemoteAddr())
			return nil
		default:
			data, err := t.Read()
			if err != nil {
				if errors.Is(err, io.EOF) {
					t.logger.Printf("连接正常关闭 (Remote: %s)", t.RemoteAddr())
					return nil
				}
				t.logger.Printf("读取消息错误: %v (Remote: %s)", err, t.RemoteAddr())
				return fmt.Errorf("读取错误: %w", err)
			}

			handleFunc(data)
		}
	}
}

// AsyncRead 异步读取消息
func (t *TCPConn) AsyncRead(handleFunc func([]byte, error)) {
	go func() {
		data, err := t.Read()
		if handleFunc != nil {
			handleFunc(data, err)
		}
	}()
}

// AsyncWrite 异步写入消息
func (t *TCPConn) AsyncWrite(data []byte, callback func(error)) {
	go func() {
		err := t.Write(data)
		if callback != nil {
			callback(err)
		}
	}()
}

// Close 安全关闭连接
func (t *TCPConn) Close() {
	t.logger.Printf("关闭连接 (Remote: %s)", t.RemoteAddr())
	t.safeClose()
}

func (t *TCPConn) safeClose() {
	t.closeOnce.Do(func() {
		// 先停止心跳
		t.StopHeartbeat()

		close(t.done) // 通知所有消费者

		// 优雅关闭连接
		if tcpConn, ok := t.conn.(*net.TCPConn); ok {
			_ = tcpConn.CloseRead()
		}

		// 设置关闭超时
		time.AfterFunc(closeGracePeriod, func() {
			_ = t.conn.Close()
		})

		// 尝试发送关闭通知
		_, _ = t.conn.Write([]byte{0xFF, 0xFF, 0xFF, 0xFF}) // 特殊关闭标记

		t.logger.Printf("连接已完全关闭 (Remote: %s)", t.RemoteAddr())
	})
}

// Done 获取关闭信号通道
func (t *TCPConn) Done() <-chan struct{} {
	return t.done
}

// RemoteAddr 获取对端地址
func (t *TCPConn) RemoteAddr() string {
	return t.conn.RemoteAddr().String()
}

// LocalAddr 获取本地地址
func (t *TCPConn) LocalAddr() string {
	return t.conn.LocalAddr().String()
}

// IsClosed 检查连接是否已关闭
func (t *TCPConn) IsClosed() bool {
	select {
	case <-t.done:
		return true
	default:
		return false
	}
}

// SetKeepAlive 设置TCP KeepAlive
func (t *TCPConn) SetKeepAlive(keepalive bool) error {
	if tcpConn, ok := t.conn.(*net.TCPConn); ok {
		return tcpConn.SetKeepAlive(keepalive)
	}
	return errors.New("连接不是TCP连接")
}

// SetKeepAlivePeriod 设置KeepAlive间隔
func (t *TCPConn) SetKeepAlivePeriod(d time.Duration) error {
	if tcpConn, ok := t.conn.(*net.TCPConn); ok {
		return tcpConn.SetKeepAlivePeriod(d)
	}
	return errors.New("连接不是TCP连接")
}
