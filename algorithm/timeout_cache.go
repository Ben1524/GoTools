package algorithm

import (
	"container/heap"
	"sync"
	"time"
)

type HeapNode[K comparable, T any] struct {
	Key      K // 唯一标识符
	Value    T
	index    int   // 在堆中的索引
	deadline int64 // 过期时间戳
}

type TimeoutCache[K comparable, T any] struct {
	Capacity   int
	heap       []*HeapNode[K, T]     // 最小堆
	cache      map[K]*HeapNode[K, T] // 哈希表
	sync.Mutex                       // 互斥锁，确保线程安全
}

// 使用heap.Interface实现最小堆的接口
func (tc *TimeoutCache[K, T]) Len() int {
	return len(tc.heap)
}
func (tc *TimeoutCache[K, T]) Less(i, j int) bool {
	return tc.heap[i].deadline < tc.heap[j].deadline
}
func (tc *TimeoutCache[K, T]) Swap(i, j int) {
	tc.heap[i], tc.heap[j] = tc.heap[j], tc.heap[i]
	tc.heap[i].index = i // 更新索引
	tc.heap[j].index = j // 更新索引
	tc.cache[tc.heap[i].Key] = tc.heap[i]
	tc.cache[tc.heap[j].Key] = tc.heap[j]
}
func (tc *TimeoutCache[K, T]) Push(x any) {
	node := x.(*HeapNode[K, T])
	node.index = len(tc.heap) // 设置新节点的索引为当前堆的长度,因为新节点将被添加到堆的末尾
	tc.cache[node.Key] = node
	tc.heap = append(tc.heap, node)
}
func (tc *TimeoutCache[K, T]) Pop() any {
	if len(tc.heap) == 0 {
		return new(HeapNode[K, T]) // 返回一个空的HeapNode
	}
	node := tc.heap[len(tc.heap)-1]
	tc.heap = tc.heap[:len(tc.heap)-1]
	delete(tc.cache, node.Key)
	return node
}

func NewTimeoutCache[K comparable, T any](capacity int) *TimeoutCache[K, T] {
	return &TimeoutCache[K, T]{
		Capacity: capacity,
		heap:     make([]*HeapNode[K, T], 0, capacity),
		cache:    make(map[K]*HeapNode[K, T]),
	}
}

func (tc *TimeoutCache[K, T]) updateHeapIndices() {
	// 更新堆中所有节点的索引
	for i, node := range tc.heap {
		node.index = i
		tc.cache[node.Key] = node // 确保哈希表中的索引是最新的
	}
}

// 懒惰删除：在获取时检查过期时间
func (tc *TimeoutCache[K, T]) Get(key K) (T, bool) {
	tc.Lock()
	defer tc.Unlock()
	currentTimeMillis := time.Now().UnixMilli()
	if node, exists := tc.cache[key]; exists {
		if node.deadline > 0 && node.deadline < currentTimeMillis {
			// 如果节点已过期，删除它
			heap.Remove(tc, node.index)
			delete(tc.cache, key) // 从哈希表中删除
			return *new(T), false
		}
		return node.Value, true
	}
	return *new(T), false
}

func (tc *TimeoutCache[K, T]) Set(key K, value T, deadline int64) {
	tc.Lock()
	defer tc.Unlock()

	if node, exists := tc.cache[key]; exists {
		// 移除最小堆中的旧节点
		heap.Remove(tc, node.index)
		// 更新已存在的节点
		node.Value = value
		node.deadline = deadline
		// 重新添加到堆中
		heap.Push(tc, node)
		tc.cache[key] = node
	} else {
		if len(tc.heap) >= tc.Capacity {
			// 如果堆已满，移除最小的节点
			oldest := heap.Pop(tc).(*HeapNode[K, T])
			delete(tc.cache, oldest.Key)
		}
		// 创建新节点并添加到堆和哈希表中
		newNode := &HeapNode[K, T]{Key: key, Value: value, deadline: deadline}
		heap.Push(tc, newNode)
		tc.cache[key] = newNode
	}
}
