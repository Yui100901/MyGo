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
	c, _ := NewMQTTClient(
		MQTTConfiguration{
			ID:       "test",
			URL:      "tcp://8.147.130.215:19683",
			Username: "",
			Password: "",
		})
	c.Subscribe("Vehicle-7_stream_mic", 0, func(client mqtt.Client, message mqtt.Message) {

	})
	for {
		time.Sleep(1 * time.Second)
		c.Publish(NewMQTTPublishRequest("test", 0, false, []byte("test message")))
	}
}
