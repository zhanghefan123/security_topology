package images

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/spf13/cobra"
	"os/exec"
	"strings"
	"zhanghefan123/security_topology/main/cmd/tools"
	"zhanghefan123/security_topology/main/cmd/variables"
	"zhanghefan123/security_topology/modules/logger"
)

var (
	mainCmdImagesLogger = logger.GetLogger(logger.ModuleMainCmdImages)

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
			mainCmdImagesLogger.Infof("user choose to conduct %s operation to %s",
				variables.UserSelectedOperation, variables.UserSelectedImage)
			err := correctnessCheck()
			if err != nil {
				mainCmdImagesLogger.Error(err)
				return
			}
			core()
		},
	}
	tools.AttachFlags(createImagesCmd, []string{tools.FlagNameOfImageName, tools.FlagNameOfOperationType})
	return createImagesCmd
}

// correctnessCheck 正确性检查
func correctnessCheck() error {
	if _, ok := variables.ExistedImages[variables.UserSelectedImage]; !ok {
		return ErrNotSupportedImage
	}
	if _, ok := variables.AvailableOperations[variables.UserSelectedOperation]; !ok {
		return ErrNotSupportedOperation
	}

	RetrieveStatus()

	if (variables.UserSelectedOperation == variables.OperationBuild) && (variables.ExistedImages[variables.UserSelectedImage]) {
		return ErrImageAlreadyExists
	}
	if (variables.UserSelectedOperation == variables.OperationRemove) && (!(variables.ExistedImages[variables.UserSelectedImage])) {
		return ErrImageAlreadyDoesntExist
	}
	return nil
}

func buildImage() error {
	// 1.创建命令
	commandStr := fmt.Sprintf("build -t %s:latest -f ../images/%s/Dockerfile ../images/%s/",
		variables.UserSelectedImage, variables.UserSelectedImage, variables.UserSelectedImage)

	cmd := exec.Command("docker", strings.Split(commandStr, " ")...)
	var out bytes.Buffer
	cmd.Stdout = &out

	// 2.运行命令并检查是否有错误
	err := cmd.Run()
	if err != nil {
		return ErrExecuteBuildCommand
	}
	mainCmdImagesLogger.Infof("build image %s successfully", variables.UserSelectedImage)
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
	mainCmdImagesLogger.Infof("remove image %s successfully", variables.UserSelectedImage)
	return nil
}

func core() {
	if variables.UserSelectedOperation == variables.OperationBuild {
		err := buildImage()
		if err != nil {
			mainCmdImagesLogger.Error(err)
		}
	} else if variables.UserSelectedOperation == variables.OperationRebuild {
		err := removeImage()
		if err != nil {
			mainCmdImagesLogger.Error(err)
			return
		}
		err = buildImage()
		if err != nil {
			mainCmdImagesLogger.Error(err)
		}
	} else if variables.UserSelectedOperation == variables.OperationRemove {
		err := removeImage()
		if err != nil {
			mainCmdImagesLogger.Error(err)
		}
	}
}
