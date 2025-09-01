package images

import (
	"fmt"
	"zhanghefan123/security_topology/configs"
	"zhanghefan123/security_topology/modules/utils/dir"
	"zhanghefan123/security_topology/modules/utils/execute"
)

// buildImageForFiscoBcos 为 FiscoBcos 进行镜像构建, 需要注意的是, 这个镜像一定要在可执行文件进行重新编译之后才能调用
func buildImageForFiscoBcos() error {
	fiscoBcosProjectPath := configs.TopConfiguration.FiscoBcosConfig.ProjectPath
	err := dir.WithContextManager(fiscoBcosProjectPath, func() error {
		var commandArgs = []string{
			"build", "-t", "fiscoorg/fiscobcos:v3.12.1",
			"-f", "./tools/.ci/Dockerfile",
			".",
		}
		err := execute.Command("docker", commandArgs)
		if err != nil {
			return fmt.Errorf("build fisco bcos image err: %s", err.Error())
		} else {
			return nil
		}
	})
	if err != nil {
		return fmt.Errorf("build fisco bcos image err: %s", err.Error())
	} else {
		return nil
	}
}

func removeImageForFiscoBcos() error {
	imageName := "fiscoorg/fiscobcos:v3.12.1"
	var commandArgs = []string{
		"rmi", imageName,
	}
	err := execute.Command("docker", commandArgs)
	if err != nil {
		fmt.Printf("remove image %s error %v\n", imageName, err)
	}
	return nil
}
