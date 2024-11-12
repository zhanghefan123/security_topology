package performance_monitor

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// UpdateInstanceCpuAndMemoryRatio 更新实例的 Cpu 和 内存信息
func (pm *PerformanceMonitor) UpdateInstanceCpuAndMemoryRatio() (err error) {
	// 打开 cgroupFile
	// ---------------------------------------------------------------------------------------------
	processCgroupFile, err := os.Open(fmt.Sprintf("/proc/%d/cgroup", pm.NormalNode.Pid))
	if err != nil {
		return fmt.Errorf("cannot open process cgroup file: %v", err)
	}
	defer func() {
		errClose := processCgroupFile.Close()
		if err == nil {
			err = errClose
		}
	}()
	// ---------------------------------------------------------------------------------------------

	// 读取 cgroupFile 的内容 -> 样例 0::/system.slice/docker-9700fc09f052961f4d6ee4a505f325e2f2f81ee1f215f9bfe331f66dc42783c6.scope
	// ---------------------------------------------------------------------------------------------
	processCgroupFileBytes, err := io.ReadAll(processCgroupFile)
	splitParts := strings.Split(string(processCgroupFileBytes), "::")
	if len(splitParts) != 2 {
		return fmt.Errorf("cannot parse process cgroup file")
	}
	// ---------------------------------------------------------------------------------------------

	// 获取容器相应的资源文件存放目录
	// ---------------------------------------------------------------------------------------------
	resourcesDirBase := fmt.Sprintf("/sys/fs/cgroup/%s/", strings.TrimSpace(splitParts[1]))
	cpuPath := filepath.Join(resourcesDirBase, "cpu.stat")
	memoryPath := filepath.Join(resourcesDirBase, "memory.current")
	// ---------------------------------------------------------------------------------------------

	// 读取相应的文件
	// ---------------------------------------------------------------------------------------------
	usageUsec, err := ReadCpuUsage(cpuPath)
	if err != nil {
		return fmt.Errorf("cannot read CPU usage: %w", err)
	}
	memoryBytes, err := ReadMemoryUsage(memoryPath)
	if err != nil {
		return fmt.Errorf("cannot read memory usage: %w", err)
	}
	memoryMBytes := memoryBytes / 1024 / 1024
	// ---------------------------------------------------------------------------------------------
	if err != nil {
		return fmt.Errorf("cannot get host resources: %w", err)
	}

	// 更新 CPU 和 内存
	// ---------------------------------------------------------------------------------------------
	containerCpuBusy := usageUsec / 1000
	cpuRatio := (containerCpuBusy - pm.LastCpuBusy) / 1000
	pm.LastCpuBusy = containerCpuBusy
	if len(pm.TimeList) == pm.FixedLength {
		pm.CpuRatioList = pm.CpuRatioList[1:]
		pm.CpuRatioList = append(pm.CpuRatioList, cpuRatio)
		pm.MemoryMBList = pm.MemoryMBList[1:]
		pm.MemoryMBList = append(pm.MemoryMBList, memoryMBytes)
	} else {
		pm.CpuRatioList = append(pm.CpuRatioList, cpuRatio)
		pm.MemoryMBList = append(pm.MemoryMBList, memoryMBytes)
	}
	// ---------------------------------------------------------------------------------------------
	return nil
}

// ReadCpuUsage 读取 Cpu 利用率
/*
cpuUsageFilePath 的文件格式, 注意 usage_usec 代表的是 cpu 使用的微秒数
usage_usec 2341374
user_usec 668964
system_usec 1672410
nr_periods 0
nr_throttled 0
throttled_usec 0
nr_bursts 0
burst_usec 0
*/
func ReadCpuUsage(cpuUsageFilePath string) (float64, error) {
	var resultInteger int
	cpuUsageFile, err := os.Open(cpuUsageFilePath)
	if err != nil {
		return -1, fmt.Errorf("cannot open cpu usage file: %w", err)
	}
	contentInBytes, err := io.ReadAll(cpuUsageFile)
	allLines := strings.Split(string(contentInBytes), "\n")
	for _, line := range allLines {
		if strings.Contains(line, "usage_usec") {
			splitContent := strings.Split(line, " ")
			resultInteger, err = strconv.Atoi(splitContent[1])
			if err != nil {
				return -1, fmt.Errorf("cannot parse cpu usage file: %w", err)
			}
			break
		}
	}
	return float64(resultInteger), nil
}

// ReadMemoryUsage 读取内存使用
func ReadMemoryUsage(memoryUsageFilePath string) (float64, error) {
	memoryUsageFile, err := os.Open(memoryUsageFilePath)
	if err != nil {
		return -1, fmt.Errorf("cannot open memory usage file: %w", err)
	}
	contentInBytes, err := io.ReadAll(memoryUsageFile)
	resultInteger, err := strconv.Atoi(strings.TrimRight(string(contentInBytes), "\n"))
	if err != nil {
		return -1, fmt.Errorf("cannot parse memory usage file: %w", err)
	}
	return float64(resultInteger), nil
}
