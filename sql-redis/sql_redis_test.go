package sql_redis

import (
	"errors"
	"fmt"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"sync"
	"testing"
)

type Test struct {
	ID         int    `gorm:"primaryKey" json:"id"`
	UserName   string `json:"user_name"`
	Pwd        string `json:"pwd"`
	CreateTime int64  `json:"create_time"`
	UpdateTime int64  `json:"update_time"`
}

func (t *Test) TableName() string {
	return "test"
}

func TestGet(t *testing.T) {
	redisCli := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	cache := NewCache(redisCli)

	testData := &Test{
		ID:         1,
		UserName:   "test_user",
		Pwd:        "test_password",
		CreateTime: 1633036800,
		UpdateTime: 1633036800,
	}

	err := cache.Set("test_key", testData)
	if err != nil {
		t.Fatalf("Failed to set cache: %v", err)
	}

	var retrievedData Test
	err = cache.Get("test_key", &retrievedData)
	if err != nil {
		t.Fatalf("Failed to get cache: %v", err)
	}

	if retrievedData.UserName != testData.UserName {
		t.Errorf("Expected %s, got %s", testData.UserName, retrievedData.UserName)
	}
}

func TestGet_NotFound(t *testing.T) {
	sql := "root:123456@tcp(localhost:3306)/GoTest?timeout=5s&readTimeout=5s&writeTimeout=5s&parseTime=true&loc=Local&charset=utf8,utf8mb4"
	_, err := gorm.Open(mysql.Open(sql), nil)
	if err != nil {
		panic(err)
	}
	cache := NewCache(redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	}))
	var retrievedData Test
	err = cache.Get("non_existent_key", &retrievedData)
	if err != nil && !errors.Is(err, ErrorNotFind) {
		t.Fatalf("Expected ErrorNotFind, got %v", err)
	}
	if retrievedData.ID != 0 {
		t.Errorf("Expected retrievedData.ID to be 0, got %d", retrievedData.ID)
	}
}

func TestGet_Update(t *testing.T) {
	sql := "root:123456@tcp(localhost:3306)/GoTest?timeout=5s&readTimeout=5s&writeTimeout=5s&parseTime=true&loc=Local&charset=utf8,utf8mb4"
	db, err := gorm.Open(mysql.Open(sql), nil)
	if err != nil {
		panic(err)
	}
	cache := NewCache(redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	}))
	var retrievedData = &Test{ // ID作为查询条件，查不到就更新缓存，查到了就更新retrievedData
		ID:         1,
		UserName:   "test_user",
		Pwd:        "test_password",
		CreateTime: 1633036800,
		UpdateTime: 1633036800,
	}
	if db.AutoMigrate(&Test{}); db.Error != nil {
		t.Fatalf("Failed to migrate database: %v", db.Error)
	}
	dbQueryFunc := func(v interface{}) error {
		fmt.Println("dbQueryFunc called")
		testData, ok := v.(*Test)
		if !ok {
			return errors.New("invalid type for dbQueryFunc")
		}
		result := db.First(testData, testData.ID)
		if result.Error != nil {
			if errors.Is(result.Error, gorm.ErrRecordNotFound) {
				return ErrorNotFind
			}
			return result.Error
		}
		return nil
	}
	cacheVal := func(v interface{}) error {
		testData, ok := v.(*Test)
		if !ok {
			return errors.New("invalid type for cacheVal")
		}
		return cache.Set("test_key", testData)
	}
	var waitGo sync.WaitGroup
	for i := 0; i < 10; i++ {
		waitGo.Add(1)
		//err = cache.TakeWithFunc("test_key", retrievedData, dbQueryFunc, cacheVal)
		go func() {
			err = cache.TakeWithFunc("test_key", retrievedData, dbQueryFunc, cacheVal)
			if err != nil {
				t.Errorf("Failed to take with func: %v", err)
			}
			waitGo.Done()
		}()
	}
	waitGo.Wait()
	if err != nil {
		t.Fatalf("Failed to take with func: %v", err)
	}

}
