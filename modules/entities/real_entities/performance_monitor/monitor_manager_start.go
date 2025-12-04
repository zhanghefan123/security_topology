package performance_monitor

import (
	"fmt"
	clientv3 "go.etcd.io/etcd/client/v3"
	"math"
	"strconv"
	"strings"
	"time"
	"zhanghefan123/security_topology/api/chain_api"
	"zhanghefan123/security_topology/modules/entities/real_entities/topology"
)

const (
	StartAllPerformanceMonitor = "StartAllPerformanceMonitor"
	StartEtcdWatch             = "StartEtcdWatch"
	StartCaculateMaxHeight     = "StartMaxCalculateHeight"
)

type StartFunction func() error

type StartModule struct {
	start         bool          // 是否启动
	startFunction StartFunction // 相应的启动函数
}

// Start 启动
func (mm *MonitorManager) Start() error {
	startSteps := []map[string]StartModule{ // step7 进行默认路由的添加
		{StartAllPerformanceMonitor: StartModule{topology.Instance.FabricEnabled || topology.Instance.ChainMakerEnabled || topology.Instance.FiscoBcosEnabled, mm.StartAllPerformanceMonitor}}, // step8 进行性能的监听
		{StartEtcdWatch: StartModule{true, mm.StartEtcdWatch}},
		{StartCaculateMaxHeight: StartModule{true, mm.StartCalculateMaxHeight}},
	}
	err := mm.startSteps(startSteps)
	if err != nil {
		return fmt.Errorf("constellation start error: %w", err)
	}
	return nil
}

// startStepsNum 获取启动的模块的数量
func (mm *MonitorManager) startStepsNum(startSteps []map[string]StartModule) int {
	result := 0
	for _, startStep := range startSteps {
		for _, startModule := range startStep {
			if startModule.start {
				result += 1
			}
		}
	}
	return result
}

// startSteps 调用所有的启动方法
func (mm *MonitorManager) startSteps(startSteps []map[string]StartModule) (err error) {
	moduleNum := mm.startStepsNum(startSteps)
	for idx, startStep := range startSteps {
		for name, startModule := range startStep {
			// 判断是否需要进行启动, 如果要进行启动，再调用
			if startModule.start {
				if err = startModule.startFunction(); err != nil {
					monitorManagerLogger.Errorf("start step [%s] failed, %s", name, err)
					return err
				}
				monitorManagerLogger.Infof("BASE START STEP (%d/%d) => start step [%s] success)", idx+1, moduleNum, name)
			}
		}
	}
	return
}

// StartAllPerformanceMonitor 启动所有的性能监测设备

func (mm *MonitorManager) StartAllPerformanceMonitor() error {
	if _, ok := mm.MonitorManagerStartSteps[StartAllPerformanceMonitor]; ok {
		monitorManagerLogger.Infof("Already start all performance monitor")
		return nil
	}

	// 启动所有的监测器
	for _, abstractNode := range topology.Instance.AllChainAbstractNodes {
		// 获取所有的 chainMakerContainer 的 name
		performanceMonitor, err := NewInstancePerformanceMonitor(abstractNode,
			topology.Instance.TopologyParams.BlockChainType,
			topology.Instance.GetChainMakerNodeContainerNames(),
			topology.Instance.GetFabricNodeContainerNames(),
			topology.Instance.GetFiscoBcosContainerNames())
		if err != nil {
			return fmt.Errorf("get performance monitor error")
		}
		KeepGettingPerformance(performanceMonitor)
	}

	// 启动 global tx rate
	chain_api.StartGlobalTxRateRecorder(topology.Instance.TopologyParams.BlockChainType)

	// 进行全局的 ticker 启动
	StartTicker()

	mm.MonitorManagerStartSteps[StartAllPerformanceMonitor] = struct{}{}
	return nil
}

func (mm *MonitorManager) StartEtcdWatch() error {
	if _, ok := mm.MonitorManagerStartSteps[StartEtcdWatch]; ok {
		monitorManagerLogger.Infof("Already start etcd watch")
		return nil
	}

	// 进行每个节点的区块高都的监视
	err := mm.WatchBlockHeight()
	if err != nil {
		return fmt.Errorf("watch block height error")
	}

	// 进行每个节点的 tcp 连接的监视
	err = mm.WatchTcpConnectedAndHalfConnected()
	if err != nil {
		return fmt.Errorf("watch tcp connection error")
	}

	// 进行每个节点超时次数的监视
	err = mm.WatchTimeout()
	if err != nil {
		return fmt.Errorf("watch timeout error")
	}

	// 进行实时的黑名单节点数量的监视
	err = mm.WatchBlackListCount()
	if err != nil {
		return fmt.Errorf("watch blacklist error")
	}

	mm.MonitorManagerStartSteps[StartEtcdWatch] = struct{}{}
	return nil
}

