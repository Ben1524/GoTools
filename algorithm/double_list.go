package algorithm

import (
	"cmp"
)

type DoubleListNode[T cmp.Ordered] struct {
	Value T
	Prev  *DoubleListNode[T]
	Next  *DoubleListNode[T]
}

type DoubleList[T cmp.Ordered] struct {
	Head *DoubleListNode[T]
	Tail *DoubleListNode[T]
	Size int
}

func NewDoubleList[T cmp.Ordered]() *DoubleList[T] {
	return &DoubleList[T]{
		Head: nil,
		Tail: nil,
		Size: 0,
	}
}

func (dl *DoubleList[T]) Append(value T) {
	node := &DoubleListNode[T]{Value: value}
	if dl.Head == nil {
		dl.Head = node
		dl.Tail = node
	} else {
		dl.Tail.Next = node
		node.Prev = dl.Tail
		dl.Tail = node
	}
	dl.Size++
}

func (dl *DoubleList[T]) Prepend(value T) {
	node := &DoubleListNode[T]{Value: value}
	if dl.Head == nil {
		dl.Head = node
		dl.Tail = node
	} else {
		node.Next = dl.Head
		dl.Head.Prev = node
		dl.Head = node
	}
	dl.Size++
}

func (dl *DoubleList[T]) Remove(node *DoubleListNode[T]) {
	if node == nil || dl.Head == nil {
		return
	}

	if node.Prev != nil {
		node.Prev.Next = node.Next
	} else {
		dl.Head = node.Next // Node is head
	}

	if node.Next != nil {
		node.Next.Prev = node.Prev
	} else {
		dl.Tail = node.Prev // Node is tail
	}

	dl.Size--
	node.Prev = nil
	node.Next = nil
	node.Value = *new(T) // 避免继续引用另一个 值
}

func (dl *DoubleList[T]) Find(value T) *DoubleListNode[T] {
	current := dl.Head
	for current != nil {
		if current.Value == value {
			return current
		}
		current = current.Next
	}
	return nil // 未找到值
}

func (dl *DoubleList[T]) Clear() {
	for current := dl.Head; current != nil; {
		next := current.Next
		current.Prev = nil
		current.Next = nil
		current.Value = *new(T) // 避免继续引用另一个 值
		current = next
	}
	dl.Head = nil
	dl.Tail = nil
	dl.Size = 0
}

func (dl *DoubleList[T]) IsEmpty() bool {
	return dl.Size == 0
}

func (dl *DoubleList[T]) Length() int {
	return dl.Size
}

func (dl *DoubleList[T]) InsertAfter(node *DoubleListNode[T], value T) {
	if node == nil {
		return
	}
	newNode := &DoubleListNode[T]{Value: value}
	newNode.Prev = node
	newNode.Next = node.Next

	if node.Next != nil {
		node.Next.Prev = newNode
	} else {
		dl.Tail = newNode // Node is tail
	}

	node.Next = newNode
	dl.Size++
}
func (dl *DoubleList[T]) InsertBefore(node *DoubleListNode[T], value T) {
	if node == nil {
		return
	}
	newNode := &DoubleListNode[T]{Value: value}
	newNode.Next = node
	newNode.Prev = node.Prev

	if node.Prev != nil {
		node.Prev.Next = newNode
	} else {
		dl.Head = newNode // Node is head
	}

	node.Prev = newNode
	dl.Size++
}
