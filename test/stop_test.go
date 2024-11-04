package test

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestStop(t *testing.T) {
	// 创建一个广播通道
	stopChan := make(chan struct{})

	// 使用 WaitGroup 来等待所有协程退出
	var wg sync.WaitGroup

	// 启动 20 个协程，每个协程都会监听 stopChan
	for i := 1; i <= 5; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done() // 确保协程退出时调用 Done()
			for {
				select {
				case <-stopChan:
					fmt.Printf("Goroutine %d received stop signal and is shutting down\n", id)
					return // 接收到关闭信号，退出协程
				default:
					// 模拟工作
					fmt.Printf("Goroutine %d is working\n", id)
					time.Sleep(500 * time.Millisecond)
				}
			}
		}(i)
	}

	// 等待几秒后发送关闭信号
	time.Sleep(2 * time.Second)
	close(stopChan) // 关闭通道，通知所有协程退出

	// 等待所有协程完成
	wg.Wait()

	fmt.Println("All goroutines have been shut down.")
}
