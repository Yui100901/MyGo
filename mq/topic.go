package mq

import (
	"errors"
	"strings"
)

//
// @Author yfy2001
// @Date 2025/7/28 11 45
//

// 保留topic定义
const (
	TopicBroadcastPrefix  Topic = "broadcast/"
	TopicP2PPrefix        Topic = "p2p/"
	TopicDeadLetterPrefix Topic = "deadLetter/"
)

type Topic string

func (t Topic) IsBroadcast() bool {
	return strings.HasPrefix(string(t), string(TopicBroadcastPrefix))
}

func (t Topic) IsP2P() bool {
	return strings.HasPrefix(string(t), string(TopicP2PPrefix))
}

func (t Topic) IsDeadLetter() bool {
	return strings.HasPrefix(string(t), string(TopicDeadLetterPrefix))
}

func (t Topic) Validate() error {
	if t == "" {
		return errors.New("topic can not be empty")
	}
	return nil
}
