package algorithm

import (
	"fmt"
	"math/rand"
	"sync"
	"testing"
	"time"
)

func TestTimeoutCache(t *testing.T) {
	cache := NewTimeoutCache[string, int](5)

	deadline := time.Now().Add(5 * time.Second).UnixMilli()
	deadline2 := time.Now().UnixMilli() + 1000 // 1秒后过期

	// 测试添加和获取元素
	cache.Set("key1", 1, deadline)
	cache.Set("key2", 2, deadline2)
	if val, ok := cache.Get("key1"); !ok || val != 1 {
		t.Errorf("Expected 1, got %v", val)
	}

	// 测试更新元素
	cache.Set("key1", 2, deadline+1500)
	if val, ok := cache.Get("key1"); !ok || val != 2 {
		t.Errorf("Expected 2, got %v", val)
	}

	// 测试过期元素
	time.Sleep(2 * time.Second)
	if _, ok := cache.Get("key2"); ok {
		t.Error("Expected key2 to be expired")
	}
}

// 测试缓存容量限制和淘汰机制
func TestCacheCapacity(t *testing.T) {
	capacity := 3
	cache := NewTimeoutCache[string, int](capacity)

	// 添加超出容量的元素
	for i := 0; i < capacity+2; i++ {
		key := "key" + string(rune('0'+i))
		cache.Set(key, i, time.Now().Add(10*time.Second).UnixMilli())
		time.Sleep(100 * time.Millisecond) // 确保每个元素的添加时间不同
	}

	// 检查总数量是否正确
	if len(cache.cache) != capacity {
		t.Errorf("Expected %d elements, got %d", capacity, len(cache.cache))
	}

	// 最早添加的元素应该被淘汰
	if _, ok := cache.Get("key0"); ok {
		t.Error("Expected key0 to be evicted")
	}

	// 较晚添加的元素应该保留
	if _, ok := cache.Get("key2"); !ok {
		t.Error("Expected key2 to exist")
	}
}

// 测试获取不存在的键
func TestGetNonExistentKey(t *testing.T) {
	cache := NewTimeoutCache[string, int](5)
	val, ok := cache.Get("nonexistent")
	if ok {
		t.Error("Expected false for non-existent key")
	}
	if val != 0 { // int类型的零值
		t.Errorf("Expected zero value, got %v", val)
	}
}

// 测试永不过期的元素(deadline=0)
func TestNeverExpire(t *testing.T) {
	cache := NewTimeoutCache[string, int](5)
	cache.Set("permanent", 100, 0) // deadline=0表示永不过期

	time.Sleep(2 * time.Second)
	val, ok := cache.Get("permanent")
	if !ok || val != 100 {
		t.Errorf("Expected permanent key to exist with value 100, got %v", val)
	}
}

// 测试并发访问
func TestConcurrentAccess(t *testing.T) {
	cache := NewTimeoutCache[string, int](100)
	var wg sync.WaitGroup

	// 并发写入
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			key := "key" + string(rune('0'+idx%10)) // 制造一些键冲突
			cache.Set(key, idx, time.Now().Add(10*time.Second).UnixMilli())
		}(i)
	}

	wg.Wait()

	// 验证数据完整性
	if len(cache.cache) > 10 { // 因为只有10个不同的键
		t.Errorf("Expected at most 10 elements, got %d", len(cache.cache))
	}

	// 并发读取
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			key := "key" + string(rune('0'+idx%10))
			_, _ = cache.Get(key)
		}(i)
	}

	wg.Wait()
}

// 测试批量更新
func TestBulkUpdate(t *testing.T) {
	cache := NewTimeoutCache[string, int](100)

	// 初始设置
	for i := 0; i < 50; i++ {
		key := "key" + string(rune('0'+i))
		i := rand.Intn(100) // 随机值
		cache.Set(key, i, time.Now().Add(time.Duration(i)*time.Millisecond).UnixMilli())
		time.Sleep(1 * time.Millisecond)
	}

	// 批量更新
	for i := 0; i < 50; i++ {
		key := "key" + string(rune('0'+i))
		cache.Set(key, i*10, time.Now().Add(10*time.Second).UnixMilli())
		time.Sleep(10 * time.Millisecond) // 确保每个更新的时间不同
	}

	for i := 0; i < 50; i++ {
		fmt.Println("Cache key:", "key"+string(rune('0'+i)))
		fmt.Println("Cache value:", cache.cache["key"+string(rune('0'+i))])
		fmt.Println("Cache deadline:", cache.cache["key"+string(rune('0'+i))].deadline)
		fmt.Println("Cache index:", cache.cache["key"+string(rune('0'+i))].index)
	}
	// 验证更新结果
	for i := 0; i < 5; i++ {
		key := "key" + string(rune('0'+i))
		val, ok := cache.Get(key)
		if !ok || val != i*10 {
			t.Errorf("Expected %d for key %s, got %d", i*10, key, val)
		}
	}
}
