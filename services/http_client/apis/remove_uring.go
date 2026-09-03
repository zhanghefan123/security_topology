package apis

import (
	"fmt"
	"zhanghefan123/security_topology/configs"
	"zhanghefan123/security_topology/services/http_client"
)

const RemoveUringUrl = "removeUring"

func RemoveUring(nodeId int) error {
	// 监听接口的获取
	containerListenPort := configs.TopConfiguration.NetworkConfig.ValidationListenPort + nodeId
	// 进行 url 的构造
	removeUringUrl := fmt.Sprintf("http://%s:%d/%s", configs.TopConfiguration.NetworkConfig.BackendAddr,
		containerListenPort, RemoveUringUrl)
	// 进行 request
	err := http_client.PostJson(removeUringUrl, nil)
	if err != nil {
		return fmt.Errorf("remove uring failed: %v", err)
	}
	return nil
}
