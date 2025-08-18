package algorithm

import (
	"container/list"
)

type cacheNode[K comparable, V any] struct {
	key   K
	value V
}

type LRUCache[K comparable, V any] struct {
	capacity int
	cache    map[K]*cacheNode[K, V]
	list     *list.List
}

func NewLRUCache[K comparable, V any](capacity int) *LRUCache[K, V] {
	return &LRUCache[K, V]{
		capacity: capacity,
		cache:    make(map[K]*cacheNode[K, V]),
		list:     list.New(),
	}
}
func (lru *LRUCache[K, V]) Get(key K) (V, bool) {
	if node, exists := lru.cache[key]; exists {
		// 移动到列表头部，表示最近使用
		for e := lru.list.Front(); e != nil; e = e.Next() {
			if e.Value.(*cacheNode[K, V]).key == key {
				lru.list.MoveToFront(e)
				break
			}
		}
		return node.value, true
	}
	var zero V
	return zero, false
}

func (lru *LRUCache[K, V]) Put(key K, value V) {
	if node, exists := lru.cache[key]; exists {
		// 更新值并移动到列表头部
		node.value = value
		for e := lru.list.Front(); e != nil; e = e.Next() {
			if e.Value.(*cacheNode[K, V]).key == key {
				lru.list.MoveToFront(e)
				break
			}
		}
	} else {
		if len(lru.cache) >= lru.capacity {
			// 移除最久未使用的元素
			back := lru.list.Back()
			if back != nil {
				lru.list.Remove(back)
				delete(lru.cache, back.Value.(*cacheNode[K, V]).key)
			}
		}
		newNode := &cacheNode[K, V]{key: key, value: value}
		lru.list.PushFront(newNode)
		lru.cache[key] = newNode
	}
}
