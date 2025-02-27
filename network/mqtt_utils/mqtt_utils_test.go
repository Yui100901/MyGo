package mqtt_utils

import (
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"testing"
	"time"
)

//
// @Author yfy2001
// @Date 2025/2/27 15 02
//

func TestMQTTClient(t *testing.T) {
	c := NewMQTTClient(
		MQTTConfiguration{
			ID:       "test",
			URL:      "tcp://127.0.0.1:1883",
			Username: "",
			Password: "",
		}, map[string]byte{
			"test": 0,
		}, func(client mqtt.Client, message mqtt.Message) {
			t.Log(message.Topic(), string(message.Payload()))
		})
	c.SubscribeDefault()
	for range 10 {
		time.Sleep(1 * time.Second)
		c.Publish(NewMQTTPublishRequest("test", 0, false, "test message"))
	}
}
