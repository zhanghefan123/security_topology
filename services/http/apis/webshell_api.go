package apis

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"zhanghefan123/security_topology/configs"
	"zhanghefan123/security_topology/modules/entities/real_entities/topology"
	"zhanghefan123/security_topology/modules/utils/network"
	"zhanghefan123/security_topology/modules/webshell"
)

type StartWebShellRequest struct {
	ContainerName string `json:"container_name"`
}

type StopWebShellRequest struct {
	Pid int `json:"pid"`
}

func StartWebShell(c *gin.Context) {
	// 如果已经不存在实例了则返回错误
	if topology.TopologyInstance == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "instance not created",
		})
		return
	}

	// 进行参数绑定
	startWebShellRequest := &StartWebShellRequest{}
	err := c.BindJSON(startWebShellRequest)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "could not bind start web shell params",
		})
	}

	// web 控制台信息
	var webShellInfo *webshell.WebShellInfo

	// 获取 ip 地址
	hostAddr := configs.TopConfiguration.NetworkConfig.LocalNetworkAddress

	// 获取可用端口
	availablePort, err := network.GetAvailablePort()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": fmt.Sprintf("get available port failed: %s", err.Error()),
		})
		return
	}

	// 获取初始化 command
	initCommand := "docker"
	initCommandArgs := []string{"exec", "-it", startWebShellRequest.ContainerName, "/bin/bash"}

	// 创建 webshell
	webShellInfo, err = webshell.StartWebShell(hostAddr, availablePort, true, initCommand, initCommandArgs, 5)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": fmt.Sprintf("start web shell failed: %s", err.Error()),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"address": webShellInfo.Address,
		"port":    webShellInfo.Port,
		"pid":     webShellInfo.Pid,
	})
}

func StopWebShell(c *gin.Context) {
	// 如果已经不存在实例了
	if topology.TopologyInstance == nil {
		c.JSON(http.StatusOK, gin.H{
			"message": "instance not created",
		})
		return
	}

	// 进行参数绑定
	stopWebShellRequest := &StopWebShellRequest{}
	err := c.BindJSON(stopWebShellRequest)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "could not bind stop web shell params",
		})
	}

	// 调用方法进行 webshell 的关闭
	err = webshell.StopWebShell(stopWebShellRequest.Pid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": fmt.Sprintf("stop web shell failed: %s", err.Error()),
		})
		return
	}

	// 打印关闭消息
	fmt.Println("stop web shell with pid: ", stopWebShellRequest.Pid)

	c.JSON(http.StatusOK, gin.H{
		"message": "stop web shell success",
	})
}
