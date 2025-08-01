package mqtt_utils

import (
	"errors"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

const (
	defaultConnectTimeout = 5 * time.Second
	reconnectDelay        = 5 * time.Second
)

// Subscription 表示一个主题订阅的详细信息
type Subscription struct {
	Topic    string
	Qos      byte
	Callback mqtt.MessageHandler
}

type MQTTClient struct {
	ClientId      string
	subscriptions map[string]*Subscription // 订阅表，存储主题和订阅详情
	client        mqtt.Client              // 客户端连接
	stopChan      chan struct{}            // 停止信号
	logger        *log.Logger              // 日志记录器
	reconnectMu   sync.Mutex               // 重连操作锁
}

type MQTTPublishRequest struct {
	Topic    string
	Qos      byte
	Retained bool
	Payload  any
}

func NewMQTTPublishRequest(topic string, qos byte, retained bool, payload any) *MQTTPublishRequest {
	return &MQTTPublishRequest{
		Topic:    topic,
		Qos:      qos,
		Retained: retained,
		Payload:  payload,
	}
}

func NewMQTTClient(config MQTTConfiguration) (*MQTTClient, error) {
	c := &MQTTClient{
		subscriptions: make(map[string]*Subscription),
		stopChan:      make(chan struct{}),
		logger:        log.New(os.Stdout, "[MQTT] ", log.LstdFlags),
	}

	opts := mqtt.NewClientOptions()
	opts.SetClientID(config.ID).
		AddBroker(config.URL).
		SetUsername(config.Username).
		SetPassword(config.Password).
		SetAutoReconnect(false). // 使用自定义重连逻辑
		SetConnectTimeout(defaultConnectTimeout).
		SetOnConnectHandler(c.OnConnectHandler).
		SetConnectionLostHandler(c.ConnectionLostHandler)

	c.client = mqtt.NewClient(opts)

	if err := c.connect(); err != nil {
		return nil, err
	}

	go c.monitorConnection()

	return c, nil
}

// connect 内部连接方法
func (c *MQTTClient) connect() error {
	c.reconnectMu.Lock()
	defer c.reconnectMu.Unlock()

	token := c.client.Connect()
	if !token.WaitTimeout(defaultConnectTimeout) {
		return errors.New("connection timeout")
	}
	return token.Error()
}

// ensureConnection 确保连接可用
func (c *MQTTClient) ensureConnection() error {
	if c.client.IsConnected() {
		return nil
	}
	return c.connect()
}

// monitorConnection 连接监控和重连
func (c *MQTTClient) monitorConnection() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-c.stopChan:
			return
		case <-ticker.C:
			if !c.client.IsConnected() {
				c.logger.Println("Attempting to reconnect...")
				if err := c.connect(); err != nil {
					c.logger.Printf("Reconnect failed: %v. Retrying in %v", err, reconnectDelay)
					time.Sleep(reconnectDelay)
				}
			}
		}
	}
}

// IsConnected 检查连接状态
func (c *MQTTClient) IsConnected() bool {
	return c.client.IsConnected()
}

// ResubscribeAll 重新订阅所有已注册的主题
func (c *MQTTClient) ResubscribeAll() {
	if len(c.subscriptions) == 0 {
		return
	}

	// 准备批量订阅参数
	topics := make(map[string]byte, len(c.subscriptions))
	handlers := make(map[string]mqtt.MessageHandler, len(c.subscriptions))

	for topic, sub := range c.subscriptions {
		topics[topic] = sub.Qos
		handlers[topic] = sub.Callback
	}

	// 执行批量订阅
	if token := c.client.SubscribeMultiple(topics, func(client mqtt.Client, msg mqtt.Message) {
		if handler, ok := handlers[msg.Topic()]; ok {
			handler(client, msg)
		}
	}); token.Wait() && token.Error() != nil {
		c.logger.Printf("Resubscribe failed: %v", token.Error())
	} else {
		c.logger.Printf("Resubscribed to %d topics", len(topics))
	}
}

