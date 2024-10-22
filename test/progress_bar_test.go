package test

import (
	"fmt"
	"math/rand"
	"sync"
	"testing"
	"time"
	"zhanghefan123/security_topology/modules/utils/progress_bar"
)

func TestProgressBar(t *testing.T) {
	progressBar := progress_bar.NewProgressBar(100, "test")
	waitGroup := sync.WaitGroup{}
	waitGroup.Add(100)
	for i := 0; i < 100; i++ {
		go func() {
			time.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond)
			progress_bar.Add(progressBar, 1)
			waitGroup.Done()
		}()
	}
	waitGroup.Wait()
	fmt.Println("\ntest")
}
