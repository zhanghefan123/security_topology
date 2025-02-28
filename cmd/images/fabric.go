package images

import (
	"fmt"
	"os"
	"strings"
	"zhanghefan123/security_topology/configs"
	"zhanghefan123/security_topology/modules/utils/dir"
	"zhanghefan123/security_topology/modules/utils/execute"
)

// buildImageForFabricPeer 为 fabric peer 进行镜像的构建
func buildImageForFabricPeer() error {
	fabricProjectPath := configs.TopConfiguration.FabricConfig.FabricProjectPath
	fmt.Println("fabricProjectPath:", fabricProjectPath)
	err := dir.WithContextManager(fabricProjectPath, func() error {
		currentDir, err := os.Getwd()
		fmt.Println("current directory: ", currentDir)
		var commandArgs []string = []string{
			"build", "--force-rm", "-f", "./images/peer/Dockerfile",
			"-t", "hyperledger/fabric-peer",
			"-t", "hyperledger/fabric-peer:3.0.0",
			"-t", "hyperledger/fabric-peer:3.0",
			".",
		}
		err = execute.Command("docker", commandArgs)
		if err != nil {
			return fmt.Errorf("build fabric peer image err: %s", err.Error())
		} else {
			return nil
		}
	})
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
		var commandArgs []string = []string{
			"build", "--force-rm", "-f", "./images/orderer/Dockerfile",
			"-t", "hyperledger/fabric-orderer",
			"-t", "hyperledger/fabric-orderer:3.0.0",
			"-t", "hyperledger/fabric-orderer:3.0",
			".",
		}
		err := execute.Command("docker", commandArgs)
		if err != nil {
			return fmt.Errorf("build fabric peer image err: %s", err.Error())
		} else {
			return nil
		}
	})
	if err != nil {
		return fmt.Errorf("build fabric peer image err: %s", err.Error())
	} else {
		return nil
	}
}

// removeImageForFabricPeer 删除 peer image
func removeImageForFabricPeer() error {
	var imageNames []string = []string{
		"hyperledger/fabric-peer:latest",
		"hyperledger/fabric-peer:3.0.0",
		"hyperledger/fabric-peer:3.0",
	}
	for _, imageName := range imageNames {
		err := execute.Command("docker", strings.Split(fmt.Sprintf("rmi %s", imageName), " "))
		if err != nil {
			fmt.Printf("remove image %s err: %v\n", imageName, err)
		}
	}
	return nil
}

// removeImageForFabricOrder 删除 orderer image
func removeImageForFabricOrder() error {
	var imageNames []string = []string{
		"hyperledger/fabric-orderer:latest",
		"hyperledger/fabric-orderer:3.0.0",
		"hyperledger/fabric-orderer:3.0",
	}
	for _, imageName := range imageNames {
		err := execute.Command("docker", strings.Split(fmt.Sprintf("rmi %s", imageName), " "))
		if err != nil {
			fmt.Printf("remove image %s error %v\n", imageName, err)
		}
	}
	return nil
}
