package performance_monitor

import (
	"fmt"
	"os"
	"zhanghefan123/security_topology/cmd/variables"
	"zhanghefan123/security_topology/modules/entities/abstract_entities/node"
	"zhanghefan123/security_topology/modules/entities/types"
	"zhanghefan123/security_topology/utils/file"
	"zhanghefan123/security_topology/utils/judge"
)

// RemovePerformanceMonitor 移除性能监听器
func RemovePerformanceMonitor(abstractNode *node.AbstractNode) error {
	normalNode, err := abstractNode.GetNormalNodeFromAbstractNode()
	if err != nil {
		return fmt.Errorf("GetNormalNodeFromAbstractNode failed: %w", err)
	}
	PerformanceMonitorMapping[normalNode.ContainerName].StopChannel <- struct{}{}
	return nil
}

// GetPerformanceMonitor 获取某个 performance monitor
func GetPerformanceMonitor(abstractNode *node.AbstractNode) (*PerformanceMonitor, error) {
	normalNode, err := abstractNode.GetNormalNodeFromAbstractNode()
	if err != nil {
		return nil, fmt.Errorf("GetNormalNodeFromAbstractNode failed: %w", err)
	}
	return PerformanceMonitorMapping[normalNode.ContainerName], nil
}

// Inter-|   Receive                                                |  Transmit
// face |bytes    packets errs drop fifo frame compressed multicast|bytes    packets errs drop fifo colls carrier compressed
//    lo:       0       0    0    0    0     0          0         0        0       0    0    0    0     0       0          0
//r1_idx1:    3428      34    0    0    0     0          0         0     3584      34    0    0    0     0       0          0
//r1_idx2:    3562      35    0    0    0     0          0         0     3680      34    0    0    0     0       0          0
//  eth0:    7122      73    0    0    0     0          0         0     1026      11    0    0    0     0       0          0

func KeepGettingPerformance(pm *PerformanceMonitor) {
	pm.WaitGroup.Add(1)
	go func() {
		defer pm.WaitGroup.Done()
	ForLoop:
		for {
			select {
			case <-pm.StopChannel:
				break ForLoop
			case <-pm.TimerChannel:
				// 进行接口速率的更新
				err := InterfaceRate(pm)
				if err != nil {
					fmt.Printf("InterfaceRate failed: %v\n", err)
					break ForLoop
				}
				// 进行 cpu 和 内存占用的更新 (这个是通过读取文件内容知道的,不需要修改)
				err = InstanceCpuAndMemoryRatio(pm)
				if err != nil {
					fmt.Printf("InstanceCpuAndMemoryRatio failed: %v\n", err)
					break ForLoop
				}
				// 进行 tcp的更新
				TcpConnectedAndHalfConnected(pm)
				// 判断节点类型 -> 只有长安链/HyperledgerFabric/FiscoBcos 才需要进行写入
				if judge.IsBlockChainType(pm.NormalNode.Type) {
					if pm.ChainType == types.ChainType_FiscoBcos {
						DownloadingQueueLength(pm) // 判断是否是 fisco bcos 链 (如果是的话需要更新 DownloadingQueueLength)
					}
					BlockHeight(pm)
					TimeoutCount(pm)
					MessageCount(pm)
					BlackListCount(pm)
				}
				// 进行队列长度的控制
				if len(pm.TimeList) == pm.FixedLength {
					pm.TimeList = pm.TimeList[1:]
					pm.TimeList = append(pm.TimeList, pm.CurrentCount)
				} else {
					pm.TimeList = append(pm.TimeList, pm.CurrentCount)
				}
				pm.CurrentCount += 1
				// fmt.Printf("Performance Monitor for container %s Write Result\n", pm.NormalNode.ContainerName)
			}
		}
	}()
}

