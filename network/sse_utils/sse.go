package sse_utils

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"sync"
	"time"
)

//
// @Author yfy2001
// @Date 2025/4/9 15:30
//

// SSEConnection 表示单个SSE连接
type SSEConnection struct {
	w       http.ResponseWriter
	flusher http.Flusher

	// 生命周期控制
	closeOnce sync.Once
	ctx       context.Context
	cancel    context.CancelFunc

	writeMu sync.Mutex  // 写锁，保证消息写入的并发安全
	logger  *log.Logger // 日志

	heartbeatTicker *time.Ticker // 心跳定时器
}

// NewConnection 创建新的SSE连接
func NewConnection(w http.ResponseWriter) (*SSEConnection, error) {
	// 设置SSE响应头
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	flusher, ok := w.(http.Flusher)
	if !ok {
		return nil, fmt.Errorf("streaming not supported")
	}

	// 发送初始空消息以建立连接
	if _, err := w.Write([]byte(":\n\n")); err != nil {
		return nil, fmt.Errorf("初始写入失败，可能客户端已断开: %w", err)
	}
	flusher.Flush()

	ctx, cancel := context.WithCancel(context.Background())
	return &SSEConnection{
		w:         w,
		flusher:   flusher,
		closeOnce: sync.Once{},
		ctx:       ctx,
		cancel:    cancel,
		logger:    log.New(os.Stdout, "[SSE] ", log.LstdFlags),
	}, nil
}

// Write 实现 io.Writer 接口
func (c *SSEConnection) Write(p []byte) (n int, err error) {
	return c.w.Write(p)
}

// SendMessage 发送SSE消息
func (c *SSEConnection) SendMessage(msg *SSEMessage) error {
	// 如果上下文已取消，直接返回错误
	select {
	case <-c.ctx.Done():
		return fmt.Errorf("连接已关闭，无法发送消息")
	default:
	}

	c.writeMu.Lock()
	defer c.writeMu.Unlock()

	data := msg.Encode()
	if _, err := c.w.Write(data); err != nil {
		c.logger.Printf("发送消息失败: %v", err)
		return err
	}
	c.flusher.Flush()
	c.logger.Printf("消息发送成功 (长度: %d字节)", len(data))
	return nil
}

// StartHeartbeat 启动心跳机制
func (c *SSEConnection) StartHeartbeat(interval time.Duration) {
	if c.heartbeatTicker != nil {
		c.heartbeatTicker.Stop()
	}

	c.heartbeatTicker = time.NewTicker(interval)
	go func() {
		defer c.heartbeatTicker.Stop()
		for {
			select {
			case <-c.ctx.Done():
				c.logger.Println("心跳协程退出")
				return
			case <-c.heartbeatTicker.C:
				if _, err := c.w.Write([]byte(":\n\n")); err != nil {
					c.logger.Printf("心跳发送失败: %v", err)
					c.Close()
					return
				}
				c.flusher.Flush()
				c.logger.Printf("心跳已发送")
			}
		}
	}()
}

// StopHeartbeat 停止心跳
func (c *SSEConnection) StopHeartbeat() {
	if c.heartbeatTicker != nil {
		c.heartbeatTicker.Stop()
		c.heartbeatTicker = nil
	}
}

// Close 关闭连接
func (c *SSEConnection) Close() {
	c.closeOnce.Do(func() {
		// 通知所有监听 ctx 的协程退出
		c.cancel()
		c.StopHeartbeat()

		// 尝试关闭底层 TCP 连接
		if hj, ok := c.w.(http.Hijacker); ok {
			if conn, _, err := hj.Hijack(); err == nil {
				_ = conn.Close()
			}
		} else if cn, ok := c.w.(net.Conn); ok {
			_ = cn.Close()
		}

		c.logger.Println("连接已关闭")
	})
}
