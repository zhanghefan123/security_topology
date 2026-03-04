package corrupt_decider

import (
	"fmt"
	"math/rand"
	"time"
)

// CreateUniformCorruptDecider 进行均匀分布的生成的损坏者的构建
func CreateUniformCorruptDecider(start, end float64) (*CorruptDecider, error) {
	// 相反则返回错误
	if start > end {
		return nil, fmt.Errorf("start must < end")
	}

	// 内部创建随机源
	rand.Seed(time.Now().UnixNano())

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
	corruptDecider := &CorruptDecider{
		DecideFunction: (*DecideFunction)(&decideFunction),
	}

	// 返回闭包
	return corruptDecider, nil
}