func (mm *MonitorManager) WatchTimeout() error {
	for _, abstractNode := range topology.Instance.AllChainAbstractNodes {
		normalNode, err := abstractNode.GetNormalNodeFromAbstractNode()
		if err != nil {
			return fmt.Errorf("get normal node from abstract node error: %v", err)
		}
		timeoutPrefix := fmt.Sprintf("/chain/%s/timeout", normalNode.ContainerName)
		mm.EtcdWatchWaitGroup.Add(1)
		go func() {
			defer func() {
				mm.EtcdWatchWaitGroup.Done()
			}()
			// 创建一个监听键值对更新事件的 channel
			watchChan := mm.EtcdClient.Watch(
				mm.ServiceContext,
				timeoutPrefix,
				clientv3.WithPrefix(),
			)
			for response := range watchChan {
				for _, event := range response.Events {
					result := string(event.Kv.Value)
					timeStampAndTimeout := strings.Split(result, ",")
					var timeInMs int64
					var timeoutCount int64
					timeInMs, err = strconv.ParseInt(timeStampAndTimeout[0], 10, 64)
					if err != nil {
						fmt.Printf("error parse int: %v", err)
					}
					// 解析毫秒时间戳
					timeStamp := time.Unix(0, timeInMs*int64(time.Millisecond))
					timeoutCount, err = strconv.ParseInt(timeStampAndTimeout[1], 10, 64)
					if err != nil {
						fmt.Printf("error parse int: %v", err)
					}
					// fmt.Printf("time: %v, timeout: %v\n", timeStamp, timeoutCount)
					// 存储
					timeoutInfo := &TimeoutInfo{
						Timestamp:    timeStamp,
						TimeoutCount: timeoutCount,
					}
					// 将其放到队列之中进行记录
					MutexForTimeoutRecorder.Lock()
					TimeoutRecorder[normalNode.ContainerName] = append(TimeoutRecorder[normalNode.ContainerName],
						timeoutInfo)
					MutexForTimeoutRecorder.Unlock()
				}
			}
		}()
	}
	return nil
}

func (mm *MonitorManager) WatchBlackListCount() error {
	for _, abstractNode := range topology.Instance.AllChainAbstractNodes {
		normalNode, err := abstractNode.GetNormalNodeFromAbstractNode()
		if err != nil {
			return fmt.Errorf("get normal node from abstract node error: %v", err)
		}
		blacklistPrefix := fmt.Sprintf("/chain/%s/blacklist", normalNode.ContainerName)
		mm.EtcdWatchWaitGroup.Add(1)
		go func() {
			defer func() {
				mm.EtcdWatchWaitGroup.Done()
			}()
			watchChan := mm.EtcdClient.Watch(
				mm.ServiceContext,
				blacklistPrefix,
				clientv3.WithPrefix())
			for response := range watchChan {
				for _, event := range response.Events {
					result := string(event.Kv.Value)
					timeStampAndBlackListCount := strings.Split(result, ",")
					var timeInMs int64
					var blackListCount int64
					timeInMs, err = strconv.ParseInt(timeStampAndBlackListCount[0], 10, 64)
					if err != nil {
						fmt.Printf("error parse int: %v", err)
					}
					// 解析毫秒时间戳
					timeStamp := time.Unix(0, timeInMs*int64(time.Millisecond))
					blackListCount, err = strconv.ParseInt(timeStampAndBlackListCount[1], 10, 64)
					if err != nil {
						fmt.Printf("error parse int: %v", err)
					}
					fmt.Printf("time: %v, blackListCount: %v\n", timeStamp, blackListCount)
					// 存储
					blackListCountInfo := &BlackListCountInfo{
						Timestamp:      timeStamp,
						BlackListCount: blackListCount,
					}
					// 将其放到队列之中进行记录
					MutexForBlackListCountRecorder.Lock()
					BlackListCountRecorder[normalNode.ContainerName] = append(BlackListCountRecorder[normalNode.ContainerName],
						blackListCountInfo)
					MutexForBlackListCountRecorder.Unlock()
				}
			}
		}()
	}
	return nil
}

