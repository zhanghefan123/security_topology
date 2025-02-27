package images

import (
	"fmt"
	"strings"
	"zhanghefan123/security_topology/configs"
	"zhanghefan123/security_topology/modules/utils/dir"
	"zhanghefan123/security_topology/modules/utils/execute"
)

// buildImageForFabricPeer 为 fabric peer 进行镜像的构建
func buildImageForFabricPeer() error {
	fabricProjectPath := configs.TopConfiguration.FabricConfig.FabricProjectPath
	err := dir.WithContextManager(fabricProjectPath, func() error {
		return nil
	})

	commandStr := fmt.Sprintf("build -f images/order/Dockerfile --build-arg GO_TAGS= -t hyperledger/fabric-orderer -t hyperledger/fabric-orderer:3.0.0 -t hyperledger/fabric-orderer:3.0 .")
	err = execute.Command("docker", strings.Split(commandStr, " "))
	if err != nil {
		return fmt.Errorf("build fabric peer image err: %s", err.Error())
	} else {
		return nil
	}
}

// buildImageForFabricOrder 为 fabric orderer 进行镜像的构建
func buildImageForFabricOrder() error {
	fabricProjectPath := configs.TopConfiguration.FabricConfig.FabricProjectPath
	err := dir.WithContextManager(fabricProjectPath, func() error {
		return nil
	})
	commandStr := fmt.Sprintf("build -f images/peer/Dockerfile --build-arg GO_TAGS= -t hyperledger/fabric-peer -t hyperledger/fabric-peer:3.0.0 -t hyperledger/fabric-peer:3.0 .")
	err = execute.Command("docker", strings.Split(commandStr, " "))
	if err != nil {
		return fmt.Errorf("build fabric peer image err: %s", err.Error())
	} else {
		return nil
	}
}
