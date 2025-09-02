package main

import (
	"fmt"
	"time"
)

func withoutDone() {
	fmt.Println("=== 不使用 <-done 的情况 ===")
	
	// 模拟设置事件监听器
	go func() {
		// 模拟事件处理需要一些时间
		time.Sleep(100 * time.Millisecond)
		fmt.Println("事件回调执行了！")
	}()
	
	// 模拟Set操作
	fmt.Println("Set操作完成")
	
	// 没有等待事件处理完成
	fmt.Println("函数结束")
}

func withDone() {
	fmt.Println("=== 使用 <-done 的情况 ===")
	done := make(chan bool)
	
	// 模拟设置事件监听器
	go func() {
		// 模拟事件处理需要一些时间
		time.Sleep(100 * time.Millisecond)
		fmt.Println("事件回调执行了！")
		done <- true
	}()
	
	// 模拟Set操作
	fmt.Println("Set操作完成")
	
	// 等待事件处理完成
	<-done
	fmt.Println("函数结束")
}

func main() {
	withoutDone()
	time.Sleep(200 * time.Millisecond) // 给第一个测试一些时间
	
	fmt.Println()
	
	withDone()
}
