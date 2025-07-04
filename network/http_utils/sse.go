package http_utils

import (
	"fmt"
	"strings"
)

//
// @Author yfy2001
// @Date 2025/4/9 15 30
//

// SSEMessage SSE标准消息
type SSEMessage struct {
	Event   string // 事件类型（可选）
	Data    []byte // 消息内容（必填），支持多行文本
	ID      string // 消息ID（可选）
	Comment string // 注释（可选），以冒号开头
	Retry   int    // 客户端重连时间（单位：毫秒，可选）
}

func (sm *SSEMessage) Encode() string {

	var buf strings.Builder

	// 写入注释（可选）
	if sm.Comment != "" {
		fmt.Fprintf(&buf, ": %s\n", sm.Comment)
	}

	// 写入事件类型（可选）
	if sm.Event != "" {
		fmt.Fprintf(&buf, "event: %s\n", sm.Event)
	}

	if sm.Data != nil {
		fmt.Fprintf(&buf, "data: %s\n", string(sm.Data))

	} else {
		fmt.Fprintf(&buf, "data: null\n")
	}

	// 写入消息ID（可选）
	if sm.ID != "" {
		fmt.Fprintf(&buf, "id: %s\n", sm.ID)
	}

	// 写入重连时间（可选）
	if sm.Retry > 0 {
		fmt.Fprintf(&buf, "retry: %d\n", sm.Retry)
	}

	buf.WriteByte('\n') // SSE 消息以空行结束
	return buf.String()
}
