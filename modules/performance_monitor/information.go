package performance_monitor

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"zhanghefan123/security_topology/configs"
)

type Information struct {
	BlockHeight          int
	ConnectedTcpCount    int
	HalfConnetedTcpCount int
	TimeoutCount         int
	BusMessageCount      int
}

func (pm *PerformanceMonitor) UpdateInformation(information *Information) (err error) {
	// 更新已建立连接队列
	if len(pm.ConnectedCountList) == pm.FixedLength {
		pm.ConnectedCountList = pm.ConnectedCountList[1:]
		pm.ConnectedCountList = append(pm.ConnectedCountList, information.ConnectedTcpCount)
	} else {
		pm.ConnectedCountList = append(pm.ConnectedCountList, information.ConnectedTcpCount)
	}

	// 更新半开连接队列
	if len(pm.HalfConnectedCountList) == pm.FixedLength {
		pm.HalfConnectedCountList = pm.HalfConnectedCountList[1:]
		pm.HalfConnectedCountList = append(pm.HalfConnectedCountList, information.HalfConnetedTcpCount)
	} else {
		pm.HalfConnectedCountList = append(pm.HalfConnectedCountList, information.HalfConnetedTcpCount)
	}

	// 更新超时列表
	if len(pm.RequestTimeoutList) == pm.FixedLength {
		pm.RequestTimeoutList = pm.RequestTimeoutList[1:]
		pm.RequestTimeoutList = append(pm.RequestTimeoutList, information.TimeoutCount)
	} else {
		pm.RequestTimeoutList = append(pm.RequestTimeoutList, information.TimeoutCount)
	}

	// 更新消息总数列表
	if len(pm.MessageCountList) == pm.FixedLength {
		pm.MessageCountList = pm.MessageCountList[1:]
		pm.MessageCountList = append(pm.MessageCountList, information.BusMessageCount)
	} else {
		pm.MessageCountList = append(pm.MessageCountList, information.BusMessageCount)
	}

	return nil
}

func ReadInformation(tcpConnsFilePath string) (information *Information, err error) {
	// 读取文件
	content, err := os.ReadFile(tcpConnsFilePath) // Go 1.16+ 直接读取文件
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// 分割字符串
	// 分割字符串
	parts := strings.Split(strings.TrimSpace(string(content)), ",")
	if len(parts) < 4 {
		return nil, fmt.Errorf("invalid format: expected 4 values, got %d", len(parts))
	}

	// 统一转换数字
	vals := make([]int, 5)
	for i, s := range parts[:5] {
		vals[i], err = strconv.Atoi(strings.TrimSpace(s))
		if err != nil {
			return nil, fmt.Errorf("invalid value at position %d: %w", i, err)
		}
	}

	information = &Information{
		BlockHeight:          vals[0],
		ConnectedTcpCount:    vals[1],
		HalfConnetedTcpCount: vals[2],
		TimeoutCount:         vals[3],
		BusMessageCount:      vals[4],
	}

	return information, nil
}

// UpdateBlockHeightInfo 更新区块链高度信息
func (pm *PerformanceMonitor) UpdateBlockHeightInfo(currentBlockHeight int) (err error) {
	var largestBlockHeight int        // 最高的区块高度
	var blockHeightPercentage float64 // 当前节点的区块 / 最大的区块高度
	var otherBlockHeight int          // 其他节点的区块高度

	// 获取当前节点的区块高度
	// ----------------------------------------------------------------------
	if currentBlockHeight > largestBlockHeight {
		largestBlockHeight = currentBlockHeight
	}
	// ----------------------------------------------------------------------

	// 长安链 获取其他节点的区块高度
	// ----------------------------------------------------------------------
	for _, chainMakerContainerName := range pm.AllChainMakerContainerNames {
		if chainMakerContainerName != pm.NormalNode.ContainerName {
			otherBlockHeightFilePath := fmt.Sprintf("%s/%s/information.stat",
				configs.TopConfiguration.PathConfig.ConfigGeneratePath,
				chainMakerContainerName)
			var information *Information
			information, err = ReadInformation(otherBlockHeightFilePath)
			if err != nil {
				fmt.Printf("cannot retrieve block height")
				otherBlockHeight = 0
			} else {
				otherBlockHeight = information.BlockHeight
			}
			if otherBlockHeight > largestBlockHeight {
				largestBlockHeight = otherBlockHeight
			}
		}
	}
	// ----------------------------------------------------------------------

	// fabric 获取其他节点的区块高度
	for _, fabricContainerName := range pm.AllFabricContainerNames {
		if fabricContainerName != pm.NormalNode.ContainerName {
			otherBlockHeightFilePath := fmt.Sprintf("%s/%s/information.stat",
				configs.TopConfiguration.PathConfig.ConfigGeneratePath,
				fabricContainerName)
			var information *Information
			information, err = ReadInformation(otherBlockHeightFilePath)
			if err != nil {
				fmt.Printf("cannot retrieve block height")
				otherBlockHeight = 0
			} else {
				otherBlockHeight = information.BlockHeight
			}
			if otherBlockHeight > largestBlockHeight {
				largestBlockHeight = otherBlockHeight
			}
		}
	}

	// 计算当前节点的区块高度占整个区块高度的百分比
	// ----------------------------------------------------------------------
	if (currentBlockHeight == 0) && (largestBlockHeight == 0) {
		blockHeightPercentage = 100
	} else {
		blockHeightPercentage = float64(currentBlockHeight) / float64(largestBlockHeight) * 100
	}
	if len(pm.TimeList) == pm.FixedLength {
		pm.BlockHeightPercentageList = pm.BlockHeightPercentageList[1:]
		pm.BlockHeightPercentageList = append(pm.BlockHeightPercentageList, blockHeightPercentage)
	} else {
		pm.BlockHeightPercentageList = append(pm.BlockHeightPercentageList, blockHeightPercentage)
	}
	// ----------------------------------------------------------------------
	return nil
}
