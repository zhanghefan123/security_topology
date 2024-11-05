package performance_monitor

import (
	"fmt"
	"time"
	"zhanghefan123/security_topology/modules/entities/abstract_entities/node"
	"zhanghefan123/security_topology/modules/entities/real_entities/normal_node"
)

var (
	PerformanceMonitorMapping = map[string]*PerformanceMonitor{}
)

type PerformanceMonitor struct {
	abstractNode      *node.AbstractNode
	normalNode        *normal_node.NormalNode
	TimeList          []int     // 时间列表
	InterfaceRateList []float64 // 接口速率监控
	LastReceivedBytes int       // 上一次收到的字节数
	CpuRatioList      []float64 // cpu 利用率
	LastCpuBusy       float64   // 上一次的 cpu 繁忙时间
	MemoryBytesList   []float64 // 内存比率
	LastMemoryBytes   float64
	fixedLength       int           // 队列的长度
	StopChannel       chan struct{} // 停止channel
	currentCount      int           // 当前的数量
}

// NewInterfaceRateMonitor 创建新的接口监听器
func NewInterfaceRateMonitor(abstractNode *node.AbstractNode) (*PerformanceMonitor, error) {
	normalNode, err := abstractNode.GetNormalNodeFromAbstractNode()
	if err != nil {
		return nil, fmt.Errorf("GetNormalNodeFromAbstractNode failed: %w", err)
	}
	performanceMonitor := &PerformanceMonitor{
		abstractNode:      abstractNode,
		normalNode:        normalNode,
		TimeList:          make([]int, 0),
		InterfaceRateList: make([]float64, 0),
		CpuRatioList:      make([]float64, 0),
		LastCpuBusy:       0,
		fixedLength:       10,
		LastReceivedBytes: 0,
		StopChannel:       nil, // 在启动之后会进行赋值
		currentCount:      0,
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

// CaptureInterfaceRate 进行接口速率的捕获
func (pm *PerformanceMonitor) CaptureInterfaceRate(abstractNode *node.AbstractNode) (err error) {
	normalNode, err := abstractNode.GetNormalNodeFromAbstractNode()
	if err != nil {
		return fmt.Errorf("GetNormalNodeFromAbstractNode: %w", err)
	}
	pm.StopChannel = pm.KeepGetingPerformance(normalNode)
	return nil
}

// Inter-|   Receive                                                |  Transmit
// face |bytes    packets errs drop fifo frame compressed multicast|bytes    packets errs drop fifo colls carrier compressed
//    lo:       0       0    0    0    0     0          0         0        0       0    0    0    0     0       0          0
//r1_idx1:    3428      34    0    0    0     0          0         0     3584      34    0    0    0     0       0          0
//r1_idx2:    3562      35    0    0    0     0          0         0     3680      34    0    0    0     0       0          0
//  eth0:    7122      73    0    0    0     0          0         0     1026      11    0    0    0     0       0          0

func (pm *PerformanceMonitor) KeepGetingPerformance(normalNode *normal_node.NormalNode) chan struct{} {
	stopChannel := make(chan struct{})
	go func() {
	InternalLoop:
		for {
			select {
			case <-stopChannel:
				break InternalLoop
			default:
				err := pm.UpdateInterfaceRate()
				if err != nil {
					fmt.Printf("UpdateInterfaceRate failed: %v\n", err)
					break InternalLoop
				}
				err = pm.UpdateInstanceCpuAndMemoryRatio()
				if err != nil {
					fmt.Printf("UpdateInstanceCpuAndMemoryRatio failed: %v\n", err)
					break InternalLoop
				}
				if len(pm.TimeList) == pm.fixedLength {
					pm.TimeList = pm.TimeList[1:]
					pm.TimeList = append(pm.TimeList, pm.currentCount)
				} else {
					pm.TimeList = append(pm.TimeList, pm.currentCount)
				}
				pm.currentCount += 1
				time.Sleep(time.Second)
			}
		}
	}()
	return stopChannel
}
