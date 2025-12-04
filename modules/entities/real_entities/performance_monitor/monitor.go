package performance_monitor

import (
	"fmt"
	"sync"
	"zhanghefan123/security_topology/modules/entities/abstract_entities/node"
	"zhanghefan123/security_topology/modules/entities/real_entities/normal_node"
	"zhanghefan123/security_topology/modules/entities/types"
)

// PerformanceMonitor 性能检测者
type PerformanceMonitor struct {
	ChainType types.ChainType // 链类型

	AbstractNode *node.AbstractNode      // 监听的抽象节点
	NormalNode   *normal_node.NormalNode // 监听的普通节点
	TimeList     []int                   // 时间列表

	InterfaceRateList []float64 // 接口速率监控
	LastReceivedBytes int       // 上一次收到的字节数

	CpuRatioList []float64 // cpu 利用率
	LastCpuBusy  float64   // 上一次的 cpu 繁忙时间
	MemoryMBList []float64 // 内存占用

	RequestTimeoutList        []int     // 请求超时列表
	MessageCountList          []int     // 消息总线消息数量
	ConnectedCountList        []int     // 已经建立的 tcp 连接的数量
	HalfConnectedCountList    []int     // 半开 tcp 连接的数量
	BlockHeightPercentageList []float64 // 区块高度占比
	BlackListCountList        []int     // 黑名单长度变化

	InterfaceRateListAll          []float64 // 接口速率总表
	BlockHeightPercentageListAll  []float64 // 区块高度占比总表
	BlockHeightDifferentListAll   []float64 // 和最高的区块高度的差异 (统计的粒度不够)
	BlockHeightListAll            []float64 // 区块的高度
	DownloadingQueueLengthListAll []int     // 下载队列长度 (fisco bcos 专属)
	RequestTimeoutListAll         []int     // 请求超时列表总量
	BlackListCountListAll         []int     // 黑名单长度变化总列表

	FixedLength  int            // 队列的长度
	StopChannel  chan struct{}  // 停止channel
	WaitGroup    sync.WaitGroup // 等待完全停止之后进行后续的步骤
	TimerChannel chan struct{}
	CurrentCount int // 当前的数量

	// 所有的链的名称
	AllChainMakerContainerNames  []string
	AllFabricOrderContainerNames []string
	AllFiscoBcosContainerNames   []string
	AllChainContainerNames       []string
}

// NewInstancePerformanceMonitor 创建新的接口监听器
func NewInstancePerformanceMonitor(abstractNode *node.AbstractNode, chainType types.ChainType,
	allChainMakerContainerNames []string, allFabricOrderContainerNames []string, allFiscoBcosContainerNames []string) (*PerformanceMonitor, error) {
	normalNode, err := abstractNode.GetNormalNodeFromAbstractNode()
	if err != nil {
		return nil, fmt.Errorf("GetNormalNodeFromAbstractNode failed: %w", err)
	}
	performanceMonitor := &PerformanceMonitor{
		ChainType: chainType,

		AbstractNode: abstractNode,
		NormalNode:   normalNode,
		TimeList:     make([]int, 0),

		InterfaceRateList: make([]float64, 0),
		LastReceivedBytes: 0,

		CpuRatioList: make([]float64, 0),
		LastCpuBusy:  0,
		MemoryMBList: make([]float64, 0),

		RequestTimeoutList:     make([]int, 0),
		MessageCountList:       make([]int, 0),
		ConnectedCountList:     make([]int, 0),
		HalfConnectedCountList: make([]int, 0),
		BlackListCountList:     make([]int, 0),

		BlockHeightPercentageList: make([]float64, 0),

		InterfaceRateListAll:          make([]float64, 0),
		BlockHeightPercentageListAll:  make([]float64, 0),
		BlockHeightDifferentListAll:   make([]float64, 0),
		BlockHeightListAll:            make([]float64, 0),
		DownloadingQueueLengthListAll: make([]int, 0),
		RequestTimeoutListAll:         make([]int, 0),
		BlackListCountListAll:         make([]int, 0),

		FixedLength:  10,
		StopChannel:  make(chan struct{}), // 在启动之后会进行赋值
		WaitGroup:    sync.WaitGroup{},
		TimerChannel: make(chan struct{}),
		CurrentCount: 0,

		AllChainMakerContainerNames:  allChainMakerContainerNames,
		AllFabricOrderContainerNames: allFabricOrderContainerNames,
		AllFiscoBcosContainerNames:   allFiscoBcosContainerNames,
		AllChainContainerNames:       make([]string, 0),
	}
	performanceMonitor.AllChainContainerNames = append(performanceMonitor.AllChainMakerContainerNames, allChainMakerContainerNames...)
	performanceMonitor.AllChainContainerNames = append(performanceMonitor.AllChainMakerContainerNames, allFabricOrderContainerNames...)
	performanceMonitor.AllChainContainerNames = append(performanceMonitor.AllChainContainerNames, allFiscoBcosContainerNames...)
	PerformanceMonitorMapping[normalNode.ContainerName] = performanceMonitor
	return performanceMonitor, nil
}
