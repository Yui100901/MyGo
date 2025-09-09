package mq

import (
	"context"
	"errors"
	"fmt"
	"github.com/Yui100901/MyGo/concurrency"
	"log"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

//
// @Author yfy2001
// @Date 2025/7/25 15 18
//

// BrokerConfig 消息代理配置
type BrokerConfig struct {
	MaxConcurrency  int           // 最大并发处理消息数量
	CleanupInterval time.Duration // 清理过期消息的间隔时间
	QueueSize       int           // 消息队列缓冲区大小
}

// DefaultBrokerConfig 返回默认的代理配置
func DefaultBrokerConfig() *BrokerConfig {
	return &BrokerConfig{
		MaxConcurrency:  100,
		CleanupInterval: 1 * time.Minute,
		QueueSize:       1000,
	}
}

// MessageBroker 消息代理，负责消息的路由、分发和管理
type MessageBroker struct {
	config              *BrokerConfig                             // 代理配置
	subscriptionManager *SubscriptionManager                      // 订阅关系管理
	subscribers         *concurrency.SafeMap[string, *Subscriber] // 订阅者注册表: subscriberID -> subscriberInterface
	messages            *concurrency.SafeMap[string, *Message]    // 消息存储: messageID -> Message
	pendingMessages     chan *Message                             // 待处理消息队列
	deliveryTimers      *concurrency.SafeMap[string, *time.Timer] // 消息投递超时定时器

	ctx    context.Context    // 上下文，用于控制组件生命周期
	cancel context.CancelFunc // 取消函数
	wg     sync.WaitGroup     // 等待组，用于优雅关闭

	msgCounter int64 // 消息计数器，用于生成唯一ID
	running    int32 // 运行状态标志（原子操作）
	logger     *log.Logger
}

// NewMessageBroker 创建新的消息代理实例
func NewMessageBroker(config *BrokerConfig) *MessageBroker {
	if config == nil {
		config = DefaultBrokerConfig()
	}

	ctx, cancel := context.WithCancel(context.Background())

	broker := &MessageBroker{
		config:              config,
		subscriptionManager: NewSubscriptionManager(),
		subscribers:         concurrency.NewSafeMap[string, *Subscriber](32),
		messages:            concurrency.NewSafeMap[string, *Message](32),
		pendingMessages:     make(chan *Message, config.QueueSize),
		deliveryTimers:      concurrency.NewSafeMap[string, *time.Timer](32),
		ctx:                 ctx,
		cancel:              cancel,
		logger:              log.New(os.Stdout, "[MQ-Broker] ", log.LstdFlags),
	}

	log.Printf("Message broker created with config: MaxConcurrency=%d, QueueSize=%d",
		config.MaxConcurrency, config.QueueSize)
	return broker
}

// Start 启动消息代理
func (b *MessageBroker) Start() error {
	if !atomic.CompareAndSwapInt32(&b.running, 0, 1) {
		return errors.New("broker is already running")
	}

	// 启动消息分发协程池
	for i := 0; i < b.config.MaxConcurrency; i++ {
		b.wg.Add(1)
		go b.messageDistributor()
	}
	b.logger.Printf("Started message distributor total %d workers", b.config.MaxConcurrency)

	// 启动清理协程
	b.wg.Add(1)
	go b.cleaner()
	b.logger.Printf("Started cleaner worker")

	// 启动监控协程
	b.wg.Add(1)
	go b.monitor()
	b.logger.Printf("Started monitor worker")

	b.logger.Printf("Message broker started successfully")
	return nil
}

// Stop 停止消息代理
func (b *MessageBroker) Stop() error {
	if !atomic.CompareAndSwapInt32(&b.running, 1, 0) {
		return errors.New("broker is not running")
	}

	b.logger.Printf("Stopping message broker")

	// 发送停止信号
	b.cancel()

	// 关闭消息队列
	close(b.pendingMessages)
	b.logger.Printf("Closed pending messages queue")

	// 停止所有定时器
	timerCount := b.deliveryTimers.Length()
	b.deliveryTimers.ForEachAsync(func(msgID string, timer *time.Timer) {
		b.logger.Printf("Stopped delivery timer for message %s", msgID)
		timer.Stop()
	})
	b.deliveryTimers = concurrency.NewSafeMap[string, *time.Timer](32)
	b.logger.Printf("Stopped %d delivery timers", timerCount)

	// 等待所有协程结束
	b.logger.Printf("Waiting for all workers to stop")
	b.wg.Wait()

	b.logger.Printf("Message broker stopped successfully")
	return nil
}

// RegisterSubscriber 注册订阅者
func (b *MessageBroker) RegisterSubscriber(subscriber *Subscriber) error {
	subscriberID := subscriber.ID()
	if _, exists := b.subscribers.Get(subscriberID); exists {
		b.logger.Printf("Subscriber %s already exists, rejecting registration", subscriberID)
		return fmt.Errorf("subscriber %s already exists", subscriberID)
	}
	b.subscribers.Set(subscriberID, subscriber)
	b.logger.Printf("Subscriber %s registered", subscriberID)
	return nil
}

// UnregisterSubscriber 注销订阅者
func (b *MessageBroker) UnregisterSubscriber(subscriberID string) {
	_, exists := b.subscribers.Get(subscriberID)
	if !exists {
		b.logger.Printf("Subscriber %s not found for unregistration", subscriberID)
		return
	}
	b.subscriptionManager.RemoveSubscriber(subscriberID)
	b.subscribers.Delete(subscriberID)
	return
}

func (b *MessageBroker) Publish(msg *Message) error {
	if atomic.LoadInt32(&b.running) == 0 {
		return errors.New("broker is not running")
	}

	b.logger.Printf("Publishing message from sender %s to topic %s (payload size: %d bytes)",
		msg.SenderID, msg.Topic, len(msg.Payload))

	// 存储消息
	b.messages.Set(msg.ID, msg)

	// 发送消息到分发队列
	select {
	case b.pendingMessages <- msg:
		// 消息发送成功
		b.logger.Printf("Message %s queued for distribution", msg.ID)
		return nil
	case <-b.ctx.Done():
		b.logger.Printf("Broker shutting down, cannot publish message %s", msg.ID)
		return errors.New("broker is shutting down")
	}
}

func (b *MessageBroker) Subscribe(subscriberID string, topicMap map[Topic]MessageHandler) error {
	// 检查订阅者是否存在
	subscriber, exists := b.subscribers.Get(subscriberID)
	if !exists {
		b.logger.Printf("Subscriber %s not found for subscription to topic %v", subscriberID, topicMap)
		return fmt.Errorf("subscriber %s not found", subscriberID)
	}
	for topic := range topicMap {
		b.subscriptionManager.AddSubscription(subscriberID, topic)
	}

	return subscriber.Subscribe(topicMap)
}

// Unsubscribe 取消订阅
func (b *MessageBroker) Unsubscribe(subscriberID string, topics []Topic) {
	if len(topics) == 0 {
		return
	}
	subscriber, exists := b.subscribers.Get(subscriberID)
	if !exists {
		return
	}

	for _, topic := range topics {
		b.subscriptionManager.RemoveSubscription(subscriberID, topic)
		b.logger.Printf("Unsubscribed from topic %s", topic)
	}
	subscriber.Unsubscribe(topics)
}

func (b *MessageBroker) CancelDelayedMessage(msgID string) bool {
	timer, ok := b.deliveryTimers.Get(msgID)
	if !ok {
		b.logger.Printf("No delayed message found with ID %s", msgID)
		return false
	}

	stopped := timer.Stop()
	if stopped {
		b.logger.Printf("Successfully cancelled delayed message %s", msgID)
	} else {
		b.logger.Printf("Failed to cancel message %s (already triggered)", msgID)
	}
	b.deliveryTimers.Delete(msgID)
	return stopped
}

// messageDistributor 消息分发器协程
func (b *MessageBroker) messageDistributor() {
	defer b.wg.Done()
	b.logger.Printf("Message distributor worker started")

	for msg := range b.pendingMessages {
		select {
		case <-b.ctx.Done():
			return
		default:
			b.distributeMessage(msg)
		}
	}
}

// distributeMessage 分发单个消息给订阅者
func (b *MessageBroker) distributeMessage(msg *Message) {
	if msg.IsExpired() {
		b.messages.Delete(msg.ID)
		b.logger.Printf("Message %s expired before distribution", msg.ID)
		return
	}

	b.logger.Printf("Distributing message %s for topic %s", msg.ID, msg.Topic)
	if msg.Delay > 0 {
		b.logger.Printf("Message %s scheduled for delayed delivery in %s", msg.ID, msg.Delay)

		timer := time.AfterFunc(msg.DeliverAt.Sub(time.Now()), func() {
			b.deliveryTimers.Delete(msg.ID) // 清理定时器
			b.logger.Printf("Delayed delivery triggered for message %s", msg.ID)
			b.sendToSubscriber(msg)
		})

		b.deliveryTimers.Set(msg.ID, timer)
	} else {
		b.sendToSubscriber(msg)
	}

}

func (b *MessageBroker) sendToSubscriber(msg *Message) {
	// 确定目标订阅者
	subscriberIDs := b.subscriptionManager.GetTopicSubscribers(msg.Topic)
	//没有目标订阅者直接丢弃消息
	if len(subscriberIDs) == 0 {
		b.logger.Printf("No subscribers found for topic %s, message %s discarded", msg.Topic, msg.ID)
		b.messages.Delete(msg.ID)
		return
	}

	targetSubscribers := make([]*Subscriber, 0, len(subscriberIDs))
	for _, cid := range subscriberIDs {
		if subscriber, exists := b.subscribers.Get(cid); exists {
			targetSubscribers = append(targetSubscribers, subscriber)
		}
	}

	// 通知所有订阅的订阅者有新消息
	for _, subscriber := range targetSubscribers {
		// 创建消息副本
		msgCopy := *msg
		subscriber.HandleMessage(&msgCopy)
	}

	b.logger.Printf("Delivered to %d subscribers about message %s", len(targetSubscribers), msg.ID)
}

// GetMessage 获取消息详情
func (b *MessageBroker) GetMessage(messageID string) (*Message, error) {

	if msg, exists := b.messages.Get(messageID); exists {
		// 返回消息副本以避免外部修改
		msgCopy := *msg
		return &msgCopy, nil
	}

	return nil, fmt.Errorf("message %s not found", messageID)
}

// cleaner 清理过期消息的协程
func (b *MessageBroker) cleaner() {
	defer b.wg.Done()
	b.logger.Printf("Cleaner started with interval %v", b.config.CleanupInterval)

	ticker := time.NewTicker(b.config.CleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			b.cleanupExpiredMessages()
		case <-b.ctx.Done():
			b.logger.Printf("Cleaner stopping")
			return
		}
	}
}

