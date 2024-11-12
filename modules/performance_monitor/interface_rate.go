package performance_monitor

import (
	"fmt"
	"io"
	"os"
	"strings"
	"zhanghefan123/security_topology/modules/entities/real_entities/interface_data"
	"zhanghefan123/security_topology/modules/entities/types"
)

func (pm *PerformanceMonitor) UpdateInterfaceRate() error {
	content, err := ReadNetworkInterfaceFile(pm.NormalNode.Pid)
	if err != nil {
		return fmt.Errorf("ReadNetworkInterfaceFile error: %w", err)
	}
	networkInterfaceLines := strings.Split(content, "\n")
	firstInterfaceName := fmt.Sprintf("%s%d_idx%d", types.GetPrefix(pm.NormalNode.Type), pm.NormalNode.Id, 1)
	for _, networkInterfaceLine := range networkInterfaceLines {
		if strings.Contains(networkInterfaceLine, firstInterfaceName) {
			interfaceData := interface_data.ResolveNetworkInterfaceLine(networkInterfaceLine) // 第一个是 loop back
			currentReceivedBytes := interfaceData.RxBytes
			delta := float64(currentReceivedBytes - pm.LastReceivedBytes)
			dataRate := delta / float64(1024) / float64(1024)
			pm.LastReceivedBytes = currentReceivedBytes
			if len(pm.TimeList) == pm.FixedLength {
				pm.InterfaceRateList = pm.InterfaceRateList[1:]
				pm.InterfaceRateList = append(pm.InterfaceRateList, dataRate)
			} else {
				pm.InterfaceRateList = append(pm.InterfaceRateList, dataRate)
			}
			// 否则进行 for 循环的跳出
			break
		} else {
			// 否则继续
			continue
		}
	}
	return nil
}

// ReadNetworkInterfaceFile 进行网络接口文件的读取
func ReadNetworkInterfaceFile(pid int) (result string, err error) {
	var bytesContent []byte
	networkInterfaceDataFile := fmt.Sprintf("/proc/%d/net/dev", pid)
	file, err := os.Open(networkInterfaceDataFile)
	defer func() {
		errClose := file.Close()
		if err == nil {
			err = errClose
		}
	}()
	if err != nil {
		return "", fmt.Errorf("open file failed: %w", err)
	}
	bytesContent, err = io.ReadAll(file)
	if err != nil {
		return "", fmt.Errorf("read network interface file failed: %w", err)
	}
	stringContent := string(bytesContent)
	return stringContent, nil
}
