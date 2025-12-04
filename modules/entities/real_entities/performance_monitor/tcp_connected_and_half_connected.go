package performance_monitor

func TcpConnectedAndHalfConnected(pm *PerformanceMonitor) {
	var connectedTcpCount int64
	var halfConnectedTcpCount int64

	MutexForTcpRecorder.Lock()
	tcpRecorder := TcpConnectedAndHalfConnectedRecorder
	if _, ok := tcpRecorder[pm.NormalNode.ContainerName]; ok {
		resultList := tcpRecorder[pm.NormalNode.ContainerName]
		if len(resultList) == 0 {
			connectedTcpCount = 0
			halfConnectedTcpCount = 0
		} else {
			connectedTcpCount = resultList[len(resultList)-1].TcpConnected
			halfConnectedTcpCount = resultList[len(resultList)-1].TcpHalfConnected
		}
	}
	MutexForTcpRecorder.Unlock()

	// 更新已建立连接队列
	if len(pm.ConnectedCountList) == pm.FixedLength {
		pm.ConnectedCountList = pm.ConnectedCountList[1:]
		pm.ConnectedCountList = append(pm.ConnectedCountList, int(connectedTcpCount))
	} else {
		pm.ConnectedCountList = append(pm.ConnectedCountList, int(connectedTcpCount))
	}

	// 更新半开连接队列
	if len(pm.HalfConnectedCountList) == pm.FixedLength {
		pm.HalfConnectedCountList = pm.HalfConnectedCountList[1:]
		pm.HalfConnectedCountList = append(pm.HalfConnectedCountList, int(halfConnectedTcpCount))
	} else {
		pm.HalfConnectedCountList = append(pm.HalfConnectedCountList, int(halfConnectedTcpCount))
	}
}
