package performance_monitor

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"zhanghefan123/security_topology/configs"
)

// UpdateBlockHeightInfo 更新区块链高度信息
func (pm *PerformanceMonitor) UpdateBlockHeightInfo() (err error) {
	// 变量
	var blockHeightFilePath string
	var blockHeightFile *os.File
	var blockHeightFileBytes []byte
	var height int                    // 临时区块高度
	var currentNodeHeight int         // 当前节点的区块高度
	var largestHeight int             // 最高的区块高度
	var blockHeightPercentage float64 // 当前节点的区块 / 最大的区块高度

	// 获取文件路径并且打开文件
	// ----------------------------------------------------------------------
	blockHeightFilePath = fmt.Sprintf("%s/%s/peers_height.stat",
		configs.TopConfiguration.PathConfig.ConfigGeneratePath,
		pm.NormalNode.ContainerName)
	blockHeightFile, err = os.Open(blockHeightFilePath)
	if err != nil {
		return fmt.Errorf("cannot open %s for reason %w", blockHeightFilePath, err)
	}
	defer func() {
		errClose := blockHeightFile.Close()
		if errClose != nil {
			err = errClose
		}
	}()
	// ----------------------------------------------------------------------

	// 读取文件的内容并进行解析
	// ----------------------------------------------------------------------
	blockHeightFileBytes, err = io.ReadAll(blockHeightFile)
	if err != nil {
		return fmt.Errorf("cannot read %s for reason %w", blockHeightFilePath, err)
	}
	lines := strings.Split(string(blockHeightFileBytes), "\n")
	for _, line := range lines {
		splitParts := strings.Split(line, "->")
		if len(splitParts) != 3 {
			return fmt.Errorf("cannot parse line: %s", line)
		}
		if splitParts[0] == "self" {
			height, err = strconv.Atoi(splitParts[2])
			if err != nil {
				return fmt.Errorf("cannot parse height: %s", splitParts[2])
			}
			currentNodeHeight = height
			if height > largestHeight {
				largestHeight = height
			}
		} else if splitParts[0] == "other" {
			height, err = strconv.Atoi(splitParts[2])
			if err != nil {
				return fmt.Errorf("cannot parse height: %s", splitParts[2])
			}
			if height > largestHeight {
				largestHeight = height
			}
		} else {
			return fmt.Errorf("cannot parse line: %s", line)
		}
	}
	// ----------------------------------------------------------------------

	// 计算当前节点的区块高度占整个区块高度的百分比
	// ----------------------------------------------------------------------
	blockHeightPercentage = float64(currentNodeHeight) / float64(largestHeight) * 100
	if len(pm.TimeList) == pm.FixedLength {
		pm.BlockHeightPercentageList = pm.BlockHeightPercentageList[1:]
		pm.BlockHeightPercentageList = append(pm.BlockHeightPercentageList, blockHeightPercentage)
	} else {
		pm.BlockHeightPercentageList = append(pm.BlockHeightPercentageList, blockHeightPercentage)
	}
	// ----------------------------------------------------------------------
	return nil
}
