package lock

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"github.com/redis/go-redis/v9"
	"sync"
	"testing"
	"time"
)

// 生成随机值作为锁的value
func generateRandomValue() string {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		panic(err)
	}
	return base64.URLEncoding.EncodeToString(b)
}

// 测试使用真实Redis，需要确保本地有Redis服务在运行
func TestRedisLock_RealRedis(t *testing.T) {
	// 连接本地Redis
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	// 清理测试数据
	defer client.Del(context.Background(), "test_lock").Err()

	key := "test_lock"
	value := generateRandomValue()
	timeout := 5 * time.Second

	lock := NewRedisLock(client, key, value, timeout)

	t.Run("Trylock and Unlock", func(t *testing.T) {
		// 尝试获取锁
		ok, err := lock.Trylock()
		if err != nil {
			t.Fatalf("获取锁失败: %v", err)
		}
		if !ok {
			t.Error("预期获取锁成功，但实际失败")
		}

		// 验证锁确实存在
		val, err := client.Get(context.Background(), key).Result()
		if err != nil {
			t.Fatalf("获取锁值失败: %v", err)
		}
		if val != value {
			t.Errorf("锁值不匹配，预期 %s，实际 %s", value, val)
		}

		// 释放锁
		err = lock.Unlock()
		if err != nil {
			t.Fatalf("释放锁失败: %v", err)
		}

		// 验证锁已被释放
		_, err = client.Get(context.Background(), key).Result()
		if err != redis.Nil {
			t.Error("预期锁已被释放，但实际仍存在")
		}
	})

	t.Run("Cannot acquire lock twice", func(t *testing.T) {
		// 第一次获取锁
		ok, err := lock.Trylock()
		if err != nil || !ok {
			t.Fatal("第一次获取锁失败")
		}
		defer lock.Unlock()

		// 尝试再次获取同一把锁
		ok2, err := lock.Trylock()
		if err != nil {
			t.Fatalf("第二次获取锁时出错: %v", err)
		}
		if ok2 {
			t.Error("预期第二次获取锁失败，但实际成功了")
		}
	})

	t.Run("Cannot unlock others' lock", func(t *testing.T) {
		// 第一个客户端获取锁
		ok, err := lock.Trylock()
		if err != nil || !ok {
			t.Fatal("获取锁失败")
		}
		defer lock.Unlock()

		// 第二个客户端尝试释放别人的锁
		otherLock := NewRedisLock(client, key, generateRandomValue(), timeout)
		err = otherLock.Unlock()
		if err != redis.Nil {
			t.Errorf("预期释放别人的锁会返回redis.Nil，实际返回: %v", err)
		}

		// 验证锁仍然存在
		_, err = client.Get(context.Background(), key).Result()
		if err != nil {
			t.Error("释放别人的锁后，原锁不应被删除")
		}
	})

	t.Run("Lock expiration", func(t *testing.T) {
		shortTimeout := 1 * time.Second
		tempLock := NewRedisLock(client, key, generateRandomValue(), shortTimeout)

		// 获取锁
		ok, err := tempLock.Trylock()
		if err != nil || !ok {
			t.Fatal("获取锁失败")
		}

		// 等待锁过期
		time.Sleep(shortTimeout + 500*time.Millisecond)

		// 验证锁已过期
		_, err = client.Get(context.Background(), key).Result()
		if err != redis.Nil {
			t.Error("预期锁已过期，但实际仍存在")
		}
	})
}

// 并发测试，验证锁的互斥性
func TestRedisLock_Concurrent(t *testing.T) {
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	key := "concurrent_test_lock"
	timeout := 5 * time.Second

	// 清理测试数据
	defer client.Del(context.Background(), key).Err()

	var (
		counter      int = 0 // 计数器，用于验证锁的互斥性
		wg           sync.WaitGroup
		numGoroutine = 100 // 并发协程数量
	)

	// 多个协程同时竞争锁
	for i := 0; i < numGoroutine; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			value := generateRandomValue()
			lock := NewRedisLock(client, key, value, timeout)

			// 尝试获取锁，最多尝试5次
			for {
				ok, err := lock.Trylock()
				if err != nil {
					t.Errorf("协程 %d 获取锁出错: %v", id, err)
					return
				}
				if ok {
					break // 成功获取锁，跳出循环
				}
			}
			// 获取到锁，模拟临界区操作
			counter++
			time.Sleep(10 * time.Millisecond) // 模拟处理时间

			// 释放锁
			if err := lock.Unlock(); err != nil {
				t.Errorf("协程 %d 释放锁出错: %v", id, err)
			}
		}(i)
	}

	wg.Wait()

	// 验证只有一个协程成功获取到了锁
	if counter != 100 {
		t.Errorf("并发测试失败，预期counter为1，实际为%d", counter)
	} else {
		t.Logf("并发测试成功，counter=%d", counter)
	}
}

// 测试锁的属性获取方法
func TestRedisLock_Getters(t *testing.T) {
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	key := "test_getters_lock"
	value := "test_value"
	timeout := 5 * time.Second

	lock := NewRedisLock(client, key, value, timeout)

	if lock.GetLockKey() != key {
		t.Errorf("GetLockKey返回值不正确，预期 %s，实际 %s", key, lock.GetLockKey())
	}

	if lock.GetLockVal() != value {
		t.Errorf("GetLockVal返回值不正确，预期 %s，实际 %s", value, lock.GetLockVal())
	}
}
