package images

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
	"os/exec"
	"strings"
	"zhanghefan123/security_topology/cmd/tools"
	"zhanghefan123/security_topology/cmd/variables"
	"zhanghefan123/security_topology/configs"
	"zhanghefan123/security_topology/modules/logger"
)

var (
	cmdImagesLogger = logger.GetLogger(logger.ModuleMainCmdImages)
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
		return fmt.Errorf("not supported image")
	}

	// 判断是否支持相应的操作
	if _, ok := variables.AvailableOperations[variables.UserSelectedOperation]; !ok {
		return fmt.Errorf("not supported operation")
	}

	// 获取当前生成的镜像
	err := tools.RetrieveStatus()
	if err != nil {
		return fmt.Errorf("retrieve status failed: %v", err)
	}

	// 如果是构建镜像，但是镜像已经构建，那么返回错误
	if (variables.UserSelectedOperation == variables.OperationBuild) && (variables.ExistedImages[variables.UserSelectedImage]) {
		return fmt.Errorf("image %s is already built", variables.UserSelectedImage)
	}

	// 如果是所有镜像, 只能是重建
	if (variables.UserSelectedImage == variables.AllImages) && (variables.UserSelectedOperation != variables.OperationRebuild) {
		return fmt.Errorf("only could rebuild all images")
	}

	return nil
}

// buildImage 构建镜像
func buildImage(userSelectedImage string) error {

	var commandStr string

	realTimePositionDir := configs.TopConfiguration.PathConfig.RealTimePositionDir

	// 1. 区分不同的镜像, 创建 build 命令
	if userSelectedImage == variables.ImageNamePosition {
		commandStr = fmt.Sprintf("build -t %s:latest -f ../../%s/Dockerfile ../../%s/",
			userSelectedImage, realTimePositionDir, realTimePositionDir)
		fmt.Println(commandStr)
	} else {
		commandStr = fmt.Sprintf("build -t %s:latest -f ../images/%s/Dockerfile ../images/%s/",
			userSelectedImage, userSelectedImage, userSelectedImage)
		fmt.Println(commandStr)
	}

	cmd := exec.Command("docker", strings.Split(commandStr, " ")...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// 2.运行命令并检查是否有错误
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("build image %s failed: %v", userSelectedImage, err)
	}

	// 3. 如果没有错误, 输出结果
	cmdImagesLogger.Infof("build image %s successfully", userSelectedImage)
	return nil
}

// removeImage 进行镜像的删除
func removeImage(userSelectedImage string) error {
	// 判断是否存在
	if ok := variables.ExistedImages[variables.UserSelectedImage]; !ok {
		cmdImagesLogger.Infof("image %s is not built", variables.UserSelectedImage)
		return nil
	}

	// 1. 创建命令
	commandStr := fmt.Sprintf("rmi %s", userSelectedImage)
	cmd := exec.Command("docker", strings.Split(commandStr, " ")...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// 2. 运行命令并检查是否有错误
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("remove image %s failed: %w", userSelectedImage, err)
	}

	// 3. 日志输出
	cmdImagesLogger.Infof("remove image %s successfully", userSelectedImage)
	return nil
}

// core 核心处理逻辑
func core() error {
	userSelectedImage := variables.UserSelectedImage

	// 进行所有的镜像的重建
	if userSelectedImage == variables.AllImages {
		// 进行所有的镜像的删除
		for _, image := range variables.ImagesInBuildOrder {
			// 判断这些镜像是否存在
			if exist, ok := variables.ExistedImages[image]; ok && exist {
				err := removeImage(image)
				if err != nil {
					return fmt.Errorf("remove image %s failed: %v", image, err)
				}
			}
		}
		// 按照指定的顺序进行镜像的生成
		for _, image := range variables.ImagesInBuildOrder {
			err := buildImage(image)
			if err != nil {
				return fmt.Errorf("build image %s failed: %v", image, err)
			}
		}
		return nil
	} else {
		switch variables.UserSelectedOperation {
		case variables.OperationBuild:
			err := buildImage(userSelectedImage)
			if err != nil {
				return fmt.Errorf("build image failed: %v", err)
			}
		case variables.OperationRebuild:
			err := removeImage(userSelectedImage)
			if err != nil {
				return fmt.Errorf("remove image failed: %v", err)
			}
			err = buildImage(userSelectedImage)
			if err != nil {
				return fmt.Errorf("build image failed: %v", err)
			}
		case variables.OperationRemove:
			err := removeImage(userSelectedImage)
			if err != nil {
				return fmt.Errorf("remove image failed: %v", err)
			}
		}
		return nil
	}
}
