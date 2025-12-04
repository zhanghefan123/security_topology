package performance_monitor

import (
	"fmt"
	"os"
	"zhanghefan123/security_topology/api/chain_api"
	"zhanghefan123/security_topology/cmd/variables"
	"zhanghefan123/security_topology/utils/file"
)

const (
	RemovePerformanceMonitors = "RemovePerformanceMonitors"
	StopLocalServices         = "StopLocalServices"
	WriteHeightIntoFile       = "WriteHeightIntoFile"
)

// RemoveFunction 删除函数
type RemoveFunction func() error

// RemoveModule 删除模块
type RemoveModule struct {
	remove         bool           // 是否进行删除 -> 只有相应的模块启动了才需要进行删除
	removeFunction RemoveFunction // 相应的删除函数
}

// Remove 删除整个拓扑
func (mm *MonitorManager) Remove() error {

	removeSteps := []map[string]RemoveModule{
		{RemovePerformanceMonitors: RemoveModule{true, mm.RemovePerformanceMonitors}},
		{StopLocalServices: RemoveModule{true, mm.StopLocalServices}},
		{WriteHeightIntoFile: RemoveModule{true, mm.WriteHeightIntoFile}},
	}
	err := mm.removeSteps(removeSteps)
	if err != nil {
		return fmt.Errorf("stop topology error: %w", err)
	}
	return nil
}

// removeStepsNum 获取删除模块的数量
func (mm *MonitorManager) removeStepsNum(removeSteps []map[string]RemoveModule) int {
	result := 0
	for _, removeStep := range removeSteps {
		for _, removeModule := range removeStep {
			if removeModule.remove {
				result += 1
			}
		}
	}
	return result
}

// startSteps 调用所有的启动方法
func (mm *MonitorManager) removeSteps(removeSteps []map[string]RemoveModule) (err error) {
	moduleNum := mm.removeStepsNum(removeSteps)
	count := 0
	for _, removeStep := range removeSteps {
		for name, removeModule := range removeStep {
			if removeModule.remove {
				if err = removeModule.removeFunction(); err != nil {
					return fmt.Errorf("remove step [%s] failed, %s", name, err)
				}
				monitorManagerLogger.Infof("BASE REMOVE STEP (%d/%d) => remove step [%s] success)", count+1, moduleNum, name)
				count += 1
			}
		}
	}
	return
}

// RemovePerformanceMonitors 进行所有的容器性能监视器的删除
func (mm *MonitorManager) RemovePerformanceMonitors() error {
	if _, ok := mm.MonitorManagerStopSteps[RemovePerformanceMonitors]; ok {
		monitorManagerLogger.Infof("already remove interface rate monitors")
		return nil
	}

	// 进行 timer 的重置
	// ---------------------------------------------------------------------------------------------------
	StopTicker()
	// ---------------------------------------------------------------------------------------------------

	// 向所有的 performanceMonitor 发送信号
	// ---------------------------------------------------------------------------------------------------
	err := RemoveAllPerformanceMonitors()
	if err != nil {
		return fmt.Errorf("remove all performance monitors error due to %v", err)
	}
	// ---------------------------------------------------------------------------------------------------

	// 进行 globalTxRateRecorder 的重置 (负责统计从最开始以来的tps rate的)
	// ---------------------------------------------------------------------------------------------------
	err = chain_api.StopGloablTxRateRecorder()
	if err != nil {
		return fmt.Errorf("stop global tx rate recorder")
	}
	// ---------------------------------------------------------------------------------------------------

	mm.MonitorManagerStopSteps[RemovePerformanceMonitors] = struct{}{}
	monitorManagerLogger.Infof("remove interface rate monitors")
	return nil
}

// StopLocalServices 进行本地服务的停止
func (mm *MonitorManager) StopLocalServices() error {
	if _, ok := mm.MonitorManagerStopSteps[StopLocalServices]; ok {
		monitorManagerLogger.Infof("already execute stop local services")
		return nil
	}

	// 停止 etcd watch
	mm.serviceContextCancelFunc()
	mm.EtcdWatchWaitGroup.Wait()

	// 停止 calculateMaxHeight
	mm.StopChannel <- struct{}{}
	mm.WaitGroup.Wait()

	mm.MonitorManagerStopSteps[StopLocalServices] = struct{}{}
	monitorManagerLogger.Infof("execute stop local services")
	return nil
}

func (mm *MonitorManager) WriteHeightIntoFile() error {
	if _, ok := mm.MonitorManagerStopSteps[WriteHeightIntoFile]; ok {
		monitorManagerLogger.Infof("already execute write result into file")
		return nil
	}

	directory := fmt.Sprintf("./result/result%d", variables.UserSelectedExperimentNumber)

	err := os.MkdirAll(fmt.Sprintf("%s", directory), os.ModePerm)
	if err != nil {
		return fmt.Errorf("mkdirall error: %v", err)
	}

	// 写入 max_block_height_list_all.txt
	finalString := ""
	for index := 0; index < len(mm.MaxBlockHeightListAll); index++ {
		if index == len(mm.MaxBlockHeightListAll)-1 {
			finalString += fmt.Sprintf("%f", mm.MaxBlockHeightListAll[index])
		} else {
			finalString += fmt.Sprintf("%f", mm.MaxBlockHeightListAll[index]) + ","
		}
	}
	err = file.WriteStringIntoFile(fmt.Sprintf("%s/max_block_height_list.txt", directory), finalString)
	if err != nil {
		return fmt.Errorf("write result into file failed: %v", err)
	}

	mm.MonitorManagerStopSteps[WriteHeightIntoFile] = struct{}{}
	monitorManagerLogger.Infof("execute write result into file")
	return nil
}
