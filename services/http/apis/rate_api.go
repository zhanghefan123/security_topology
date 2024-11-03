package apis

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"net/http"
	"zhanghefan123/security_topology/modules/interface_rate"
)

type CaptureRateRequest struct {
	ContainerName string `json:"container_name"`
}

func StartCaptureInterfaceRate(c *gin.Context) {
	// 1. 进行参数绑定
	captureRateRequest := CaptureRateRequest{}
	err := c.ShouldBindJSON(&captureRateRequest)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "could not bind capture rate params",
		})
		return
	}
	// 2. 判断是否已经存在了相应的监听实例, 如果已经存在直接返回即可
	if _, ok := interface_rate.InterfaceRateMonitorMapping[captureRateRequest.ContainerName]; ok {
		c.JSON(http.StatusBadRequest, gin.H{
			"state": "down",
		})
		return
	}
	// 3. 拿到对应的抽象节点
	abstractNode := TopologyInstance.AbstractNodesMap[captureRateRequest.ContainerName]
	interfaceRateMonitor, err := interface_rate.NewInterfaceRateMonitor(abstractNode)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "could not create interface_rate monitor",
		})
		return
	}
	err = interfaceRateMonitor.CaptureInterfaceRate(abstractNode)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": fmt.Sprintf("could not capture interface rate: %s", err.Error()),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "successfully captured interface rate",
	})
}

func StopCaptureInterfaceRate(c *gin.Context) {
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
	err = interface_rate.RemoveInterfaceRateMonitor(abstractNode)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "could not remove interface_rate monitor",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "successfully stop captured interface rate",
	})
}
