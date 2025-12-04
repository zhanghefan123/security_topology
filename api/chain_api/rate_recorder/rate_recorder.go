package rate_recorder

import (
	"sync"
	"sync/atomic"
)

var TxRateRecorderInstance *TxRateRecorder = nil

type TxRateRecorder struct {
	TimeList      []int
	RateList      []float64
	FixedLength   int
	StopQueue     chan struct{}
	WaitGroup     *sync.WaitGroup
	TickerChannel chan struct{}
	WorkerPool    chan struct{}
	TxCount       atomic.Uint64
	TimeCount     int
}

// NewTxRateRecorder 创建新的 TxRateRecorder
func NewTxRateRecorder() *TxRateRecorder {
	return &TxRateRecorder{
		TimeList:      make([]int, 0),
		RateList:      make([]float64, 0),
		FixedLength:   100,
		StopQueue:     make(chan struct{}),
		WaitGroup:     &sync.WaitGroup{},
		TickerChannel: make(chan struct{}),
		WorkerPool:    make(chan struct{}, 100),
		TxCount:       atomic.Uint64{},
		TimeCount:     1,
	}
}