func (mm *MonitorManager) WatchTcpConnectedAndHalfConnected() error {
	for _, abstractNode := range topology.Instance.AllChainAbstractNodes {
		normalNode, err := abstractNode.GetNormalNodeFromAbstractNode()
		if err != nil {
			return fmt.Errorf("get normal node from abstract node error: %v", err)
		}
		tcpPrefix := fmt.Sprintf("/chain/%s/tcp", normalNode.ContainerName)
		mm.EtcdWatchWaitGroup.Add(1)
		go func() {
			defer func() {
				mm.EtcdWatchWaitGroup.Done()
			}()
			// 创建一个监听键值对更新事件的 channel
			watchChan := mm.EtcdClient.Watch(
				mm.ServiceContext,
				tcpPrefix,
				clientv3.WithPrefix(),
			)
			for response := range watchChan {
				for _, event := range response.Events {
					result := string(event.Kv.Value)
					tcpConnectedAndHalfConnectedString := strings.Split(result, ",")
					var tcpConnected int64
					var tcpHalfConnected int64
					tcpConnected, err = strconv.ParseInt(tcpConnectedAndHalfConnectedString[0], 10, 64)
					if err != nil {
						fmt.Printf("error parse int: %v", err)
					}
					tcpHalfConnected, err = strconv.ParseInt(tcpConnectedAndHalfConnectedString[1], 10, 64)
					if err != nil {
						fmt.Printf("error parse int: %v", err)
					}
					tcpConnectedAndHalfConnected := &TcpConnectedAndHalfConnectedInfo{
						TcpConnected:     tcpConnected,
						TcpHalfConnected: tcpHalfConnected,
					}
					//fmt.Printf("tcp connected: %v, tcp half connected: %v\n", tcpConnected, tcpHalfConnected)

					// 将其放到队列之中进行记录
					MutexForTcpRecorder.Lock()
					TcpConnectedAndHalfConnectedRecorder[normalNode.ContainerName] = append(TcpConnectedAndHalfConnectedRecorder[normalNode.ContainerName],
						tcpConnectedAndHalfConnected)
					MutexForTcpRecorder.Unlock()
				}
			}
		}()
	}
	return nil
}

func (mm *MonitorManager) WatchBlockHeight() error {
	for _, abstractNode := range topology.Instance.AllChainAbstractNodes {
		normalNode, err := abstractNode.GetNormalNodeFromAbstractNode()
		if err != nil {
			return fmt.Errorf("get normal node from abstract node error: %v", err)
		}
		blockHeightPrefix := fmt.Sprintf("/chain/%s/height", normalNode.ContainerName)
		mm.EtcdWatchWaitGroup.Add(1)
		go func() {
			var previousTimeStamp *time.Time = nil
			defer func() {
				mm.EtcdWatchWaitGroup.Done()
			}()
			// 创建一个监听键值对更新事件的 channel
			watchChan := mm.EtcdClient.Watch(
				mm.ServiceContext,
				blockHeightPrefix,
				clientv3.WithPrefix(),
			)
			for response := range watchChan {
				for _, event := range response.Events {
					result := string(event.Kv.Value)
					timeStampAndHeight := strings.Split(result, ",")
					var timeInMs int64
					var height int64
					timeInMs, err = strconv.ParseInt(timeStampAndHeight[0], 10, 64)
					if err != nil {
						fmt.Printf("error parse int: %v", err)
					}
					// 解析毫秒时间戳
					timeStamp := time.Unix(0, timeInMs*int64(time.Millisecond))
					height, err = strconv.ParseInt(timeStampAndHeight[1], 10, 64)
					if err != nil {
						fmt.Printf("error parse int: %v", err)
					}

					//fmt.Printf("time: %v, current height: %v\n", time.UnixMilli(timeInMs), height)

					if previousTimeStamp == nil {
						previousTimeStamp = &timeStamp
					} else {
						if timeStamp.Before(*previousTimeStamp) {
							fmt.Printf("received previous timestamp")
						} else {
							previousTimeStamp = &timeStamp
						}
					}
					// 进行打印
					// fmt.Printf("timestamp: %v, height: %v\n", timeStamp, height)
					blockHeightInfo := &BlockHeightInfo{
						TimeStamp:   timeStamp,
						BlockHeight: height,
					}
					// 将其放到队列之中进行记录
					MutexForHeightRecorder.Lock()
					BlockHeightRecorder[normalNode.ContainerName] = append(BlockHeightRecorder[normalNode.ContainerName],
						blockHeightInfo)
					MutexForHeightRecorder.Unlock()
				}
			}
		}()
	}
	return nil
}

func (mm *MonitorManager) StartCalculateMaxHeight() error {
	if _, ok := mm.MonitorManagerStartSteps[StartCaculateMaxHeight]; ok {
		monitorManagerLogger.Infof("Already start calculate max height")
		return nil
	}
	mm.WaitGroup.Add(1)
	go func() {
		defer mm.WaitGroup.Done()
	ForLoop:
		for {
			select {
			case <-mm.StopChannel:
				break ForLoop
			case <-mm.TimerChannel:
				var otherHeight, maxHeight float64
				MutexForHeightRecorder.Lock()
				blockHeightRecorder := BlockHeightRecorder
				// 找到最大高度, 然后进行记录
				for _, heightList := range blockHeightRecorder {
					if len(heightList) == 0 {
						otherHeight = 0
						maxHeight = math.Max(maxHeight, otherHeight)
					} else {
						otherHeight = float64(heightList[len(heightList)-1].BlockHeight)
						maxHeight = math.Max(maxHeight, otherHeight)
					}
				}
				mm.MaxBlockHeightListAll = append(mm.MaxBlockHeightListAll, maxHeight)
				MutexForHeightRecorder.Unlock()
			}
		}
	}()

	mm.MonitorManagerStartSteps[StartCaculateMaxHeight] = struct{}{}
	return nil
}
