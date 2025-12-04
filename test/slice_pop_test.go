package test

import (
	"fmt"
	"testing"
)

func TestSlicePoP(t *testing.T) {
	// 初始切片
	convergencePoints := []string{"A", "B", "C", "D"}
	fmt.Println("Before:", convergencePoints, "len:", len(convergencePoints), "cap:", cap(convergencePoints))

	// 弹出第一个元素
	convergencePoints = convergencePoints[1:]
	fmt.Println("After:", convergencePoints, "len:", len(convergencePoints), "cap:", cap(convergencePoints))

	// 多次弹出
	convergencePoints = convergencePoints[1:] // 弹出 B
	convergencePoints = convergencePoints[1:] // 弹出 C
	fmt.Println("Final:", convergencePoints)  // 输出: [D]
}