// Subscribe 订阅单个主题
func (c *MQTTClient) Subscribe(topic string, qos byte, callback mqtt.MessageHandler) {
	// 更新订阅表
	c.subscriptions[topic] = &Subscription{
		Topic:    topic,
		Qos:      qos,
		Callback: callback,
	}

	if c.IsConnected() {
		if token := c.client.Subscribe(topic, qos, callback); token.Wait() && token.Error() != nil {
			c.logger.Printf("Error subscribing to topic %s: %v", topic, token.Error())
		} else {
			c.logger.Printf("Subscribed to topic: %s (QoS: %d)", topic, qos)
		}
	} else {
		c.logger.Printf("Offline: Topic %s added to subscription list", topic)
	}
}

// SubscribeMultiple 批量订阅主题
func (c *MQTTClient) SubscribeMultiple(subscriptions map[string]byte, callback mqtt.MessageHandler) {
	if len(subscriptions) == 0 {
		return
	}

	// 更新订阅表
	for topic, qos := range subscriptions {
		c.subscriptions[topic] = &Subscription{
			Topic:    topic,
			Qos:      qos,
			Callback: callback,
		}
	}

	if c.IsConnected() {
		// 准备批量订阅
		handlers := make(map[string]mqtt.MessageHandler, len(subscriptions))
		for topic := range subscriptions {
			handlers[topic] = callback
		}

		if token := c.client.SubscribeMultiple(subscriptions, func(client mqtt.Client, msg mqtt.Message) {
			if handler, ok := handlers[msg.Topic()]; ok {
				handler(client, msg)
			}
		}); token.Wait() && token.Error() != nil {
			c.logger.Printf("Batch subscribe failed: %v", token.Error())
		} else {
			c.logger.Printf("Subscribed to %d topics", len(subscriptions))
		}
	} else {
		c.logger.Printf("Offline: Added %d topics to subscription list", len(subscriptions))
	}
}

// Unsubscribe 批量取消订阅
func (c *MQTTClient) Unsubscribe(topics ...string) {
	if len(topics) == 0 {
		return
	}

	// 更新订阅表
	for _, topic := range topics {
		delete(c.subscriptions, topic)
	}

	if c.IsConnected() {
		if token := c.client.Unsubscribe(topics...); token.Wait() && token.Error() != nil {
			c.logger.Printf("Batch unsubscribe failed: %v", token.Error())
		} else {
			c.logger.Printf("Unsubscribed from %d topics", len(topics))
		}
	} else {
		c.logger.Printf("Offline: Removed %d topics from subscription list", len(topics))
	}
}

// GetSubscription 获取主题的订阅详情
func (c *MQTTClient) GetSubscription(topic string) (qos byte, callback mqtt.MessageHandler, exists bool) {
	if sub, ok := c.subscriptions[topic]; ok {
		return sub.Qos, sub.Callback, true
	}
	return 0, nil, false
}

// Publish 发布消息（自动重连）
func (c *MQTTClient) Publish(r *MQTTPublishRequest) error {
	if err := c.ensureConnection(); err != nil {
		return fmt.Errorf("connection not available: %w", err)
	}

	token := c.client.Publish(r.Topic, r.Qos, r.Retained, r.Payload)
	if !token.WaitTimeout(defaultConnectTimeout) {
		return errors.New("publish timeout")
	}

	if err := token.Error(); err != nil {
		return err
	}

	c.logger.Printf("Published to topic %s (QoS: %d)", r.Topic, r.Qos)
	return nil
}

// Disconnect 断开连接
func (c *MQTTClient) Disconnect() {
	close(c.stopChan)
	c.client.Disconnect(250) // 等待250ms完成操作
	c.logger.Println("Disconnected from broker")
}

func (c *MQTTClient) OnConnectHandler(client mqtt.Client) {
	// 重新订阅所有已注册的主题
	c.ResubscribeAll()
	c.logger.Println("Connected and subscriptions restored")
}

func (c *MQTTClient) ConnectionLostHandler(client mqtt.Client, err error) {
	c.logger.Printf("Connection Lost! Error: %v", err)
}

// SetLogger 设置自定义日志器
func (c *MQTTClient) SetLogger(logger *log.Logger) {
	c.logger = logger
}