// WriteResultIntoFile 将结果写到文件之中
func WriteResultIntoFile(pm *PerformanceMonitor) error {
	directory := fmt.Sprintf("./result/result%d", variables.UserSelectedExperimentNumber)

	err := os.MkdirAll(fmt.Sprintf("%s/%s", directory, pm.NormalNode.ContainerName), os.ModePerm)
	if err != nil {
		return fmt.Errorf("mkdirall error: %v", err)
	}

	// 写入 interface_rate_list.txt
	finalString := ""
	for index := 0; index < len(pm.InterfaceRateListAll); index++ {
		if index == len(pm.InterfaceRateListAll)-1 {
			finalString += fmt.Sprintf("%f", pm.InterfaceRateListAll[index])
		} else {
			finalString += fmt.Sprintf("%f", pm.InterfaceRateListAll[index]) + ","
		}
	}
	err = file.WriteStringIntoFile(fmt.Sprintf("%s/%s/interface_rate_list.txt", directory, pm.NormalNode.ContainerName), finalString)
	if err != nil {
		return fmt.Errorf("write result into file failed: %v", err)
	}

	// 写入 block_height_percentage.txt
	finalString = ""
	for index := 0; index < len(pm.BlockHeightPercentageListAll); index++ {
		if index == len(pm.BlockHeightPercentageListAll)-1 {
			finalString += fmt.Sprintf("%f", pm.BlockHeightPercentageListAll[index])
		} else {
			finalString += fmt.Sprintf("%f", pm.BlockHeightPercentageListAll[index]) + ","
		}
	}
	err = file.WriteStringIntoFile(fmt.Sprintf("%s/%s/block_percentage_list.txt", directory, pm.NormalNode.ContainerName), finalString)
	if err != nil {
		return fmt.Errorf("write result into file failed: %v", err)
	}

	// 写入 block_height_difference.txt
	finalString = ""
	for index := 0; index < len(pm.BlockHeightDifferentListAll); index++ {
		if index == len(pm.BlockHeightDifferentListAll)-1 {
			finalString += fmt.Sprintf("%f", pm.BlockHeightDifferentListAll[index])
		} else {
			finalString += fmt.Sprintf("%f", pm.BlockHeightDifferentListAll[index]) + ","
		}
	}
	err = file.WriteStringIntoFile(fmt.Sprintf("%s/%s/block_difference_list.txt", directory, pm.NormalNode.ContainerName), finalString)
	if err != nil {
		return fmt.Errorf("write result into file failed: %v", err)
	}

	// 写入 block_height_list_all.txt
	finalString = ""
	for index := 0; index < len(pm.BlockHeightListAll); index++ {
		if index == len(pm.BlockHeightListAll)-1 {
			finalString += fmt.Sprintf("%f", pm.BlockHeightListAll[index])
		} else {
			finalString += fmt.Sprintf("%f", pm.BlockHeightListAll[index]) + ","
		}
	}
	err = file.WriteStringIntoFile(fmt.Sprintf("%s/%s/block_height_list.txt", directory, pm.NormalNode.ContainerName), finalString)
	if err != nil {
		return fmt.Errorf("write result into file failed: %v", err)
	}

	// 写入 downloading_queue.txt
	if pm.ChainType == types.ChainType_FiscoBcos {
		finalString = ""
		for index := 0; index < len(pm.DownloadingQueueLengthListAll); index++ {
			if index == len(pm.DownloadingQueueLengthListAll)-1 {
				finalString += fmt.Sprintf("%d", pm.DownloadingQueueLengthListAll[index])
			} else {
				finalString += fmt.Sprintf("%d", pm.DownloadingQueueLengthListAll[index]) + ","
			}
		}
		err = file.WriteStringIntoFile(fmt.Sprintf("%s/%s/downloading_queue_length.txt", directory, pm.NormalNode.ContainerName), finalString)
		if err != nil {
			return fmt.Errorf("write result into file failed: %v", err)
		}
	}

	// 写入 timeout_count.txt
	finalString = ""
	for index := 0; index < len(pm.RequestTimeoutListAll); index++ {
		if index == len(pm.RequestTimeoutListAll)-1 {
			finalString += fmt.Sprintf("%d", pm.RequestTimeoutListAll[index])
		} else {
			finalString += fmt.Sprintf("%d", pm.RequestTimeoutListAll[index]) + ","
		}
	}
	err = file.WriteStringIntoFile(fmt.Sprintf("%s/%s/request_timeout.txt", directory, pm.NormalNode.ContainerName), finalString)
	if err != nil {
		return fmt.Errorf("write result into file failed: %v", err)
	}

	// 写入 blacklist_count.txt
	finalString = ""
	for index := 0; index < len(pm.BlackListCountListAll); index++ {
		if index == len(pm.BlackListCountListAll)-1 {
			finalString += fmt.Sprintf("%d", pm.BlackListCountListAll[index])
		} else {
			finalString += fmt.Sprintf("%d", pm.BlackListCountListAll[index]) + ","
		}
	}
	err = file.WriteStringIntoFile(fmt.Sprintf("%s/%s/black_list_count.txt", directory, pm.NormalNode.ContainerName), finalString)
	if err != nil {
		return fmt.Errorf("write result into file failed: %v", err)
	}

	return nil
}

// RemoveAllPerformanceMonitors 删除所有的 mapping 之中的 performance_monitor
func RemoveAllPerformanceMonitors() error {
	for _, performanceMonitor := range PerformanceMonitorMapping {
		err := WriteResultIntoFile(performanceMonitor)
		if err != nil {
			return fmt.Errorf("write result into file error %v", err)
		}
		err = RemovePerformanceMonitor(performanceMonitor.AbstractNode)
		if err != nil {
			return fmt.Errorf("remove performance monitor error")
		}
	}
	return nil
}
