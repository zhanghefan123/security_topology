package performance_monitor

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"zhanghefan123/security_topology/configs"
)

// UpdateBlockHeightInfo 更新区块链高度信息
func (pm *PerformanceMonitor) UpdateBlockHeightInfo() (err error) {
	var currentBlockHeight int        // 当前节点的区块高度
	var largestBlockHeight int        // 最高的区块高度
	var blockHeightPercentage float64 // 当前节点的区块 / 最大的区块高度
	var otherBlockHeight int          // 其他节点的区块高度

	// 获取当前节点的区块高度
	// ----------------------------------------------------------------------
	currentBlockHeightFilePath := fmt.Sprintf("%s/%s/block_height.stat",
		configs.TopConfiguration.PathConfig.ConfigGeneratePath,
		pm.NormalNode.ContainerName)
	currentBlockHeight, err = ReadBlockHeight(currentBlockHeightFilePath)
	if err != nil {
		return fmt.Errorf("read current block height error: %w", err)
	}
	if currentBlockHeight > largestBlockHeight {
		largestBlockHeight = currentBlockHeight
	}
	// ----------------------------------------------------------------------

	// 获取其他节点的区块高度
	// ----------------------------------------------------------------------
	for _, chainMakerContainerName := range pm.AllChainMakerContainerNames {
		if chainMakerContainerName != pm.NormalNode.ContainerName {
			otherBlockHeightFilePath := fmt.Sprintf("%s/%s/block_height.stat",
				configs.TopConfiguration.PathConfig.ConfigGeneratePath,
				chainMakerContainerName)
			otherBlockHeight, err = ReadBlockHeight(otherBlockHeightFilePath)
			if err != nil {
				return fmt.Errorf("read other block height error: %w", err)
			}
			if otherBlockHeight > largestBlockHeight {
				largestBlockHeight = otherBlockHeight
			}
		}
	}
	// ----------------------------------------------------------------------

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

// ReadBlockHeight 进行区块的高度的读取
func ReadBlockHeight(blockHeightFilePath string) (result int, err error) {
	var contentInBytes []byte
	var finalResult int
	var blockHeightFile *os.File
	blockHeightFile, err = os.Open(blockHeightFilePath)
	if err != nil {
		return -1, fmt.Errorf("cannot open %s for reason %w", blockHeightFilePath, err)
	}
	defer func() {
		closeErr := blockHeightFile.Close()
		if err == nil {
			err = closeErr
		}
	}()
	contentInBytes, err = io.ReadAll(blockHeightFile)
	if err != nil {
		return -1, fmt.Errorf("cannot read %s for reason %w", blockHeightFilePath, err)
	}
	finalResult, err = strconv.Atoi(string(contentInBytes))
	if err != nil {
		return -1, fmt.Errorf("cannot parse %s for reason %w", string(contentInBytes), err)
	}
	return finalResult, nil
}
