package main

import (
	"fmt"
	"os"
	"zhanghefan123/security_topology/cmd/constellation"
	"zhanghefan123/security_topology/cmd/http_service"
	"zhanghefan123/security_topology/cmd/images"
	"zhanghefan123/security_topology/cmd/root"
	"zhanghefan123/security_topology/cmd/test"
	"zhanghefan123/security_topology/configs"
)

func main() {
	// 首先进行配置的加载
	err := configs.InitLocalConfig()
	if err != nil {
		fmt.Printf("init local config failed, err:%v\n", err)
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
