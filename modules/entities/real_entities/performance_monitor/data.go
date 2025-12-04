package performance_monitor

import (
	"sync"
	"time"
)

type BlockHeightInfo struct {
	TimeStamp   time.Time
	BlockHeight int64
}

type TcpConnectedAndHalfConnectedInfo struct {
	TcpConnected     int64
	TcpHalfConnected int64
}

type TimeoutInfo struct {
	Timestamp    time.Time
	TimeoutCount int64
}

type BlackListCountInfo struct {
	Timestamp      time.Time
	BlackListCount int64
}

var (
	PerformanceMonitorMapping = map[string]*PerformanceMonitor{}

	MutexForHeightRecorder = sync.Mutex{}
	BlockHeightRecorder    = make(map[string][]*BlockHeightInfo)

	MutexForTcpRecorder                  = sync.Mutex{}
	TcpConnectedAndHalfConnectedRecorder = make(map[string][]*TcpConnectedAndHalfConnectedInfo)

	MutexForTimeoutRecorder = sync.Mutex{}
	TimeoutRecorder         = make(map[string][]*TimeoutInfo)

	MutexForBlackListCountRecorder = sync.Mutex{}
	BlackListCountRecorder         = make(map[string][]*BlackListCountInfo)
)

func InitializeRecorderForContainer(containerName string) {
	BlockHeightRecorder[containerName] = make([]*BlockHeightInfo, 0)
	TcpConnectedAndHalfConnectedRecorder[containerName] = make([]*TcpConnectedAndHalfConnectedInfo, 0)
	TimeoutRecorder[containerName] = make([]*TimeoutInfo, 0)
	BlackListCountRecorder[containerName] = make([]*BlackListCountInfo, 0)
}
