package algorithm

import (
	"cmp"
)

func BinarySearch[T cmp.Ordered](arr []T, target T) int {
	left, right := 0, len(arr)-1

	for left <= right {
		mid := left + (right-left)/2

		if arr[mid] == target {
			return mid // 找到目标元素，返回索引
		}
		if arr[mid] < target {
			left = mid + 1 // 目标在右半部分
		} else {
			right = mid - 1 // 目标在左半部分
		}
	}

	return -1 // 未找到目标元素，返回 -1
}
