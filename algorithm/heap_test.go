package algorithm

import (
	"container/heap"
	"fmt"
	"testing"
)

func TestContainerHeap(t *testing.T) {
	var pq = &PriorityQueueInt{}
	heap.Push(pq, &Item[int]{val: 4})
	heap.Push(pq, &Item[int]{val: 2})
	heap.Push(pq, &Item[int]{val: 5})
	heap.Push(pq, &Item[int]{val: 1})

	fmt.Println("Priority Queue after pushes:")
	for pq.Len() > 0 {
		item := heap.Pop(pq).(*Item[int])
		fmt.Printf("Value: %d, Priority: %d\n", item.val, item.priority)
	}

}
