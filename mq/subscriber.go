package mq

import (
	"context"
	"errors"
	"github.com/Yui100901/MyGo/concurrency"
	"log"
	"os"
	"strings"
	"time"
)

//
// @Author yfy2001
// @Date 2025/8/12 09 54
//

// MessageHandler 消息处理函数类型
type MessageHandler func(ctx context.Context, msg *Message) error

// TopicSubscription 主题订阅信息
type TopicSubscription struct {
	Topic        Topic          // 主题名称
	Handler      MessageHandler // 该主题的消息处理函数
	SubscribedAt time.Time      // 订阅时间
}

// Subscriber 消息订阅者
type Subscriber struct {
	id            string                                          // 客户端ID
	subscriptions *concurrency.SafeMap[Topic, *TopicSubscription] // 主题订阅映射: topic -> subscription

	logger *log.Logger
}

// NewSubscriber 创建基于函数的客户端
func NewSubscriber(id string) *Subscriber {

	subscriber := &Subscriber{
		id:            id,
		subscriptions: concurrency.NewSafeMap[Topic, *TopicSubscription](32),
		logger:        log.New(os.Stdout, "[MQ-Subscriber ]", log.LstdFlags|log.Lshortfile),
	}

	log.Printf("Subscriber %s", subscriber.id)
	return subscriber
}

func (s *Subscriber) ID() string {
	return s.id
}

func (s *Subscriber) HandleMessage(message *Message) {
	sub, exists := s.subscriptions.Get(message.Topic)
	if !exists {
		return
	}
	if sub.Handler != nil {
		go s.processMessage(message, sub.Handler)
	}
}

func (s *Subscriber) processMessage(message *Message, handler MessageHandler) {
	defer func() {
		if err := recover(); err != nil {
			s.logger.Printf("message id :%s topic:%s,handler got err:%s", message.ID, message.Topic, err)
		}
	}()
	err := handler(context.Background(), message)
	if err != nil {
		s.logger.Printf("message id :%s topic:%s,handler got err:%s", message.ID, message.Topic, err)
		return
	}
}

func (s *Subscriber) Subscribe(topicMap map[Topic]MessageHandler) error {
	if len(topicMap) == 0 {
		return nil
	}

	for topic, handler := range topicMap {
		err := s.TopicValidate(topic)
		if err != nil {
			return err
		}

		// 创建订阅
		subscription := &TopicSubscription{
			Topic:        topic,
			Handler:      handler,
			SubscribedAt: time.Now(),
		}

		s.subscriptions.Set(topic, subscription)
	}

	return nil
}

func (s *Subscriber) Unsubscribe(topics []Topic) {
	for _, topic := range topics {
		s.subscriptions.Delete(topic)
	}
	s.logger.Printf("Unscribe from %d topics", len(topics))
}

func (s *Subscriber) TopicValidate(topic Topic) error {
	topicParts := strings.Split(string(topic), "/")
	//订阅时topic规则
	//点对点消息只能订阅自己的
	if topic.IsP2P() {
		if topicParts[1] != s.id {
			return errors.New("p2p topic id does not match")
		}
	}
	return nil
}
