package multithread

import (
	"context"
	"fmt"
	"sync"
	"zhanghefan123/security_topology/modules/utils/progress_bar"
)

// TaskFunc 由于不确定 satellite 的类型
type TaskFunc[T any] func(object T) error

// RunInMultiThread 多线程运行
// description -> 任务描述
// taskFunc -> 任务
// objects -> 任务参数
// T -> 泛型
func RunInMultiThread[T any](description string, taskFunc TaskFunc[T], objects []T) error {
	// 任务数量
	total := len(objects)
	// 创建 progressBar 描述
	progressBar := progress_bar.NewProgressBar(total, description)
	// 创建等待组
	waitGroup := sync.WaitGroup{}
	waitGroup.Add(total)
	// 创建取消上下文
	ctx, cancel := context.WithCancel(context.Background())
	errChan := make(chan error, total)
	// 遍历所有的任务参数
	for _, object := range objects {
		objectTemp := object
		go func() {
			defer waitGroup.Done()
			defer progress_bar.Add(progressBar, 1)

			// select 处理不同情况
			select {
			case <-ctx.Done():
				return
			default:
				if err := taskFunc(objectTemp); err != nil {
					errChan <- err
					cancel()
					return
				}
			}
		}()
	}

	// 等待所有任务完成并关闭 errChan
	go func() {
		waitGroup.Wait()
		close(errChan)
	}()

	// 检查是否有错误
	if err, ok := <-errChan; ok {
		cancel() // 取消所有正在运行的协程
		return err
	} else {
		cancel()
		// 如果没有错误
		if progressBar.IsFinished() {
			fmt.Println()
		}
		return nil
	}
}
