package images

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/spf13/cobra"
	"os/exec"
	"strings"
	"zhanghefan123/security_topology/cmd/tools"
	"zhanghefan123/security_topology/cmd/variables"
	"zhanghefan123/security_topology/modules/logger"
)

var (
	cmdImagesLogger = logger.GetLogger(logger.ModuleMainCmdImages)

	ErrNotSupportedImage       = errors.New("not supported image")
	ErrNotSupportedOperation   = errors.New("not supported operation")
	ErrImageAlreadyExists      = errors.New("image already exists")
	ErrImageAlreadyDoesntExist = errors.New("image already does not exist")
	ErrExecuteBuildCommand     = errors.New("err execute build command")
	ErrExecuteRemoveCommand    = errors.New("err execute remove command")
)

// CreateImagesCmd 创建镜像管理子命令
func CreateImagesCmd() *cobra.Command {
	var createImagesCmd = &cobra.Command{
		Use:   "images",
		Short: "manage images",
		Long:  "manage images",
		Run: func(cmd *cobra.Command, args []string) {
			cmdImagesLogger.Infof("user choose to conduct %s operation to %s",
				variables.UserSelectedOperation, variables.UserSelectedImage)
			err := correctnessCheck()
			if err != nil {
				cmdImagesLogger.Errorf("correctness check failed: %v", err)
				return
			}
			// 核心处理逻辑
			err = core()
			if err != nil {
				cmdImagesLogger.Errorf("core failed: %v", err)
			}
		},
	}
	tools.AttachFlags(createImagesCmd, []string{tools.FlagNameOfImageName, tools.FlagNameOfOperationType})
	return createImagesCmd
}

// correctnessCheck 正确性检查
func correctnessCheck() error {
	// 判断是否支持相应的镜像
	if _, ok := variables.ExistedImages[variables.UserSelectedImage]; !ok {
		return ErrNotSupportedImage
	}

	// 判断是否支持相应的操作
	if _, ok := variables.AvailableOperations[variables.UserSelectedOperation]; !ok {
		return ErrNotSupportedOperation
	}

	// 获取当前生成的镜像
	err := tools.RetrieveStatus()
	if err != nil {
		return fmt.Errorf("retrieve status failed: %v", err)
	}

	// 如果是构建镜像，但是镜像已经构建，那么返回错误
	if (variables.UserSelectedOperation == variables.OperationBuild) && (variables.ExistedImages[variables.UserSelectedImage]) {
		return ErrImageAlreadyExists
	}

	// 如果是删除镜像，但是镜像已经被删除，那么返回错误
	if (variables.UserSelectedOperation == variables.OperationRemove) && (!(variables.ExistedImages[variables.UserSelectedImage])) {
		return ErrImageAlreadyDoesntExist
	}
	return nil
}

// buildImage 构建镜像
func buildImage() error {
	// 1.创建命令
	commandStr := fmt.Sprintf("build -t %s:latest -f ../images/%s/Dockerfile ../images/%s/",
		variables.UserSelectedImage, variables.UserSelectedImage, variables.UserSelectedImage)

	fmt.Println(commandStr)

	cmd := exec.Command("docker", strings.Split(commandStr, " ")...)
	var out bytes.Buffer
	cmd.Stdout = &out

	// 2.运行命令并检查是否有错误
	err := cmd.Run()
	if err != nil {
		return ErrExecuteBuildCommand
	}

	// 3. 如果没有错误, 输出结果
	cmdImagesLogger.Infof("build image %s successfully", variables.UserSelectedImage)
	return nil
}

func removeImage() error {
	// 1. 创建命令
	commandStr := fmt.Sprintf("rmi %s", variables.UserSelectedImage)
	cmd := exec.Command("docker", strings.Split(commandStr, " ")...)
	var out bytes.Buffer
	cmd.Stdout = &out

	// 2. 运行命令并检查是否有错误
	err := cmd.Run()
	if err != nil {
		return ErrExecuteRemoveCommand
	}

	// 3. 日志输出
	cmdImagesLogger.Infof("remove image %s successfully", variables.UserSelectedImage)
	return nil
}

// core 核心处理逻辑
func core() error {
	if variables.UserSelectedOperation == variables.OperationBuild { // 处理 build 命令
		err := buildImage()
		if err != nil {
			return fmt.Errorf("build image failed: %v", err)
		}
	} else if variables.UserSelectedOperation == variables.OperationRebuild { // 处理 rebuild 命令
		err := removeImage()
		if err != nil {
			return fmt.Errorf("remove image failed: %v", err)
		}
		err = buildImage()
		if err != nil {
			return fmt.Errorf("build image failed: %v", err)
		}
	} else if variables.UserSelectedOperation == variables.OperationRemove { // 处理 remove 命令
		err := removeImage()
		if err != nil {
			return fmt.Errorf("remove image failed: %v", err)
		}
	}
	return nil
}
