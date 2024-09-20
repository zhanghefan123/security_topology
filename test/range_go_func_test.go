package test

import (
	"fmt"
	"sync"
	"testing"
)

func TestRangeGoFunc(t *testing.T) {
	satellites := []int{1, 2, 3}
	waitGroup := sync.WaitGroup{}
	waitGroup.Add(len(satellites))
	for _, satellite := range satellites {
		go func() {
			fmt.Println(satellite)
			waitGroup.Done()
		}()
	}
	waitGroup.Wait()
}
