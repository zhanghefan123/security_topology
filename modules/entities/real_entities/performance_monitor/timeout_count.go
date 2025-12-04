package performance_monitor

// TimeoutCount 更新超时信息
func TimeoutCount(pm *PerformanceMonitor) {
	var timeoutCount int
	timeoutRecorder := TimeoutRecorder
	MutexForTimeoutRecorder.Lock()
	timeoutList := timeoutRecorder[pm.NormalNode.ContainerName]
	MutexForTimeoutRecorder.Unlock()
	if len(timeoutList) == 0 {
		timeoutCount = 0
	} else {
		timeoutCount = int(timeoutList[len(timeoutList)-1].TimeoutCount)
	}

	if len(pm.RequestTimeoutList) == pm.FixedLength {
		pm.RequestTimeoutList = pm.RequestTimeoutList[1:]
		pm.RequestTimeoutList = append(pm.RequestTimeoutList, timeoutCount)
	} else {
		pm.RequestTimeoutList = append(pm.RequestTimeoutList, timeoutCount)
	}

	pm.RequestTimeoutListAll = append(pm.RequestTimeoutListAll, timeoutCount)
}
