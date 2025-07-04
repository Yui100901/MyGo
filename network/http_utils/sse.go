package http_utils

import (
	"bytes"
	"net/http"
	"strconv"
	"strings"
	"sync"
)

//
// @Author yfy2001
// @Date 2025/4/9 15 30
//

// SSEWriter 表示一个 SSE 客户端连接
type SSEWriter struct {
	mu      sync.RWMutex
	flusher http.Flusher
	writer  http.ResponseWriter
	closed  bool
}

// NewSSEHandler 返回一个 HTTP Handler，用于升级连接为 SSE
func NewSSEHandler(onClient func(client *SSEWriter)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 设置必要的头部
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("Access-Control-Allow-Origin", "*") // 可选：支持跨域

		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
			return
		}

		client := &SSEWriter{
			writer:  w,
			flusher: flusher,
		}

		// 通知外部处理函数
		onClient(client)

		// 等待连接关闭
		<-r.Context().Done()
		client.Close()
	}
}

// Send 发送一条 SSE 消息
func (c *SSEWriter) Send(msg SSEMessage) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.closed {
		return nil
	}

	if _, err := c.writer.Write([]byte(msg.Encode())); err != nil {
		return err
	}
	c.flusher.Flush()
	return nil
}

// Close 手动关闭连接
func (c *SSEWriter) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.closed = true
}

// SSEMessage SSE标准消息
type SSEMessage struct {
	Event   string // 事件类型（可选）
	Data    []byte // 消息内容（必填），支持多行文本
	ID      string // 消息ID（可选）
	Comment string // 注释（可选），以冒号开头
	Retry   int    // 客户端重连时间（单位：毫秒，可选）
}

func (sm *SSEMessage) Encode() []byte {
	// 预分配缓冲区大小 (基础大小 + 各字段预估长度)
	bufSize := 64 + len(sm.Comment) + len(sm.Event) + len(sm.Data) + len(sm.ID)
	var buf strings.Builder
	buf.Grow(bufSize)

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
