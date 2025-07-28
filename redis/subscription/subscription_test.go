package subscription

import (
	"context"
	"github.com/redis/go-redis/v9"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestPubSub(t *testing.T) {
	// Create a new Redis client
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	ctx := context.Background()
	cancelCtx, cancel := context.WithCancel(ctx)
	defer func() {
		cancel()
		if err := rdb.Close(); err != nil {
			t.Errorf("Failed to close Redis client: %v", err)
		}
	}()
	rec := rdb.Subscribe(cancelCtx, "test_channel")

	received := make(chan struct{})

	go func() {
		for msg := range rec.Channel() {
			t.Logf("Received message: %s from channel: %s", msg.Payload, msg.Channel)
			if msg.Payload == "quit" {
				break
			}
		}
		t.Log("Subscription channel closed")
		received <- struct{}{}
	}()
	// 发布消息
	rdb.Publish(ctx, "test-channel", "hello")

	time.Sleep(100 * time.Millisecond) // 等待消息被处理

	// 发布结束消息
	rdb.Publish(ctx, "test_channel", "quit")

	// 等待订阅者确认
	select {
	case <-received:
		t.Log("Message received")
	case <-time.After(2 * time.Second):
		t.Fatal("Message not received")
	}
}

func TestPubSub_MultiSubscribers(t *testing.T) {
	client := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	var wg sync.WaitGroup
	var count int32

	// 启动 5 个订阅者
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			pubsub := client.Subscribe(context.Background(), "test-channel")
			defer pubsub.Close()
			for msg := range pubsub.Channel() {
				if msg.Payload == "broadcast" {
					atomic.AddInt32(&count, 1)
				}
			}
		}()
	}

	// 发布消息
	client.Publish(context.Background(), "test-channel", "broadcast")
	time.Sleep(100 * time.Millisecond) // 等待消息分发

	wg.Wait()
	if count != 5 {
		t.Fatalf("Expected 5 subscribers to receive message, got %d", count)
	}
}

func TestPubSub_PatternSubscription(t *testing.T) {
	client := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	pubsub := client.PSubscribe(context.Background(), "sports.*")
	defer pubsub.Close()

	received := make(chan bool)
	go func() {
		for msg := range pubsub.Channel() {
			t.Logf("Received message: %s from pattern: %s", msg.Payload, msg.Channel)
			if msg.Payload == "quit" {
				break
			}
		}
		received <- true
	}()

	client.Publish(context.Background(), "sports.football", "goal1")
	client.Publish(context.Background(), "sports.basketball", "goal2")
	client.Publish(context.Background(), "sports.tennis", "goal3")
	client.Publish(context.Background(), "sports.quit", "quit")
	select {
	case <-received:
		t.Log("Pattern match succeeded")
	case <-time.After(2 * time.Second):
		t.Fatal("Pattern match failed")
	}
}
