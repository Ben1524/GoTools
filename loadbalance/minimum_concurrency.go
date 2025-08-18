package loadbalance

import (
	"math"
	"math/rand"
	"sync/atomic"
)

// MinimumConcurrencySampler 用于最小并发采样
type MinimumConcurrencySampler struct {
	endpoints      []string // 存储所有的端点
	minConcurrency []int64  // 存储每个端点的最小并发数
}

func NewMinimumConcurrencySampler(endpoints []string, minConcurrency []int64) *MinimumConcurrencySampler {
	if len(endpoints) != len(minConcurrency) {
		panic("endpoints and minConcurrency must have the same length")
	}
	return &MinimumConcurrencySampler{
		endpoints:      endpoints,
		minConcurrency: minConcurrency,
	}
}

func (m *MinimumConcurrencySampler) Sample() string {
	if len(m.endpoints) == 0 {
		return ""
	}

	if len(m.endpoints) == 1 {
		return m.endpoints[0]
	}
	var minValue int64 = math.MaxInt64
	idx := -1
	begin := rand.Intn(len(m.endpoints)) // 随机开始位置
	for i := 0; i < len(m.endpoints); i++ {
		index := (begin + i) % len(m.endpoints)         // 环形遍历
		c := atomic.LoadInt64(&m.minConcurrency[index]) // 获取当前端点的最小并发数
		if c < minValue {
			minValue = m.minConcurrency[index]
			idx = index
		}
	}
	if idx == -1 {
		return ""
	}
	atomic.AddInt64(&m.minConcurrency[idx], 1) // 增加该端点的最小并发数
	return m.endpoints[idx]
}
