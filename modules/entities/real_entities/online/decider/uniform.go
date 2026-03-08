package decider

import (
	"fmt"
	"math/rand"
)

// CreateUniformDecider 进行均匀分布的生成的损坏者的构建
func CreateUniformDecider(start, end float64, nodeIndex int) (*ActionDecider, error) {
	// 相反则返回错误
	if start > end {
		return nil, fmt.Errorf("start must < end")
	}

	// 内部创建随机源 (这里使用 nodeIndex 作为种子, 确保每次运行结果一致)
	rand.Seed(int64(nodeIndex))

	// 创建 decide function
	decideFunction := func() int {
		currentDropRate := start + rand.Float64()*(end-start)
		if rand.Float64() < currentDropRate {
			return 1
		} else {
			return 0
		}
	}

	// 创建 decider
	decider := &ActionDecider{
		DecideFunction: (*DecideFunction)(&decideFunction),
	}

	// 返回闭包
	return decider, nil
}
