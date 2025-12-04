package performance_monitor

import (
	"context"
	clientv3 "go.etcd.io/etcd/client/v3"
	"sync"
	"zhanghefan123/security_topology/modules/logger"
)

var (
	Instance             *MonitorManager
	monitorManagerLogger = logger.GetLogger(logger.ModuleMonitorManager)
)

type MonitorManager struct {
	EtcdClient               *clientv3.Client
	ServiceContext           context.Context    // 服务上下文
	serviceContextCancelFunc context.CancelFunc // 服务上下文的取消函数
	EtcdWatchWaitGroup       sync.WaitGroup     // 等待队列

	// 进行节点的状态的存储
	MaxBlockHeightListAll []float64 // 最高区块高度列表

	MonitorManagerInitSteps  map[string]struct{}
	MonitorManagerStartSteps map[string]struct{}
	MonitorManagerStopSteps  map[string]struct{}

	TimerChannel chan struct{}
	StopChannel  chan struct{}
	WaitGroup    sync.WaitGroup
}

func NewMonitorManager(etcdClient *clientv3.Client) *MonitorManager {
	return &MonitorManager{
		EtcdClient:               etcdClient,
		EtcdWatchWaitGroup:       sync.WaitGroup{},
		MonitorManagerInitSteps:  make(map[string]struct{}),
		MonitorManagerStartSteps: make(map[string]struct{}),
		MonitorManagerStopSteps:  make(map[string]struct{}),

		MaxBlockHeightListAll: make([]float64, 0), // 最高区块高度列表
		TimerChannel:          make(chan struct{}),
		StopChannel:           make(chan struct{}),
		WaitGroup:             sync.WaitGroup{},
	}
}
