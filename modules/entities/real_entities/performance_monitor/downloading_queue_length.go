package performance_monitor

func DownloadingQueueLength(pm *PerformanceMonitor) {
	pm.DownloadingQueueLengthListAll = append(pm.DownloadingQueueLengthListAll, 0)
}
