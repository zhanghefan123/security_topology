package main

import (
	"fmt"
	"os"
	"path/filepath"
	"zhanghefan123/security_topology/cmd/constellation"
	"zhanghefan123/security_topology/cmd/http_service"
	"zhanghefan123/security_topology/cmd/images"
	"zhanghefan123/security_topology/cmd/root"
	"zhanghefan123/security_topology/cmd/test"
	"zhanghefan123/security_topology/configs"
	"zhanghefan123/security_topology/modules/utils/permission"
)

func main() {
	err := PrepareWorks()
	if err != nil {
		fmt.Printf("error executing prepare Works %v", err)
		return
	}
	rootCmd := root.CreateRootCmd()
	httpServiceCmd := http_service.CreateHttpServiceCmd()
	constellationCmd := constellation.CreateConstellationCmd()
	imagesCmd := images.CreateImagesCmd()
	testCmd := test.CreateTestCommand()
	rootCmd.AddCommand(httpServiceCmd)
	rootCmd.AddCommand(constellationCmd)
	rootCmd.AddCommand(imagesCmd)
	rootCmd.AddCommand(testCmd)
	err = rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

// PrepareWorks 准备工作
func PrepareWorks() error {
	// 首先进行配置的加载
	err := configs.InitLocalConfig()
	if err != nil {
		return fmt.Errorf("init local config failed, err:%w", err)
	}
	// 1. 然后为 gotty 文件分配可执行权限
	gottyFilePath := configs.TopConfiguration.PathConfig.GottyPath
	err = permission.AddExecutePermission(gottyFilePath)
	if err != nil {
		return fmt.Errorf("add execute permission to %s failed, err:%w", gottyFilePath, err)
	}
	// 2. 然后为 chainmakerCryptoGen 分配可执行权限
	cryptoGenProjectPath := configs.TopConfiguration.ChainMakerConfig.CryptoGenProjectPath
	cryptoGenBinPath := filepath.Join(cryptoGenProjectPath, "bin/chainmaker-cryptogen")
	err = permission.AddExecutePermission(cryptoGenBinPath)
	if err != nil {
		return fmt.Errorf("add execute permission to %s failed, err:%w", cryptoGenBinPath, err)
	}
	return nil
}
