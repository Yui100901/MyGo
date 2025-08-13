package mq

import (
	"fmt"
	"time"
)

//
// @Author yfy2001
// @Date 2025/7/25 15 17
//

//// MessageStatus 消息状态枚举
//type MessageStatus int
//
//const (
//	MessageStatusPending    MessageStatus = iota // 待处理
//	MessageStatusDelivering                      // 投递中（等待客户端确认）
//	MessageStatusAcked                           // 已确认
//	MessageStatusFailed                          // 处理失败
//	MessageStatusDead                            // 死信
//)
//
//// String 返回消息状态的字符串表示
//func (ms MessageStatus) String() string {
//	switch ms {
//	case MessageStatusPending:
//		return "PENDING"
//	case MessageStatusDelivering:
//		return "DELIVERING"
//	case MessageStatusAcked:
//		return "ACKED"
//	case MessageStatusFailed:
//		return "FAILED"
//	case MessageStatusDead:
//		return "DEAD"
//	default:
//		return "UNKNOWN"
//	}
//}

const defaultMessageTTL = 1 * time.Hour

// Message 消息结构体，包含消息的所有属性和元数据
type Message struct {
	ID       string                 `json:"id"`                  // 消息唯一标识符
	Topic    Topic                  `json:"topic"`               // 消息主题
	SenderID string                 `json:"sender_id,omitempty"` // 发送方客户端ID
	Payload  []byte                 `json:"payload"`             // 消息载荷
	Headers  map[string]string      `json:"headers,omitempty"`   // 消息头部信息
	Metadata map[string]interface{} `json:"metadata,omitempty"`  // 扩展元数据
	//Status      MessageStatus          `json:"status"`                 // 消息状态
	Delay     time.Duration `json:"delay,omitempty"` // 发送延迟
	CreatedAt time.Time     `json:"created_at"`      // 创建时间
	DeliverAt time.Time     `json:"delivered_at"`    // 发送时间
	TTL       time.Duration `json:"ttl"`             // 存活时间
	ExpiresAt time.Time     `json:"expires_at"`      // 过期时间
}

func NewMessage(topic Topic, payload []byte) *Message {
	msgID := fmt.Sprintf("msg_%d", time.Now().UnixNano())
	now := time.Now()
	msg := &Message{
		ID:      msgID,
		Topic:   topic,
		Payload: payload,
		//Status:  MessageStatusPending,
		Delay:     0,
		CreatedAt: now,
		DeliverAt: now,
		TTL:       defaultMessageTTL,
		ExpiresAt: now.Add(defaultMessageTTL),
	}
	return msg
}

func (m *Message) SetDelay(delay time.Duration) {
	m.Delay = delay
	m.DeliverAt = m.CreatedAt.Add(delay)
}

func (m *Message) SetTTL(ttl time.Duration) {
	m.TTL = ttl
	m.ExpiresAt = m.CreatedAt.Add(ttl)
}

func (m *Message) IsExpired() bool {
	now := time.Now()
	return now.After(m.ExpiresAt)
}
