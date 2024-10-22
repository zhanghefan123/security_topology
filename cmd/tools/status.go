package tools

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
	"zhanghefan123/security_topology/cmd/variables"
)

// InitializeExistedImages 进行状态的重置
func InitializeExistedImages() {
	variables.ExistedImages = map[string]bool{
		variables.ImageUbuntuWithSoftware: false,
		variables.ImagePythonEnv:          false,
		variables.ImageGoEnv:              false,
		variables.ImageNormalSatellite:    false,
		variables.ImageNameEtcd:           false,
		variables.ImageNamePosition:       false,
	}
}

// RetrieveStatus 获取镜像的状态
func RetrieveStatus() error {
	// 进行状态的重置
	InitializeExistedImages()

	// 创建并执行 "docker images" 命令
	cmd := exec.Command("docker", "images")

	// 使用 bytes.Buffer 来捕获标准输出
	var out bytes.Buffer
	cmd.Stdout = &out

	// 运行命令并检查是否有错误
	err := cmd.Run()
	if err != nil {
		//mainCmdStatusLogger.Errorf("Error running docker images command: %v\n", err)
		return fmt.Errorf("Error running docker images command: %v\n", err)
	}

	// 获取命令输出结果并转换为字符串
	output := out.String()

	// 进一步处理输出，比如将其按行分割
	lines := strings.Split(output, "\n")

	// 遍历每一行
	for i, line := range lines {
		// 跳过第一行表头
		if i == 0 {
			continue
		}

		// 打印每一行的镜像信息
		line = strings.TrimSpace(line)
		if len(line) > 0 {
			differentParts := strings.Split(line, " ")
			imageName := differentParts[0]
			if _, ok := variables.ExistedImages[imageName]; ok {
				variables.ExistedImages[imageName] = true
			}
		}
	}
	return nil
}
