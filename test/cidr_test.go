package test

import (
	"fmt"
	"strconv"
	"testing"
)

func TestCidr(t *testing.T) {
	convergencePoints := []int{100} // 只有一个元素
	fmt.Println("原始切片:", convergencePoints, "长度:", len(convergencePoints))

	convergencePoints = convergencePoints[1:] // 删除第一个元素
	fmt.Println("操作后:", convergencePoints, "长度:", len(convergencePoints))
	// 输出: [] 长度: 0
}

func TestExample(t *testing.T) {
	for index := range 20 {
		fmt.Println("index", index)
	}
}

func TestTime(t *testing.T) {
	value, _ := strconv.ParseBool("true")
	fmt.Println(value)
}
