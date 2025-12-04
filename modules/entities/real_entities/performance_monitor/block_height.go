package performance_monitor

import (
	"math"
)

// BlockHeight 更新区块链高度信息
func BlockHeight(pm *PerformanceMonitor) {
	// 计算当前节点的区块高度占整个区块高度的百分比
	var myHeight float64
	var maxHeight float64 = math.MinInt
	var otherHeight float64
	var blockHeightPercentage float64
	var blockHeightDifference float64

	// 拿到当前和最高区块高度
	// ----------------------------------------------------------------------
	MutexForHeightRecorder.Lock()
	blockHeightRecorder := BlockHeightRecorder
	for containerName, heightList := range blockHeightRecorder {
		if len(heightList) == 0 {
			if pm.NormalNode.ContainerName == containerName {
				myHeight = 0
				maxHeight = math.Max(maxHeight, myHeight)
			} else {
				otherHeight = 0
				maxHeight = math.Max(maxHeight, otherHeight)
			}
		} else {
			if pm.NormalNode.ContainerName == containerName {
				myHeight = float64(heightList[len(heightList)-1].BlockHeight)
				maxHeight = math.Max(maxHeight, myHeight)
			} else {
				otherHeight = float64(heightList[len(heightList)-1].BlockHeight)
				maxHeight = math.Max(maxHeight, otherHeight)
			}
		}
	}
	MutexForHeightRecorder.Unlock()
	// ----------------------------------------------------------------------

	// 进行存储
	// ----------------------------------------------------------------------
	if maxHeight == 0 {
		blockHeightPercentage = 100
		blockHeightDifference = 0
	} else {
		blockHeightPercentage = myHeight / maxHeight * 100
		blockHeightDifference = maxHeight - myHeight
	}
	if len(pm.TimeList) == pm.FixedLength {
		pm.BlockHeightPercentageList = pm.BlockHeightPercentageList[1:]
		pm.BlockHeightPercentageList = append(pm.BlockHeightPercentageList, blockHeightPercentage)
	} else {
		pm.BlockHeightPercentageList = append(pm.BlockHeightPercentageList, blockHeightPercentage)
	}
	pm.BlockHeightPercentageListAll = append(pm.BlockHeightPercentageListAll, blockHeightPercentage)
	pm.BlockHeightDifferentListAll = append(pm.BlockHeightDifferentListAll, blockHeightDifference)
	pm.BlockHeightListAll = append(pm.BlockHeightListAll, myHeight)
	// ----------------------------------------------------------------------
}
