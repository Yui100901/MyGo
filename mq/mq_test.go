package mq

import (
	"context"
	"testing"
	"time"
)

//
// @Author yfy2001
// @Date 2025/8/8 14 57
//

func TestMQ(t *testing.T) {
	b := NewMessageBroker(nil)
	b.Start()
	c1 := NewSubscriber("c1")
	c2 := NewSubscriber("c2")
	b.RegisterSubscriber(c1)
	b.RegisterSubscriber(c2)
	b.Subscribe(c1.id, map[Topic]MessageHandler{
		"test": func(ctx context.Context, msg *Message) error {
			t.Log("c1 receive", msg)
			return nil
		},
	})
	b.Publish(NewMessage("test", []byte("hello world")))
	time.Sleep(20 * time.Second)
	b.Stop()

}
