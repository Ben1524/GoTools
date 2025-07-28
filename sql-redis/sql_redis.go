package sql_redis

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/golang/groupcache/singleflight"
	"github.com/redis/go-redis/v9"
	"log"
	"math"
	"time"
)

const (
	CacheKeyBaseExpiration  = 24 * time.Hour   // 缓存过期时间
	CacheKeyRoundExpiration = 10 * time.Minute // 缓存轮询过期时间
)

var (
	ErrorNotFind     = errors.New("not found in cache or database")
	ErrorPlaceholder = errors.New("placeholder value, not found in cache or database")
)

type Cache struct {
	redisCli   *redis.Client
	singleCall singleflight.Group
}

func NewCache(redisCli *redis.Client) *Cache {
	return &Cache{
		redisCli:   redisCli,
		singleCall: singleflight.Group{},
	}
}

func (c *Cache) Set(key string, v interface{}) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	var expiration time.Duration = CacheKeyBaseExpiration
	expiration += time.Duration(int(math.Round(20))) * CacheKeyRoundExpiration // 根据key长度增加过期时间
	return c.redisCli.Set(context.Background(), key, data, expiration).Err()
}

func (c *Cache) Get(key string, v interface{}) error {
	data, err := c.redisCli.Get(context.Background(), key).Result()
	if errors.Is(err, redis.Nil) {
		return ErrorNotFind
	}
	if err != nil {
		return err
	}
	if data == `"*"` {
		return ErrorPlaceholder
	}
	if err = json.Unmarshal([]byte(data), v); err == nil {
		return nil
	}
	// 如果反序列化失败，可能是因为数据格式不正确，删除缓存
	if err = c.redisCli.Del(context.Background(), key).Err(); err != nil {
		log.Println("del redis key  : ", key, " err :", err.Error())
		return err
	}
	return errors.New("failed to unmarshal data from cache")
}

type callFunc func(v interface{}) error

// v是一个指针，指向要存储或查询的数据结构，v承担着保存查询结果和提供查询参数的双重角色
func (c *Cache) TakeWithFunc(key string, v interface{}, dbQueryFunc callFunc, cacheVal callFunc) error {
	val, err := c.singleCall.Do(key, func() (interface{}, error) {
		if err := c.Get(key, v); err != nil {
			if errors.Is(err, ErrorPlaceholder) {
				return nil, ErrorNotFind
			} else if !errors.Is(err, ErrorNotFind) {
				return nil, err
			}
			if err := dbQueryFunc(v); err == ErrorNotFind {
				// 如果查询函数返回了 ErrorNotFind，表示数据不存在
				// 则将缓存值设置为占位符
				if err := c.Set(key, "*"); err != nil {
					return nil, err
				}
				return nil, ErrorNotFind
			} else if err != nil {
				return nil, err
			}

			if err = cacheVal(v); err != nil {
				log.Println("cacheVal error:", err.Error())
				return nil, err
			}
		}
		return json.Marshal(v) // 返回序列化后的数据
	})
	if err != nil {
		if errors.Is(err, ErrorNotFind) {
			return ErrorNotFind
		}
		return err
	}
	return json.Unmarshal(val.([]byte), v) // 将序列化后的数据反序列化到 v 中
}

func (c *Cache) Take(key string, v interface{}, dbQueryFunc callFunc) error {
	return c.TakeWithFunc(key, v, dbQueryFunc, func(v interface{}) error {
		return c.Set(key, v)
	}) // 默认使用 Set 方法将数据存入缓存
}
