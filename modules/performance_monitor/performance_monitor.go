package performance_monitor

import (
	"fmt"
	"time"
	"zhanghefan123/security_topology/modules/entities/abstract_entities/node"
	"zhanghefan123/security_topology/modules/entities/real_entities/normal_node"
	"zhanghefan123/security_topology/modules/entities/types"
)

var (
	PerformanceMonitorMapping = map[string]*PerformanceMonitor{}
)

// PerformanceMonitor 性能检测者
type PerformanceMonitor struct {
	AbstractNode                *node.AbstractNode      // 监听的抽象节点
	NormalNode                  *normal_node.NormalNode // 监听的普通节点
	AllChainMakerContainerNames []string                // 所有长安链节点的名称

	TimeList []int // 时间列表

	InterfaceRateList []float64 // 接口速率监控
	LastReceivedBytes int       // 上一次收到的字节数

	CpuRatioList []float64 // cpu 利用率
	LastCpuBusy  float64   // 上一次的 cpu 繁忙时间

	MemoryMBList []float64 // 内存占用

	BlockHeightPercentageList []float64 // 区块高度占比

	FixedLength  int           // 队列的长度
	StopChannel  chan struct{} // 停止channel
	CurrentCount int           // 当前的数量
}

// NewInstancePerformanceMonitor 创建新的接口监听器
func NewInstancePerformanceMonitor(abstractNode *node.AbstractNode, allChainMakerContainerNames []string) (*PerformanceMonitor, error) {
	normalNode, err := abstractNode.GetNormalNodeFromAbstractNode()
	if err != nil {
		return nil, fmt.Errorf("GetNormalNodeFromAbstractNode failed: %w", err)
	}
	performanceMonitor := &PerformanceMonitor{
		AbstractNode:                abstractNode,
		NormalNode:                  normalNode,
		AllChainMakerContainerNames: allChainMakerContainerNames,
		TimeList:                    make([]int, 0),

		InterfaceRateList: make([]float64, 0),
		LastReceivedBytes: 0,

		CpuRatioList: make([]float64, 0),
		LastCpuBusy:  0,

		MemoryMBList: make([]float64, 0),

		BlockHeightPercentageList: make([]float64, 0),

		FixedLength:  10,
		StopChannel:  nil, // 在启动之后会进行赋值
		CurrentCount: 0,
	}
	PerformanceMonitorMapping[normalNode.ContainerName] = performanceMonitor
	return performanceMonitor, nil
}

// RemovePerformanceMonitor 移除性能监听器
func RemovePerformanceMonitor(abstractNode *node.AbstractNode) error {
	normalNode, err := abstractNode.GetNormalNodeFromAbstractNode()
	if err != nil {
		return fmt.Errorf("GetNormalNodeFromAbstractNode failed: %w", err)
	}
	PerformanceMonitorMapping[normalNode.ContainerName].StopChannel <- struct{}{}
	delete(PerformanceMonitorMapping, normalNode.ContainerName)
	return nil
}

// Inter-|   Receive                                                |  Transmit
// face |bytes    packets errs drop fifo frame compressed multicast|bytes    packets errs drop fifo colls carrier compressed
//    lo:       0       0    0    0    0     0          0         0        0       0    0    0    0     0       0          0
//r1_idx1:    3428      34    0    0    0     0          0         0     3584      34    0    0    0     0       0          0
//r1_idx2:    3562      35    0    0    0     0          0         0     3680      34    0    0    0     0       0          0
//  eth0:    7122      73    0    0    0     0          0         0     1026      11    0    0    0     0       0          0

func (pm *PerformanceMonitor) KeepGettingPerformance() {
	stopChannel := make(chan struct{})
	pm.StopChannel = stopChannel
	go func() {
	InternalLoop:
		for {
			select {
			case <-stopChannel:
				break InternalLoop
			default:
				// 进行接口速率的更新
				err := pm.UpdateInterfaceRate()
				if err != nil {
					fmt.Printf("UpdateInterfaceRate failed: %v\n", err)
					break InternalLoop
				}
				// 进行 cpu 和 内存占用的更新
				err = pm.UpdateInstanceCpuAndMemoryRatio()
				if err != nil {
					fmt.Printf("UpdateInstanceCpuAndMemoryRatio failed: %v\n", err)
					break InternalLoop
				}
				// 更新区块链的高度信息
				// 判断节点类型 -> 只有共识节点才需要进行写入
				if pm.NormalNode.Type == types.NetworkNodeType_ChainMakerNode {
					err = pm.UpdateBlockHeightInfo()
					if err != nil {
						fmt.Printf("UpdateBlockHeightInfo failed: %v\n", err)
						break InternalLoop
					}
				}
				// 进行队列长度的控制
				if len(pm.TimeList) == pm.FixedLength {
					pm.TimeList = pm.TimeList[1:]
					pm.TimeList = append(pm.TimeList, pm.CurrentCount)
				} else {
					pm.TimeList = append(pm.TimeList, pm.CurrentCount)
				}
				pm.CurrentCount += 1
				time.Sleep(time.Second)
			}
		}
	}()
}
