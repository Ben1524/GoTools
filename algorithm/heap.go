package algorithm

import (
	"cmp"
)

type MinHeap[T cmp.Ordered] struct {
	heap []T
}

/*
	堆的特性：
	left = 2*i + 1
	right = 2*i + 2
	parent = (i - 1) / 2
*/

func NewMinHeap[T cmp.Ordered]() *MinHeap[T] {
	return &MinHeap[T]{heap: []T{}}
}

func (h *MinHeap[T]) adjustDown(index int) { // 向下调整堆，使得节点下沉
	n := len(h.heap)
	smallest := index
	left := 2*index + 1  // 左子节点索引
	right := 2*index + 2 // 右子节点索引

	if left < n && h.heap[left] < h.heap[smallest] {
		smallest = left
	}
	if right < n && h.heap[right] < h.heap[smallest] {
		smallest = right
	}

	// 如果当前节点不是最小的，交换并继续调整
	if smallest != index {
		h.heap[index], h.heap[smallest] = h.heap[smallest], h.heap[index]
		h.adjustDown(smallest)
	}
}

func (h *MinHeap[T]) adjustUp(index int) { // 向上调整堆
	for index > 0 {
		parent := (index - 1) / 2 // 父节点索引
		if h.heap[index] >= h.heap[parent] {
			break // 如果当前节点大于等于父节点，调整结束
		}
		// 交换当前节点和父节点
		h.heap[index], h.heap[parent] = h.heap[parent], h.heap[index]
		index = parent // 继续向上调整
	}
}

//=====================优先队列=====================

type Item[T cmp.Ordered] struct {
	val      T
	priority int // 优先级
	index    int // 在堆中的索引
}

type PriorityQueueInt []*Item[int]

// 这是heap.Interface接口的实现
func (pq PriorityQueueInt) Len() int { return len(pq) }
func (pq PriorityQueueInt) Less(i, j int) bool {
	return pq[i].val < pq[j].val // 优先级小的在前
}
func (pq PriorityQueueInt) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index = i
	pq[j].index = j
}
func (pq *PriorityQueueInt) Push(x any) {
	n := len(*pq)
	item := x.(*Item[int])
	item.index = n
	*pq = append(*pq, item)
}

func (pq *PriorityQueueInt) Pop() any {
	old := *pq
	n := len(old)
	item := old[n-1]
	old[n-1] = nil  // 避免内存泄漏,去除对旧元素的引用
	item.index = -1 // 安全起见
	*pq = old[0 : n-1]
	return item
}
