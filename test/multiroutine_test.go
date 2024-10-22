package test

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"zhanghefan123/security_topology/modules/utils/progress_bar"
)

type Satellite struct {
	name string
}

var satellites = []Satellite{
	{
		"zhf",
	},
	{
		"zzg",
	},
	{
		"gyf",
	},
}

func TestMultiRoutine(t *testing.T) {
	Core()
}

func Core() error {
	// 获取卫星数量
	satelliteNumber := len(satellites)
	// 创建 progressBar 描述
	description := fmt.Sprintf("%20s", "start satellites")
	// 创建 progressBar
	progressBar := progress_bar.NewProgressBar(satelliteNumber, description)
	// 创建等待组
	waitGroup := sync.WaitGroup{}
	// 为等待组添加任务
	waitGroup.Add(satelliteNumber)
	// 创建取消上下文
	ctx, cancel := context.WithCancel(context.Background())
	// 创建存储错误的 channel
	errChan := make(chan error)
	// 遍历每一个卫星
	for _, satellite := range satellites {
		// 为了避免都是绑定到同一个卫星，创建一个局部变量
		sat := satellite
		go func() {
			select {
			case <-ctx.Done():
				return
			default:
				err := CreateContainer(sat)
				if err != nil {
					errChan <- err
					cancel()
					return
				}
				err = StartContainer(sat)
				if err != nil {
					errChan <- err
					cancel()
					return
				}
				// 如果都没有错误那么才会进行任务完成 + 1 以及进度条 + 1
				waitGroup.Done()
				progress_bar.Add(progressBar, 1)
			}
		}()
	}

	// 开启一个协程, 等待 waitGroup 的完成, 一旦完成之后关闭 errChan
	go func() {
		waitGroup.Wait()
		close(errChan)
	}()

	if err, ok := <-errChan; ok { // 如果存在错误
		fmt.Println("cancel here")
		cancel()
		return err
	} else { // 如果不存在错误
		cancel()
		if progressBar.IsFinished() {
			fmt.Println()
		}
		return nil
	}
}

func CreateContainer(sat Satellite) error {
	if sat.name == "zzg" {
		return fmt.Errorf("create satellite error")
	} else {
		return nil
	}
}

func StartContainer(sat Satellite) error {
	return fmt.Errorf("start satellite error")
}
