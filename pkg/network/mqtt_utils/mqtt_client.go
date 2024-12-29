package mqtt_utils

import (
	"github.com/Yui100901/MyGo/pkg/log_utils"
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
	ClientId        string          // 客户端id
	topicMap        map[string]byte // 订阅表
	client          mqtt.Client     // 客户端连接
	defaultCallback mqtt.MessageHandler
}

func NewMQTTClient(clientID string, config MQTTConfiguration, topicMap map[string]byte, defaultCallback mqtt.MessageHandler) *MQTTClient {
	c := &MQTTClient{
		ClientId:        clientID,
		topicMap:        topicMap,
		defaultCallback: defaultCallback,
	}
	opts := mqtt.NewClientOptions()
	opts.SetClientID(clientID)
	// 设置断开连接时自动重新连接
	opts.SetAutoReconnect(true)
	opts.AddBroker(config.URL)
	opts.SetUsername(config.Username)
	opts.SetPassword(config.Password)
	opts.SetOnConnectHandler(c.OnConnectHandler)
	opts.SetConnectionLostHandler(c.ConnectionLostHandler)
	c.client = mqtt.NewClient(opts)
	if conn := c.client.Connect(); conn.Wait() && conn.Error() != nil {
		log_utils.Error.Println(conn.Error())
		return nil
	}
	return c
}

func (c *MQTTClient) SubscribeDefault() {
	if c.topicMap == nil || len(c.topicMap) == 0 {
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

func (c *MQTTClient) Publish(topic string, qos byte, payload any) {
	res := c.client.Publish(topic, qos, false, payload)
	res.Wait()
	log_utils.Info.Println(c.ClientId, "Published a message", "Topic", topic, payload)
}

func (c *MQTTClient) OnConnectHandler(client mqtt.Client) {
	// 重新订阅之前的所有主题
	c.SubscribeDefault()
	log_utils.Info.Println(c.ClientId, "Connected!")
}

func (c *MQTTClient) ConnectionLostHandler(client mqtt.Client, err error) {
	log_utils.Error.Println(c.ClientId, "Connection Lost!\nError:", err.Error())
}
