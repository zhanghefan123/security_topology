package test

import (
	"fmt"
	"math/rand"
	"testing"
)

func TestGenerateSampleSeq(t *testing.T) {
	numRouters := 5
	sequence := make([]int, 100)

	// 1. 确定性分配：确保每个 Router 分配到的次数尽可能相等
	for i := 0; i < 100; i++ {
		sequence[i] = i % numRouters
	}

	// 2. 随机洗牌：打乱顺序，让攻击者无法预测
	// 使用加密安全的随机数种子更佳
	rand.Shuffle(len(sequence), func(i, j int) {
		sequence[i], sequence[j] = sequence[j], sequence[i]
	})

	fmt.Println(sequence)
}
