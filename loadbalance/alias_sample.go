package loadbalance

import "math/rand"

// alias采样

type AliasSampler struct {
	accept []float64 // 接受概率,原本的概率
	alias  []int     // 别名,用于快速采样，从哪里填补过来的
}

func NewAliasSampler(probabilities []float64) *AliasSampler {
	n := len(probabilities)
	accept := make([]float64, n)
	alias := make([]int, n)

	// 计算总概率
	total := 0.0
	for i, p := range probabilities {
		total += p
		alias[i] = -1   // 初始化别名为自身
		accept[i] = 1.0 // 初始化接受概率为1.0
	}

	// 归一化概率
	for i, p := range probabilities {
		accept[i] = p * float64(n) / total // 归一化到 [0, n]
	}

	small := []int{} // 小于1的索引
	large := []int{} // 大于等于1的索引

	for i, p := range accept {
		if p < 1.0 {
			small = append(small, i)
		} else {
			large = append(large, i)
		}
	}

	for len(small) > 0 && len(large) > 0 {
		smallIndex := small[len(small)-1] // 取出最后一个小于1的索引
		small = small[:len(small)-1]

		largeIndex := large[len(large)-1] // 取出最后一个大于等于1的索引
		large = large[:len(large)-1]

		alias[smallIndex] = largeIndex                                     // 别名指向大于等于1的索引，从largeIndex填补过来的
		accept[largeIndex] = accept[largeIndex] + accept[smallIndex] - 1.0 // 去除填补部分
		// 接受数组中，largeIndex的概率减少了(1 - accept[smallIndex])，因为accept[smallIndex]部分已经被smallIndex接受了
		if accept[largeIndex] < 1.0 { // 如果填补给对方后的概率小于1，说明largeIndex也需要被填补
			small = append(small, largeIndex)
		} else {
			large = append(large, largeIndex)
		}
	}

	return &AliasSampler{accept: accept, alias: alias}
}

// 生成索引i
// 生成一个0-1之间的随机数f
// 如果f < accept[i]，则返回i
// 否则返回alias[i]，即别名索引
func (as *AliasSampler) Sample() int {
	i := rand.Intn(len(as.alias)) // 随机选择一个索引
	f := rand.Float64()           // 随机选择一个概率
	if f < as.accept[i] {
		return i // 直接返回索引
	} else {
		return as.alias[i] // 返回别名索引
	}
}
