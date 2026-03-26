package probs

import "math/rand"

// SampleDiscrete 根据输入的概率分布进行离散采样，返回采样结果的下标
func SampleDiscrete(probabilities []float64) int {
	r := rand.Float64() // 生成 [0,1) 均匀随机数
	sum := 0.0          // 累计概率

	for i, p := range probabilities {
		sum += p      // 把概率一个个累加
		if r <= sum { // 如果随机数落在当前累计区间
			return i // 返回当前下标
		}
	}

	return len(probabilities) - 1 // 防止浮点误差兜底
}
