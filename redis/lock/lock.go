package lock

import (
	"context"
	"github.com/redis/go-redis/v9"
	"time"
)

type RedisLockServer interface {
	TryLock() (bool, error)
	UnLock() error
	GetLockKey() string
	GetLockVal() string
}

type RedisLock struct {
	redisCli *redis.Client // Redis客户端
	timeout  time.Duration // 锁的超时时间
	key      string        // 锁的键
	value    string        // 锁的值
}

func NewRedisLock(redisCli *redis.Client, key string, value string, timeout time.Duration) *RedisLock {
	return &RedisLock{
		redisCli: redisCli,
		timeout:  timeout,
		key:      key,
		value:    value,
	}
}

func (rl *RedisLock) Trylock() (bool, error) {
	// 使用SETNX命令尝试获取锁
	// setnx = SET if Not eXists
	ok, err := rl.redisCli.SetNX(context.Background(), rl.key, rl.value, rl.timeout).Result()
	if err != nil {
		return false, err
	}
	if !ok {
		return false, nil // 锁已被其他客户端持有
	}
	return true, nil // 成功获取锁
}

func (rl *RedisLock) Unlock() error {
	// 使用Lua脚本确保只有持有锁的客户端才能释放锁
	luaScript := `
	if redis.call("get", KEYS[1]) == ARGV[1] then
		return redis.call("del", KEYS[1])
	else
		return 0
	end`
	result, err := rl.redisCli.Eval(context.Background(), luaScript, []string{rl.key}, rl.value).Result() // 执行Lua脚本
	if err != nil {
		return err
	}
	if result.(int64) == 0 {
		return redis.Nil // 锁不存在或不是当前客户端持有的锁
	}
	return nil // 成功释放锁
}

func (lock *RedisLock) GetLockKey() string {
	return lock.key
}

func (lock *RedisLock) GetLockVal() string {
	return lock.value
}
