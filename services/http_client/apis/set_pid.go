package apis

import (
	"fmt"
	"zhanghefan123/security_topology/configs"
	"zhanghefan123/security_topology/services/http_client"
	"zhanghefan123/security_topology/services/http_client/params"
)

const SetPidUrl = "setPid"

func SetPid(nodeId int, pid int) error {
	// 监听接口的获取
	containerListenPort := configs.TopConfiguration.NetworkConfig.ValidationListenPort + nodeId
	// 进行 url 的构造
	setPidUrl := fmt.Sprintf("http://%s:%d/%s", configs.TopConfiguration.NetworkConfig.BackendAddr,
		containerListenPort, SetPidUrl)
	// 进行 setPidParams 的构造
	pidInformation := params.NewPidInformation(pid)
	// 进行 request
	err := http_client.PostJson(setPidUrl, pidInformation)
	if err != nil {
		return fmt.Errorf("set pid failed: %v", err)
	}
	return nil
}
