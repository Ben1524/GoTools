package algorithm

import (
	"cmp"
	"sort"
)

func JaccardSimilarity[T cmp.Ordered](setA, setB []T) float64 {
	if len(setA) == 0 && len(setB) == 0 {
		return 1.0 // 两个空集合的相似度为 1
	}

	if len(setA) == 0 || len(setB) == 0 {
		return 0.0 // 一个空集合和非空集合的相似度为 0
	}

	setAMap := make(map[T]struct{})
	for _, item := range setA {
		setAMap[item] = struct{}{}
	}

	setBMap := make(map[T]struct{})
	for _, item := range setB {
		setBMap[item] = struct{}{}
	}

	intersectionCount := 0
	for item := range setAMap {
		if _, exists := setBMap[item]; exists {
			intersectionCount++
		}
	}

	unionCount := len(setAMap) + len(setBMap) - intersectionCount

	return float64(intersectionCount) / float64(unionCount)
}

func JaccardSimilarityBySorted[T cmp.Ordered](setA, setB []T) float64 {
	if len(setA) == 0 && len(setB) == 0 {
		return 1.0 // 两个空集合的相似度为 1
	}

	if len(setA) == 0 || len(setB) == 0 {
		return 0.0 // 一个空集合和非空集合的相似度为 0
	}

	sort.Slice(setA, func(i, j int) bool { return setA[i] < setA[j] })
	sort.Slice(setB, func(i, j int) bool { return setB[i] < setB[j] })

	var intersectionCount int
	for i, j := 0, 0; i < len(setA) && j < len(setB); {
		if setA[i] == setB[j] {
			intersectionCount++
			i++
			j++
		} else if setA[i] < setB[j] {
			i++
		} else {
			j++
		}
	}
	unionCount := len(setA) + len(setB) - intersectionCount
	return float64(intersectionCount) / float64(unionCount)
}
