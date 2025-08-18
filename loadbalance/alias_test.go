package loadbalance

import (
	"math"
	"testing"
)

// 测试正常概率分布的采样结果
func TestAliasSampler_NormalDistribution(t *testing.T) {
	// 测试概率分布：索引0的概率是0.5，索引1是0.3，索引2是0.2
	probabilities := []float64{0.5, 0.3, 0.2}
	sampler := NewAliasSampler(probabilities)

	// 大量采样以验证分布
	sampleCount := 1000000
	counts := make([]int, len(probabilities))

	for i := 0; i < sampleCount; i++ {
		idx := sampler.Sample()
		if idx < 0 || idx >= len(probabilities) {
			t.Errorf("采样结果无效: %d，超出有效范围", idx)
		}
		counts[idx]++
	}

	// 计算实际频率
	actual := make([]float64, len(probabilities))
	for i, c := range counts {
		actual[i] = float64(c) / float64(sampleCount)
	}

	// 检查每个概率是否在预期范围内（允许一定误差）
	tolerance := 0.01 // 1%的误差容忍度
	for i, expected := range probabilities {
		if math.Abs(actual[i]-expected) > tolerance {
			t.Errorf("索引 %d 的采样频率不符合预期: 实际=%.4f, 预期=%.4f",
				i, actual[i], expected)
		}
	}
}

// 测试单个元素的情况
func TestAliasSampler_SingleElement(t *testing.T) {
	probabilities := []float64{1.0}
	sampler := NewAliasSampler(probabilities)

	// 多次采样，应该总是返回0
	for i := 0; i < 1000; i++ {
		idx := sampler.Sample()
		if idx != 0 {
			t.Errorf("采样结果错误: 预期0，实际%d", idx)
		}
	}
}

// 测试概率总和不为1的情况（应该被归一化）
func TestAliasSampler_UnnormalizedProbabilities(t *testing.T) {
	// 这些概率总和为2.0，但应该被正确归一化
	probabilities := []float64{1.0, 0.6, 0.4}
	sampler := NewAliasSampler(probabilities)

	sampleCount := 1000000
	counts := make([]int, len(probabilities))

	for i := 0; i < sampleCount; i++ {
		idx := sampler.Sample()
		counts[idx]++
	}

	// 预期概率应该是 [0.5, 0.3, 0.2]
	expected := []float64{0.5, 0.3, 0.2}
	actual := make([]float64, len(probabilities))
	for i, c := range counts {
		actual[i] = float64(c) / float64(sampleCount)
	}

	tolerance := 0.01
	for i := range expected {
		if math.Abs(actual[i]-expected[i]) > tolerance {
			t.Errorf("索引 %d 的采样频率不符合预期: 实际=%.4f, 预期=%.4f",
				i, actual[i], expected[i])
		}
	}
}

// 测试空概率数组（应该panic）
func TestAliasSampler_EmptyProbabilities(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("空概率数组应该触发panic，但没有")
		}
	}()

	// 尝试创建空的采样器
	NewAliasSampler([]float64{})
}
