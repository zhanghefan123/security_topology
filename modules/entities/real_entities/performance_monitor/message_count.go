package performance_monitor

// MessageCount 更新消息总线之中的消息数量
func MessageCount(pm *PerformanceMonitor) {
	if len(pm.MessageCountList) == pm.FixedLength {
		pm.MessageCountList = pm.MessageCountList[1:]
		pm.MessageCountList = append(pm.MessageCountList, 0)
	} else {
		pm.MessageCountList = append(pm.MessageCountList, 0)
	}
}
