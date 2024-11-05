package host_resources

import (
	"fmt"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
)

type HostResources struct {
	CpuTotal        float64
	CpuBusy         float64
	UsedMemoryByte  uint64
	TotalMemoryByte uint64
	UsedSwapMemByte uint64
	TotalSwapByte   uint64
}

// GetHostResources 获取主机资源
func GetHostResources() (*HostResources, error) {
	var result = new(HostResources)
	cpuStateInfo, err := cpu.Times(false)
	if err != nil {
		return nil, fmt.Errorf("get cpu infor error %w", err)
	}
	result.CpuTotal = cpuStateInfo[0].Total() - cpuStateInfo[0].Guest - cpuStateInfo[0].GuestNice
	result.CpuBusy = result.CpuTotal - cpuStateInfo[0].Idle - cpuStateInfo[0].Iowait
	memStateInfo, err := mem.VirtualMemory()
	if err != nil {
		return nil, fmt.Errorf("get mem info error %w", err)
	}
	result.UsedSwapMemByte = memStateInfo.Total - memStateInfo.Free
	result.TotalMemoryByte = memStateInfo.Total
	result.UsedSwapMemByte = memStateInfo.SwapTotal - memStateInfo.SwapFree
	result.TotalMemoryByte = memStateInfo.SwapTotal
	result.CpuTotal = result.CpuTotal * 1000
	result.CpuBusy = result.CpuBusy * 1000
	return result, nil
}
