package sse_utils

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

//
// @Author yfy2001
// @Date 2025/4/9 15 30
//

// SSEMessage 表示符合SSE标准的消息
type SSEMessage struct {
	Event   string // 事件类型（可选）
	Data    []byte // 消息内容（必填），支持多行文本
	ID      string // 消息ID（可选）
	Comment string // 注释（可选），以冒号开头
	Retry   int    // 客户端重连时间（单位：毫秒，可选）
}

// Encode 将消息编码为SSE格式的字节流
func (sm *SSEMessage) Encode() []byte {
	var buf strings.Builder

	// 写入注释（支持多行）
	if sm.Comment != "" {
		lines := strings.Split(sm.Comment, "\n")
		for _, line := range lines {
			if line == "" {
				buf.WriteString(":\n")
			} else {
				buf.WriteString(":")
				buf.WriteString(line)
				buf.WriteString("\n")
			}
		}
	}

	// 写入事件类型
	if sm.Event != "" {
		buf.WriteString("event:")
		buf.WriteString(sm.Event)
		buf.WriteString("\n")
	}

	// 写入数据（支持多行）
	if len(sm.Data) > 0 {
		lines := bytes.Split(sm.Data, []byte{'\n'})
		for _, line := range lines {
			buf.WriteString("data:")
			buf.Write(line)
			buf.WriteString("\n")
		}
	} else {
		buf.WriteString("data:\n")
	}

	// 写入消息ID
	if sm.ID != "" {
		buf.WriteString("id:")
		buf.WriteString(sm.ID)
		buf.WriteString("\n")
	}

	// 写入重连时间
	if sm.Retry > 0 {
		buf.WriteString("retry:")
		buf.WriteString(strconv.Itoa(sm.Retry))
		buf.WriteString("\n")
	}

	// 消息以空行结束
	buf.WriteString("\n")
	return []byte(buf.String())
}

// SSEConnection 表示单个SSE连接
type SSEConnection struct {
	w          http.ResponseWriter
	flusher    http.Flusher
	logger     *log.Logger
	pingTicker *time.Ticker
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

	return &SSEConnection{
		w:       w,
		flusher: flusher,
		logger:  log.New(os.Stdout, "[SSE] ", log.LstdFlags),
	}, nil
}

func (c *SSEConnection) Write(p []byte) (n int, err error) {
	return c.w.Write(p)
}

// SendMessage 发送SSE消息
func (c *SSEConnection) SendMessage(msg *SSEMessage) error {
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
	if c.pingTicker != nil {
		c.pingTicker.Stop()
	}

	c.pingTicker = time.NewTicker(interval)
	go func() {
		for range c.pingTicker.C {
			if _, err := c.w.Write([]byte(":\n\n")); err != nil {
				c.logger.Printf("心跳发送失败: %v", err)
				return
			}
			c.flusher.Flush()
			c.logger.Printf("心跳已发送")
		}
	}()
}

// StopHeartbeat 停止心跳
func (c *SSEConnection) StopHeartbeat() {
	if c.pingTicker != nil {
		c.pingTicker.Stop()
		c.pingTicker = nil
	}
}

// Close 关闭连接
func (c *SSEConnection) Close() {
	c.StopHeartbeat()
	c.logger.Printf("连接已关闭")
}
