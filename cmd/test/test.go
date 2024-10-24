package test

import (
	"github.com/spf13/cobra"
	"zhanghefan123/security_topology/modules/logger"
)

var (
	cmdTestLogger = logger.GetLogger(logger.ModuleMainCmdTest)
)

func CreateTestCommand() *cobra.Command {
	var createCommand = &cobra.Command{
		Use:   "test",
		Short: "test",
		Long:  "test",
		Run: func(cmd *cobra.Command, args []string) {
			cmdTestLogger.Infof("start test")
			// call prepare
			core()
		},
	}
	return createCommand
}

func core() {

}
