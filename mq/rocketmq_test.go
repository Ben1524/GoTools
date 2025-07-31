package mq

import (
	"context"
	"fmt"
	"github.com/apache/rocketmq-client-go/v2"
	_ "github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/consumer"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/apache/rocketmq-client-go/v2/producer"
	"math"
	"os"
	"testing"
	"time"
)

var RocketMQProducer rocketmq.Producer
var (
	RocketMQAddress = []string{"127.0.0.1:9876"}
)

func TestMain(m *testing.M) {
	var err error
	RocketMQProducer, err = rocketmq.NewProducer(
		producer.WithNameServer(RocketMQAddress),
		producer.WithRetry(2),
	)
	if err != nil {
		fmt.Printf("Failed to create producer: %v\n", err)
		os.Exit(1)
	}

	err = RocketMQProducer.Start()
	if err != nil {
		fmt.Printf("Failed to start producer: %v\n", err)
		os.Exit(1)
	}

	os.Exit(m.Run())

}

func TestRocketMQProducer_SendSync(t *testing.T) {
	err := RocketMQProducer.SendOneWay(context.Background(), &primitive.Message{
		Topic: "msg",
		Body:  []byte("test"),
	})
	if err != nil {
		return
	}
	t.Log("Message sent successfully")
}

func TestRocketMQProducer_SendAsync(t *testing.T) {
	err := RocketMQProducer.SendAsync(context.Background(), func(ctx context.Context, result *primitive.SendResult, err error) {
		if err != nil {
			t.Errorf("Failed to send message: %v", err)
			return
		}
		t.Logf("Message sent successfully: %s", result.String())
	}, &primitive.Message{
		Topic: "msg",
		Body:  []byte("test async"),
	})
	if err != nil {
		t.Errorf("Failed to send async message: %v", err)
		return
	}
	t.Log("Async message sent successfully")
}

func TestMqSubscribe(t *testing.T) {
	//注册消费者
	mqPushConsumer, err := rocketmq.NewPushConsumer( // push模式，当有
		consumer.WithGroupName("test_consumer"),
		consumer.WithNameServer(RocketMQAddress),
	)
	if err != nil {
		panic(err)
	}
	msgHandler := func(ctx context.Context,
		msgs ...*primitive.MessageExt) (consumer.ConsumeResult, error) {
		for i := range msgs {
			fmt.Printf("接受到消息: %v /n", msgs[i])
		}
		return consumer.ConsumeSuccess, nil

	}
	err = mqPushConsumer.Subscribe("msg", consumer.MessageSelector{}, msgHandler)
	if err != nil {
		t.Error(err)
	}
	fmt.Println("启动消费者")
	mqPushConsumer.Start()
	select {
	case <-time.After(10 * time.Second):
	}
}

func TestMqSendAndReceive(t *testing.T) {
	// 发送消息
	go func() {
		for i := 0; i < math.MaxInt64; i++ {
			time.Sleep(100 * time.Millisecond)
			msg := &primitive.Message{
				Topic: "msg",
				Body:  []byte(fmt.Sprintf("test message %d", i)),
			}
			if i%2 == 0 {
				msg.WithTag("tag1")
			} else {
				msg.WithTag("tag2")
			}

			msg.WithKeys([]string{fmt.Sprintf("key%d", i)})

			err := RocketMQProducer.SendAsync(context.Background(), func(ctx context.Context, result *primitive.SendResult, err error) {
				if err != nil {
					t.Errorf("Failed to send message: %v", err)
					return
				}
				t.Logf("Message sent successfully: %s", result.String())
			}, msg)
			if err != nil {
				t.Fatal(fmt.Errorf("Failed to send async message: %v", err))
			}
		}
	}()

	go func() {
		//注册消费者
		mqPushConsumer, err := rocketmq.NewPushConsumer(
			consumer.WithGroupName("test_consumer1"),
			consumer.WithNameServer(RocketMQAddress),
		)
		if err != nil {
			panic(err)
		}
		msgHandler := func(ctx context.Context,
			msgs ...*primitive.MessageExt) (consumer.ConsumeResult, error) {
			for i := range msgs {
				fmt.Printf("接受到消息: %v /n", msgs[i])
			}
			return consumer.ConsumeSuccess, nil
		}
		err = mqPushConsumer.Subscribe("msg", consumer.MessageSelector{
			Type:       consumer.TAG,
			Expression: "tag1",
		}, msgHandler)
		if err != nil {
			t.Error(err)
		}
		fmt.Println("启动消费者")
		mqPushConsumer.Start()

		select {}
	}()

	go func() {
		//注册消费者
		mqPushConsumer, err := rocketmq.NewPushConsumer(
			consumer.WithGroupName("test_consumer2"),
			consumer.WithNameServer(RocketMQAddress),
		)
		if err != nil {
			panic(err)
		}
		msgHandler := func(ctx context.Context,
			msgs ...*primitive.MessageExt) (consumer.ConsumeResult, error) {
			for i := range msgs {
				fmt.Printf("接受到消息: %v /n", msgs[i])
			}
			return consumer.ConsumeSuccess, nil
		}
		err = mqPushConsumer.Subscribe("msg", consumer.MessageSelector{
			Type:       consumer.SQL92,
			Expression: "TAG in ('tag1','tag2')",
		}, msgHandler)
		if err != nil {
			t.Error(err)
		}
		fmt.Println("启动消费者")
		mqPushConsumer.Start()

		select {}
	}()

	// 阻止主函数退出
	select {}

}
