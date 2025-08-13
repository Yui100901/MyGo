package mq

import (
	"github.com/Yui100901/MyGo/concurrency"
)

//
// @Author yfy2001
// @Date 2025/7/31 13 50
//

// SubscriptionManager 订阅关系管理器
type SubscriptionManager struct {
	topicSubscribers *concurrency.SafeMap[Topic, map[string]struct{}] // topic -> subscribers
	subscriberTopics *concurrency.SafeMap[string, map[Topic]struct{}] // subscriber -> topics
}

// NewSubscriptionManager 创建订阅管理器
func NewSubscriptionManager() *SubscriptionManager {
	return &SubscriptionManager{
		topicSubscribers: concurrency.NewSafeMap[Topic, map[string]struct{}](32),
		subscriberTopics: concurrency.NewSafeMap[string, map[Topic]struct{}](32),
	}
}

// AddSubscription 添加订阅
func (r *SubscriptionManager) AddSubscription(subscriberID string, topic Topic) {
	r.topicSubscribers.Update(topic, func(old map[string]struct{}) (map[string]struct{}, bool) {
		if old == nil {
			old = make(map[string]struct{})
		}
		old[subscriberID] = struct{}{}
		return old, true
	})

	r.subscriberTopics.Update(subscriberID, func(old map[Topic]struct{}) (map[Topic]struct{}, bool) {
		if old == nil {
			old = make(map[Topic]struct{})
		}
		old[topic] = struct{}{}
		return old, true
	})

}

// RemoveSubscription 移除订阅
func (r *SubscriptionManager) RemoveSubscription(subscriberID string, topic Topic) {
	r.topicSubscribers.Update(topic, func(old map[string]struct{}) (map[string]struct{}, bool) {
		if old == nil {
			return nil, false
		}
		delete(old, subscriberID)
		if len(old) == 0 {
			return nil, false
		}
		return old, true
	})

	r.subscriberTopics.Update(subscriberID, func(old map[Topic]struct{}) (map[Topic]struct{}, bool) {
		if old == nil {
			return nil, false
		}
		delete(old, topic)
		if len(old) == 0 {
			return nil, false
		}
		return old, true
	})

}

// GetSubscribersCount 获取主题订阅者数量
func (r *SubscriptionManager) GetSubscribersCount(topic Topic) int {
	subscribers, exists := r.topicSubscribers.Get(topic)
	if !exists {
		return 0
	}

	return len(subscribers)
}

// GetTopicSubscribers 获取主题订阅者
func (r *SubscriptionManager) GetTopicSubscribers(topic Topic) []string {
	subscribers, exists := r.topicSubscribers.Get(topic)
	if !exists {
		return nil
	}

	result := make([]string, 0, len(subscribers))
	for subscriberID := range subscribers {
		result = append(result, subscriberID)
	}
	return result
}

// GetSubscriberTopicsCount 获取客户端订阅的主题的数量
func (r *SubscriptionManager) GetSubscriberTopicsCount(subscriberID string) int {
	topics, exists := r.subscriberTopics.Get(subscriberID)
	if !exists {
		return 0
	}
	return len(topics)
}

// GetSubscriberTopics 获取客户端订阅的主题
func (r *SubscriptionManager) GetSubscriberTopics(subscriberID string) []Topic {
	topics, exists := r.subscriberTopics.Get(subscriberID)
	if !exists {
		return nil
	}

	result := make([]Topic, 0, len(topics))
	for topic := range topics {
		result = append(result, topic)
	}
	return result
}

// RemoveSubscriber 移除客户端所有订阅
func (r *SubscriptionManager) RemoveSubscriber(subscriberID string) {
	topics := r.GetSubscriberTopics(subscriberID)
	for _, topic := range topics {
		r.RemoveSubscription(subscriberID, topic)
	}
}

// GetAllTopics 获取所有主题
func (r *SubscriptionManager) GetAllTopics() []Topic {
	return r.topicSubscribers.Keys()
}

// GetAllSubscribers 获取所有客户端
func (r *SubscriptionManager) GetAllSubscribers() []string {
	return r.subscriberTopics.Keys()
}

// GetAllTopicsCount 返回所有的主题数量
func (r *SubscriptionManager) GetAllTopicsCount() int {
	return r.topicSubscribers.Length()
}

// GetAllSubscribersCount 返回所有的客户端数量
func (r *SubscriptionManager) GetAllSubscribersCount() int {
	return r.subscriberTopics.Length()
}