// cleanupExpiredMessages 清理过期的已确认和已死亡消息
func (b *MessageBroker) cleanupExpiredMessages() {

	msgIdListToDelete := make([]string, 0)
	b.messages.ForEach(func(id string, msg *Message) bool {
		if msg.IsExpired() {
			msgIdListToDelete = append(msgIdListToDelete, id)
		}
		return true
	})

	for _, id := range msgIdListToDelete {
		b.messages.Delete(id)
	}
	cleanedCount := len(msgIdListToDelete)
	if cleanedCount > 0 {
		b.logger.Printf("Cleaned up %d expired messages", cleanedCount)
	}
}

// monitor 监控协程
func (b *MessageBroker) monitor() {
	defer b.wg.Done()
	b.logger.Printf("Monitor started")

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			stats := b.GetStats()
			b.logger.Printf("Stats - Subscribers: %v, Topics: %v, Messages: %v, Pending: %v, Timers: %v",
				stats["total_subscribers"], stats["total_topics"], stats["total_messages"],
				stats["pending_queue_size"], stats["delivery_timers"])
		case <-b.ctx.Done():
			b.logger.Printf("Monitor stopping")
			return
		}
	}
}

// GetStats 获取消息代理的统计信息
func (b *MessageBroker) GetStats() map[string]interface{} {

	stats := map[string]interface{}{
		"total_subscribers":  b.subscribers.Length(),
		"total_topic":        b.subscriptionManager.GetAllTopicsCount(),
		"total_messages":     b.messages.Length(),
		"pending_queue_size": len(b.pendingMessages),
		"delivery_timers":    b.deliveryTimers.Length(),
		"running":            atomic.LoadInt32(&b.running) == 1,
		"message_counter":    atomic.LoadInt64(&b.msgCounter),
	}

	// 按状态统计消息
	statusCount := make(map[string]int)
	b.messages.ForEach(func(id string, msg *Message) bool {
		statusCount[id] = statusCount[id] + 1
		return true
	})

	stats["message_status"] = statusCount

	// 统计每个主题的订阅者数量
	topicStats := make(map[Topic]int)
	topics := b.subscriptionManager.GetAllTopics()
	for _, topic := range topics {
		topicStats[topic] = b.subscriptionManager.GetSubscribersCount(topic)
	}
	stats["topic_subscribers"] = topicStats

	return stats
}

func (b *MessageBroker) IsRunning() bool {
	return atomic.LoadInt32(&b.running) == 1
}

func (b *MessageBroker) GetPendingMessageCount() int {
	return len(b.pendingMessages)
}

func (b *MessageBroker) Shutdown() error {
	return b.Stop()
}
