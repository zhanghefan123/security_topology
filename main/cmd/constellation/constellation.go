package constellation

import (
	"flag"
	"fmt"
	"github.com/spf13/cobra"
	"os"
	"os/signal"
	"zhanghefan123/security_topology/modules/config/system"
	"zhanghefan123/security_topology/modules/lifecycle"
	"zhanghefan123/security_topology/modules/logger"
)

var mainCmdConstellationMainLogger = logger.GetLogger(logger.ModuleMainCmdConstellation)

func CreateConstellationCmd() *cobra.Command {
	var constellationCmd = &cobra.Command{
		Use:   "constellation",
		Short: "manage constellation",
		Long:  "manage constellation",
		Run: func(cmd *cobra.Command, args []string) {
			mainCmdConstellationMainLogger.Infof("start manage the constellation")
			core()
		},
	}
	return constellationCmd
}

func core() {
	ParseFlag()

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	defer signal.Stop(signalChan)

	// 启动流程
	// =======================================================
	lifecycle.Initialize()
	PrintExitLogo()
	// =======================================================

	<-signalChan

	// 删除流程
	// =======================================================
	lifecycle.Delete()
	PrintRemovedLogo()
	// =======================================================

	fmt.Println(system.ConfigurationFilePath)

}

func ParseFlag() {
	flag.StringVar(&system.ConfigurationFilePath, "config", system.ConfigurationFilePath, "config file path")
}

func PrintExitLogo() {
	mainCmdConstellationMainLogger.Infof("<------------------------------------->")
	mainCmdConstellationMainLogger.Infof("        enter ctl+c exit        ")
	mainCmdConstellationMainLogger.Infof("<------------------------------------->")
}

func PrintRemovedLogo() {
	mainCmdConstellationMainLogger.Infof("<------------------------------------->")
	mainCmdConstellationMainLogger.Infof("        constellation killed        ")
	mainCmdConstellationMainLogger.Infof("<------------------------------------->")
}
