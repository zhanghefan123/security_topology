package apis

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"zhanghefan123/security_topology/modules/entities/real_entities/performance_monitor"
	"zhanghefan123/security_topology/modules/entities/real_entities/topology"
	"zhanghefan123/security_topology/utils/judge"

	"github.com/gin-gonic/gin"
)

type CapturePerformanceRequest struct {
	ContainerName string `json:"container_name"`
}

// StartCaptureInstancePerformance 开启接口速率监听
func StartCaptureInstancePerformance(c *gin.Context) {
	if topology.Instance == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "already shutdown",
		})
		return
	}
	// 1. 进行参数绑定 -> 从而进行容器名的获取
	var performanceMonitor *performance_monitor.PerformanceMonitor
	var ok bool
	captureRateRequest := CapturePerformanceRequest{}
	err := c.ShouldBindJSON(&captureRateRequest)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "could not capture instance performance",
		})
		return
	}

	fmt.Printf("capture %v\n", captureRateRequest.ContainerName)

	// 2. 判断是否已经存在了相应的监听实例, 如果已经存在就进行数据的返回
	if performanceMonitor, ok = performance_monitor.PerformanceMonitorMapping[captureRateRequest.ContainerName]; ok {
		// 判断节点的类型
		// 判断是否是区块链的类型

		if judge.IsBlockChainType(performanceMonitor.NormalNode.Type) {
			c.JSON(http.StatusOK, gin.H{
				"time_list":                 performanceMonitor.TimeList,
				"interface_rate_list":       performanceMonitor.InterfaceRateList,
				"cpu_ratio_list":            performanceMonitor.CpuRatioList,
				"memory_list":               performanceMonitor.MemoryMBList,
				"block_ratio_list":          performanceMonitor.BlockHeightPercentageList,
				"connected_count_list":      performanceMonitor.ConnectedCountList,
				"half_connected_count_list": performanceMonitor.HalfConnectedCountList,
				"time_out_list":             performanceMonitor.RequestTimeoutList,
				"message_count_list":        performanceMonitor.MessageCountList,
			})
		} else {
			// 如果是其他类型 (并非长安链节点类型)
			c.JSON(http.StatusOK, gin.H{
				"time_list":           performanceMonitor.TimeList,
				"interface_rate_list": performanceMonitor.InterfaceRateList,
				"cpu_ratio_list":      performanceMonitor.CpuRatioList,
				"memory_list":         performanceMonitor.MemoryMBList,
			})
		}
		return
	}
	// 已经不用 else 的条件了, 因为默认所有的都已经创建好了
	//} else {
	//	fmt.Println("empty")
	//	// 2.2 如果不存在，则创建新的并返回空的数据
	//	abstractNode := topology.Instance.AbstractNodesMap[captureRateRequest.ContainerName]
	//	// 获取所有的 chainMakerContainer 的 name
	//	performanceMonitor, err = monitor.NewInstancePerformanceMonitor(abstractNode)
	//	if err != nil {
	//		c.JSON(http.StatusInternalServerError, gin.H{
	//			"message": "could not create performance_monitor monitor",
	//		})
	//		return
	//	}
	//	performance_monitor.KeepGettingPerformance(performanceMonitor)
	//	c.JSON(http.StatusOK, gin.H{
	//		"message":             "successfully captured instance performance",
	//		"time_list":           make([]int, 0),
	//		"interface_rate_list": make([]float64, 0),
	//		"cpu_ratio_list":      make([]float64, 0),
	//	})
	//	return
	//}

}

// StopCaptureInstancePerformance 停止接口速率监听
func StopCaptureInstancePerformance(c *gin.Context) {
	// 1. 如果已经不存在了就返回错误
	if topology.Instance == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"state": "down",
		})
		return
	}
	body, err := ioutil.ReadAll(c.Request.Body)
	// 2. 进行参数绑定
	captureRateRequest := CapturePerformanceRequest{}
	err = json.Unmarshal(body, &captureRateRequest)
	//err = c.ShouldBindJSON(&captureRateRequest)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "could not bind capture rate params",
		})
		return
	}

	// 2. 进行写入
	abstractNode := topology.Instance.AbstractNodesMap[captureRateRequest.ContainerName]
	performanceMonitor, err := performance_monitor.GetPerformanceMonitor(abstractNode)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "get performance monitor failed",
		})
		return
	}
	if performanceMonitor == nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "performance monitor not found",
		})
		return
	}
	err = performance_monitor.WriteResultIntoFile(performanceMonitor)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "write result into file error",
		})
		return
	}

	// 3. 拿到对应的抽象节点并调用 Remove 逻辑
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
