package apis

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"net/http"
	"zhanghefan123/security_topology/modules/performance_monitor"
)

type CaptureRateRequest struct {
	ContainerName string `json:"container_name"`
}

// StartCaptureInstancePerformance 开启接口速率监听
func StartCaptureInstancePerformance(c *gin.Context) {
	// 1. 进行参数绑定 -> 从而进行容器名的获取
	var performanceMonitor *performance_monitor.PerformanceMonitor
	var ok bool
	captureRateRequest := CaptureRateRequest{}
	err := c.ShouldBindJSON(&captureRateRequest)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "could not capture instance performance",
		})
		return
	}
	// 2. 判断是否已经存在了相应的监听实例, 如果已经存在就进行数据的返回
	if performanceMonitor, ok = performance_monitor.PerformanceMonitorMapping[captureRateRequest.ContainerName]; ok {
		// 2.1 如果已经存在，则进行数据的返回
		c.JSON(http.StatusOK, gin.H{
			"time_list":           performanceMonitor.TimeList,
			"interface_rate_list": performanceMonitor.InterfaceRateList,
			"cpu_ratio_list":      performanceMonitor.CpuRatioList,
			"memory_list":         performanceMonitor.MemoryMBList,
		})
		return
	} else {
		// 2.2 如果不存在，则创建新的并返回空的数据
		abstractNode := TopologyInstance.AbstractNodesMap[captureRateRequest.ContainerName]
		performanceMonitor, err = performance_monitor.NewInstancePerformanceMonitor(abstractNode)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "could not create performance_monitor monitor",
			})
			return
		}
		performanceMonitor.KeepGettingPerformance()
		c.JSON(http.StatusOK, gin.H{
			"message":             "successfully captured instance performance",
			"time_list":           make([]int, 0),
			"interface_rate_list": make([]float64, 0),
			"cpu_ratio_list":      make([]float64, 0),
		})
		return
	}

}

// StopCaptureInstancePerformance 停止接口速率监听
func StopCaptureInstancePerformance(c *gin.Context) {
	// 1. 如果已经不存在了就返回错误
	if TopologyInstance == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"state": "down",
		})
		return
	}
	body, err := ioutil.ReadAll(c.Request.Body)
	// 2. 进行参数绑定
	captureRateRequest := CaptureRateRequest{}
	err = json.Unmarshal(body, &captureRateRequest)
	//err = c.ShouldBindJSON(&captureRateRequest)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "could not bind capture rate params",
		})
		return
	}
	// 3. 拿到对应的抽象节点并调用 Remove 逻辑
	abstractNode := TopologyInstance.AbstractNodesMap[captureRateRequest.ContainerName]
	err = performance_monitor.RemovePerformanceMonitor(abstractNode)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "could not remove performance_monitor monitor",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "successfully stop captured instance performance",
	})
}
