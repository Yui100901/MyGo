package sse_utils

import (
	"bytes"
	"strconv"
	"strings"
)

//
// @Author yfy2001
// @Date 2025/8/19 10 24
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
