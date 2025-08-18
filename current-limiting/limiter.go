package current_limiting

import (
	"context"
	"sync"
	"time"
)

type Limiter struct {
	mu        sync.Mutex // 互斥锁，确保并发安全
	rate      int64      // 每秒允许的请求数
	capacity  int64      // 令牌桶的容量
	tokens    float64    // 当前令牌数
	lastCheck time.Time  // 上次检查时间
}

func NewLimiter(rate int64, capacity int64) *Limiter {
	return &Limiter{
		rate:      rate,
		capacity:  capacity,
		tokens:    0, // 初始化令牌数为空，避免请求太多击穿后端服务
		lastCheck: time.Now(),
	}
}

func (l *Limiter) durationFromTokens(tokens float64) time.Duration {
	if tokens <= 0 {
		return 0
	}
	return time.Duration(float64(time.Second) * (tokens / float64(l.rate)))
}

func (l *Limiter) updateTokens() {
	now := time.Now()
	elapsed := now.Sub(l.lastCheck).Seconds() // 计算自上次检查以来经过的秒数
	l.lastCheck = now

	// 增加令牌数
	l.tokens += float64(elapsed) * float64(l.rate)
	if l.tokens > float64(l.capacity) {
		l.tokens = float64(l.capacity) // 限制令牌数不超过容量
	}
}

func (l *Limiter) SetRate(rate int64) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.updateTokens() // 更新令牌数
	l.rate = rate
}

func (l *Limiter) SetCapacity(capacity int64) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.updateTokens() // 更新令牌数
	l.capacity = capacity
	if l.tokens > float64(capacity) {
		l.tokens = float64(capacity) // 确保令牌数不超过新的容量
	}
}

func (l *Limiter) Allow() bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.updateTokens() // 更新令牌数
	if l.tokens >= 1 {
		l.tokens -= 1 // 消耗一个令牌
		return true
	}
	return false // 没有足够的令牌
}

func (l *Limiter) Wait() {
	for {
		l.mu.Lock()
		l.updateTokens()
		if l.tokens >= 1 {
			l.tokens -= 1
			l.mu.Unlock()
			return
		}
		waitTime := l.durationFromTokens(1 - l.tokens)
		l.mu.Unlock()
		time.Sleep(waitTime)
	}
}

// 归还令牌
func (l *Limiter) Return(tokens float64) {
	if tokens < 0 {
		return // 不允许归还负令牌
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	l.tokens += tokens
	if l.tokens > float64(l.capacity) {
		l.tokens = float64(l.capacity) // 确保令牌数不超过容量
	}
}

// 等待N个令牌
func (l *Limiter) WaitN(ctx context.Context, n int64) error {
	if n <= 0 {
		return nil // 不需要等待
	}

	l.mu.Lock()
	l.updateTokens() // 更新令牌数

	// 检查是否有足够的令牌
	if l.tokens >= float64(n) {
		l.tokens -= float64(n)
		l.mu.Unlock()
		return nil
	}

	// 计算需要等待的时间
	needTokens := float64(n) - l.tokens
	waitDuration := l.durationFromTokens(needTokens)
	l.tokens -= float64(n)   // 先扣除令牌，防止其他请求进来时重复计算
	l.lastCheck = time.Now() // 记录当前时间，避免重复计算
	l.mu.Unlock()

	// 等待指定时间或上下文取消
	for {
		select {
		case <-time.After(waitDuration): // 不考虑过分严苛的限流
			return nil // 成功获取所有令牌
		case <-ctx.Done():
			// 上下文取消，归还令牌
			l.mu.Lock()
			l.tokens += float64(n)
			l.updateTokens() // 更新令牌数
			l.mu.Unlock()
			return ctx.Err()
		}
	}
}
