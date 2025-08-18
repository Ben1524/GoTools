package current_limiting

import (
	"context"
	"testing"
	"time"
)

// 测试限流器的创建
func TestNewLimiter(t *testing.T) {
	rate := int64(10)
	capacity := int64(20)
	limiter := NewLimiter(rate, capacity)

	if limiter.rate != rate {
		t.Errorf("期望速率为 %d，实际为 %d", rate, limiter.rate)
	}

	if limiter.capacity != capacity {
		t.Errorf("期望容量为 %d，实际为 %d", capacity, limiter.capacity)
	}

	if limiter.tokens != 0 {
		t.Errorf("新创建的限流器初始令牌数应为 0，实际为 %f", limiter.tokens)
	}
}

// 测试Allow方法
func TestAllow(t *testing.T) {
	// 每秒10个令牌，容量10
	limiter := NewLimiter(10, 10)

	// 初始令牌为0，应该不允许
	if limiter.Allow() {
		t.Error("初始状态下不应该允许请求")
	}

	// 等待足够长时间让令牌充满
	time.Sleep(time.Second * 2)

	// 应该允许10个请求
	count := 0
	for i := 0; i < 20; i++ {
		if limiter.Allow() {
			count++
		}
	}

	if count != 10 {
		t.Errorf("期望允许10个请求，实际允许了 %d 个", count)
	}

	// 再等待0.5秒，应该补充5个令牌
	time.Sleep(time.Millisecond * 500)
	count = 0
	for i := 0; i < 10; i++ {
		if limiter.Allow() {
			count++
		}
	}

	if count != 5 {
		t.Errorf("期望允许5个请求，实际允许了 %d 个", count)
	}
}

// 测试Wait方法
func TestWait(t *testing.T) {
	// 每秒1个令牌，容量1
	limiter := NewLimiter(1, 1)

	// 先获取一个令牌，令牌桶为空
	if limiter.Allow() {
		t.Error("第一次请求不应该被允许")
	}

	// 测试等待时间
	start := time.Now()
	limiter.Wait()
	duration := time.Since(start)

	// 应该等待大约1秒
	if duration < time.Millisecond*900 || duration > time.Millisecond*1100 {
		t.Errorf("等待时间不符合预期，实际等待了 %v", duration)
	}
}

// 测试Return方法
func TestReturn(t *testing.T) {
	limiter := NewLimiter(10, 10)

	// 先获取5个令牌
	for i := 0; i < 5; i++ {
		limiter.Allow()
	}

	// 等待一段时间让令牌补充一点
	time.Sleep(time.Millisecond * 100)

	// 记录当前令牌数
	currentTokens := limiter.tokens

	// 归还2个令牌
	limiter.Return(2)

	// 检查是否正确归还
	if limiter.tokens != currentTokens+2 {
		t.Errorf("归还令牌不正确，期望 %f，实际 %f", currentTokens+2, limiter.tokens)
	}

	// 测试归还负数令牌（应该被忽略）
	limiter.Return(-1)
	if limiter.tokens != currentTokens+2 {
		t.Error("不应该允许归还负数令牌")
	}

	// 测试归还令牌超过容量
	limiter.Return(100)
	if limiter.tokens != float64(limiter.capacity) {
		t.Errorf("归还令牌不应超过容量，期望 %d，实际 %f", limiter.capacity, limiter.tokens)
	}
}

// 测试SetRate方法
func TestSetRate(t *testing.T) {
	limiter := NewLimiter(10, 20)
	newRate := int64(20)

	limiter.SetRate(newRate)
	if limiter.rate != newRate {
		t.Errorf("设置速率失败，期望 %d，实际 %d", newRate, limiter.rate)
	}
}

// 测试SetCapacity方法
func TestSetCapacity(t *testing.T) {
	limiter := NewLimiter(10, 20)

	// 先充满令牌
	time.Sleep(time.Second * 3)

	// 缩小容量
	newCapacity := int64(10)
	limiter.SetCapacity(newCapacity)

	if limiter.capacity != newCapacity {
		t.Errorf("设置容量失败，期望 %d，实际 %d", newCapacity, limiter.capacity)
	}

	if limiter.tokens > float64(newCapacity) {
		t.Errorf("令牌数不应超过新容量，期望 <=%d，实际 %f", newCapacity, limiter.tokens)
	}
}

// 测试WaitN方法
func TestWaitN(t *testing.T) {
	// 每秒2个令牌，容量5
	limiter := NewLimiter(2, 5)

	// 测试正常情况
	err := limiter.WaitN(context.Background(), 3)
	if err != nil {
		t.Errorf("WaitN 不应返回错误，实际错误: %v", err)
	}

	// 测试需要等待的情况
	start := time.Now()
	err = limiter.WaitN(context.Background(), 6)
	duration := time.Since(start)
	if err != nil {
		t.Errorf("WaitN 不应返回错误，实际错误: %v", err)
	}

	if duration < time.Second*2 || duration > time.Second*4 {
		t.Errorf("等待时间不符合预期，实际等待了 %v", duration)
	}

	// 测试上下文取消
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(time.Millisecond * 100)
		cancel()
	}()

	start = time.Now()
	err = limiter.WaitN(ctx, 10)
	if err == nil {
		t.Error("WaitN 应该返回上下文取消错误")
	}

	// 确保没有等待太长时间
	duration = time.Since(start)
	if duration > time.Millisecond*200 {
		t.Errorf("上下文取消后应立即返回，实际等待了 %v", duration)
	}

	// 测试n为0的情况
	err = limiter.WaitN(context.Background(), 0)
	if err != nil {
		t.Errorf("n为0时不应返回错误，实际错误: %v", err)
	}
}

// 测试并发情况下的限流器
func TestConcurrentAllow(t *testing.T) {
	// 每秒100个令牌，容量100
	limiter := NewLimiter(100, 100)

	// 等待令牌充满
	time.Sleep(time.Second * 2)

	count := 0
	ch := make(chan bool, 200)

	// 并发200个请求
	for i := 0; i < 200; i++ {
		go func() {
			ch <- limiter.Allow()
		}()
	}

	// 统计允许的请求数
	for i := 0; i < 200; i++ {
		if <-ch {
			count++
		}
	}

	// 应该允许100个请求
	if count != 100 {
		t.Errorf("并发情况下期望允许100个请求，实际允许了 %d 个", count)
	}
}
