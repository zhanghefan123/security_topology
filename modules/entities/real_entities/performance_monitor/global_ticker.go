package performance_monitor

import (
	"sync"
	"time"
	"zhanghefan123/security_topology/api/chain_api"
)

var GlobalTickerInstance *GlobalTicker

type GlobalTicker struct {
	Ticker    *time.Ticker
	StopQueue chan struct{}
	WaitGroup sync.WaitGroup
}

func NewGlobalTicker() *GlobalTicker {
	return &GlobalTicker{
		Ticker:    time.NewTicker(time.Second),
		StopQueue: make(chan struct{}),
		WaitGroup: sync.WaitGroup{},
	}
}

func StartTicker() {
	GlobalTickerInstance = NewGlobalTicker()
	GlobalTickerInstance.WaitGroup.Add(1)
	go func() {
		defer GlobalTickerInstance.WaitGroup.Done()
	ForLoop:
		for {
			select {
			case <-GlobalTickerInstance.StopQueue:
				break ForLoop
			case <-GlobalTickerInstance.Ticker.C:
				// 触发最大区块高度的计算
				Instance.TimerChannel <- struct{}{}

				// 触发所有的 performance write
				for _, performanceMonitor := range PerformanceMonitorMapping {
					performanceMonitor.TimerChannel <- struct{}{}
				}

				// 触发 tps rate write
				if chain_api.GloablTxRateRecorderInstance != nil {
					chain_api.GloablTxRateRecorderInstance.TickerChannel <- struct{}{}
				}
			}
		}
	}()
}

func StopTicker() {
	if GlobalTickerInstance != nil {
		GlobalTickerInstance.StopQueue <- struct{}{}
		GlobalTickerInstance.WaitGroup.Wait()
	}
}
