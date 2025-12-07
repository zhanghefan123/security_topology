package test

import (
	"fmt"
	"strconv"
	"sync"
	"testing"
	"time"
)

func TestCidr(t *testing.T) {
	var wg sync.WaitGroup

	// 创建共享的计时器
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	// 吞吐量计算协程
	wg.Add(1)
	go func() {
		defer wg.Done()
		for _ = range ticker.C {
			fmt.Println("test1")
		}
	}()

	// 接口收包速率计算协程
	wg.Add(1)
	go func() {
		defer wg.Done()
		for _ = range ticker.C {
			fmt.Println("test2")
		}
	}()

	// 运行一段时间后停止
	time.Sleep(10 * time.Second)
	wg.Wait()
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
