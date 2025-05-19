package http_service

import (
	"fmt"
	"github.com/spf13/cobra"
	"zhanghefan123/security_topology/modules/logger"
)

var (
	cmdHttpServiceLogger = logger.GetLogger(logger.ModuleMainCmdHttpService)
)

// CreateHttpServiceCmd http_service 命令
func CreateHttpServiceCmd() *cobra.Command {
	var httpServiceCmd = &cobra.Command{
		Use:   "http_service",
		Short: "http_service",
		Long:  "http_service",
		Run: func(cmd *cobra.Command, args []string) {
			cmdHttpServiceLogger.Infof("start http service")
			core()
		},
	}
	return httpServiceCmd
}

// core http_service 命令的核心
func core() {
	fmt.Println("not support windows")
}
