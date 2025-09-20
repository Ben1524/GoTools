package websockettools

import (
	"context"
	"testing"
	"time"
	"github.com/gin-gonic/gin"
)

func TestStream(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	stream := NewWebSocketStream(ctx, 3*time.Second, 6*time.Second, []string{`^localhost$`, `^127\.0\.0\.1$`})

	route := gin.Default()
	route.GET("/ws", stream.GinHandler)

	if err := route.Run(":8080"); err != nil {
		t.Error(err)
	}
}

func TestClose(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	stream := NewWebSocketStream(ctx, 3*time.Second, 6*time.Second, []string{`^localhost$`, `^127\.0\.0\.1$`})

	route := gin.Default()
	route.GET("/ws", stream.GinHandler)

	// 5秒后关闭服务
	go func() {
		time.Sleep(15 * time.Second)
		stream.Close()
	}()

	if err := route.Run(":8080"); err != nil {
		t.Error(err)
	}
}