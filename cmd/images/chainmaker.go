package images

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"zhanghefan123/security_topology/cmd/variables"
	"zhanghefan123/security_topology/configs"
	"zhanghefan123/security_topology/modules/utils/dir"
	"zhanghefan123/security_topology/modules/utils/execute"
)

// buildImageForChainMaker 为 chainmaker 进行镜像的构建
/*
# ---------------------------- zeusnet 添加的代码 ----------------------------
第一步
# zeusnet add code: 添加一个预构建的镜像 (用来添加依赖的 go文件)
previous-go-build:
rm -rf build/ data/ log/ bin/
docker build -t chainmaker-go -f ./DOCKER/Dockerfile_go .

第二步
# zeusnet add code: 构建最终的镜像
final_image:
docker build -t chainmaker -f ./DOCKER/Dockerfile_final .
docker tag chainmaker chainmaker:${VERSION}
# ---------------------------- zeusnet 添加的代码 ----------------------------
*/
func buildImageForChainMakerEnv() error {
	chainMakerGoProjectPath := configs.TopConfiguration.ChainMakerConfig.ChainMakerGoProjectPath
	err := dir.WithContextManager(chainMakerGoProjectPath, func() error {
		// 1. remove dirs
		removedDirectories := []string{
			filepath.Join(chainMakerGoProjectPath, "/build/"),
			filepath.Join(chainMakerGoProjectPath, "/data/"),
			filepath.Join(chainMakerGoProjectPath, "/log/"),
			filepath.Join(chainMakerGoProjectPath, "/bin/"),
		}
		for _, removedDir := range removedDirectories {
			err := os.RemoveAll(removedDir)
			if err != nil {
				return fmt.Errorf("fail to remove dir %v", err)
			}
		}

		// 2. execute command to build chainmaker-go (go environment for chainmaker) (param1 -> image name) (param2 -> docker file name)
		commandStr := fmt.Sprintf("build -t %s -f ./DOCKER/%s .", variables.ImageNameChainMakerEnv, "Dockerfile_chainmaker_env")
		err := execute.Command("docker", strings.Split(commandStr, " "))
		if err != nil {
			return fmt.Errorf("fail to build chainmaker-env (go environment for chainmaker)")
		}
		return nil
	})
	if err != nil {
		return err
	} else {
		return nil
	}
}

// buildImageForChainMaker 进行 chainmaker 镜像的构建
func buildImageForChainMaker() error {
	chainMakerGoProjectPath := configs.TopConfiguration.ChainMakerConfig.ChainMakerGoProjectPath
	err := dir.WithContextManager(chainMakerGoProjectPath, func() error {
		// 1. execute command to build chainmaker (param1 -> image name) (param2 -> docker file name)
		commandStr := fmt.Sprintf("build -t %s -f ./DOCKER/%s .", variables.ImageNameChainMaker, "Dockerfile_chainmaker")
		err := execute.Command("docker", strings.Split(commandStr, " "))
		if err != nil {
			return fmt.Errorf("fail to build chainmaker")
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("build image for chainmaker failed")
	} else {
		return nil
	}
}
