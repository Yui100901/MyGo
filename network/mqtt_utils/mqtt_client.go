package mqtt_utils

import (
	"fmt"
	"github.com/Yui100901/MyGo/log_utils"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"maps"
)

//
// @Author yfy2001
// @Date 2024/8/15 16 00
//

type MQTTHandler interface {
	OnConnectHandler(client mqtt.Client)
	ConnectionLostHandler(client mqtt.Client, err error)
}

type MQTTClient struct {
	ClientId        string
	topicMap        map[string]byte // 订阅表
	client          mqtt.Client     // 客户端连接
	defaultCallback mqtt.MessageHandler
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

func NewMQTTClient(config MQTTConfiguration, topicMap map[string]byte, defaultCallback mqtt.MessageHandler) *MQTTClient {
	id := config.ID
	c := &MQTTClient{
		ClientId:        id,
		topicMap:        topicMap,
		defaultCallback: defaultCallback,
	}
	opts := mqtt.NewClientOptions()
	opts.SetClientID(id).
		SetAutoReconnect(true).
		AddBroker(config.URL).
		SetUsername(config.Username).
		SetPassword(config.Password).
		SetOnConnectHandler(c.OnConnectHandler).
		SetConnectionLostHandler(c.ConnectionLostHandler)
	c.client = mqtt.NewClient(opts)
	if conn := c.client.Connect(); conn.Wait() && conn.Error() != nil {
		log_utils.Error.Println(conn.Error())
		return nil
	}
	return c
}

func (c *MQTTClient) SubscribeDefault() {
	if len(c.topicMap) == 0 {
		return
	}
	// 重新订阅之前的所有主题
	if token := c.client.SubscribeMultiple(c.topicMap, c.defaultCallback); token.Wait() && token.Error() != nil {
		log_utils.Error.Println(c.ClientId, "Error resubscribing to topics:", token.Error())
	} else {
		log_utils.Info.Println(c.ClientId, "Resubscribed to topics:", c.topicMap)
	}
}

func (c *MQTTClient) Subscribe(topicMap map[string]byte, callback func(client mqtt.Client, msg mqtt.Message)) {
	maps.Copy(c.topicMap, topicMap)
	if token := c.client.SubscribeMultiple(topicMap, callback); token.Wait() && token.Error() != nil {
		log_utils.Error.Println(c.ClientId, "Error subscribing to topics:", token.Error())
	} else {
		log_utils.Info.Println(c.ClientId, "Subscribed to topics:", topicMap)
	}
}

func (c *MQTTClient) Unsubscribe(topicList []string) {
	for _, topic := range topicList {
		delete(c.topicMap, topic)
		if token := c.client.Unsubscribe(topic); token.Wait() && token.Error() != nil {
			log_utils.Error.Println(c.ClientId, "Error unsubscribing from topic:", topic, token.Error())
		} else {
			log_utils.Info.Println(c.ClientId, "Unsubscribed from topic:", topic)
		}
	}
}

func (c *MQTTClient) Publish(r *MQTTPublishRequest) ([]byte, error) {
	if !c.client.IsConnected() {
		if token := c.client.Connect(); token.Wait() && token.Error() != nil {
			log_utils.Error.Println("连接失败: %v", token.Error())
			return nil, token.Error()
		}
	}
	res := c.client.Publish(r.Topic, r.Qos, r.Retained, r.Payload)
	res.Wait()
	log_utils.Info.Println(c.ClientId, "Published a message", "Topic", r.Topic, r.Payload)
	return []byte(fmt.Sprintf(c.ClientId, "Published a message", "Topic", r.Topic, r.Payload)), nil
}

func (c *MQTTClient) OnConnectHandler(client mqtt.Client) {
	// 重新订阅之前的所有主题
	c.SubscribeDefault()
	log_utils.Info.Println(c.ClientId, "Connected!")
}

func (c *MQTTClient) ConnectionLostHandler(client mqtt.Client, err error) {
	log_utils.Error.Println(c.ClientId, "Connection Lost!\nError:", err.Error())
}
