package performance_monitor

func BlackListCount(pm *PerformanceMonitor) {
	var blackListCount int
	blackListCountRecorder := BlackListCountRecorder
	MutexForBlackListCountRecorder.Lock()
	countList := blackListCountRecorder[pm.NormalNode.ContainerName]
	MutexForBlackListCountRecorder.Unlock()
	if len(countList) == 0 {
		blackListCount = 0
	} else {
		blackListCount = int(countList[len(countList)-1].BlackListCount)
	}

	if len(pm.BlackListCountList) == pm.FixedLength {
		pm.BlackListCountList = pm.BlackListCountList[1:]
		pm.BlackListCountList = append(pm.BlackListCountList, blackListCount)
	} else {
		pm.BlackListCountList = append(pm.BlackListCountList, blackListCount)
	}

	pm.BlackListCountListAll = append(pm.BlackListCountListAll, blackListCount)
}
