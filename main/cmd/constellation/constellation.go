package constellation

import (
	"flag"
	"fmt"
	"github.com/spf13/cobra"
	"os"
	"os/signal"
	"zhanghefan123/security_topology/modules/logger"
	"zhanghefan123/security_topology/modules/sysconfig"
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
	Initialize()
	PrintExitLogo()
	// =======================================================

	<-signalChan

	// 删除流程
	// =======================================================
	Delete()
	PrintRemovedLogo()
	// =======================================================

	fmt.Println(sysconfig.ConfigurationFilePath)

}

func ParseFlag() {
	flag.StringVar(&sysconfig.ConfigurationFilePath, "config", sysconfig.ConfigurationFilePath, "config file path")
}

func PrintExitLogo() {
	mainCmdConstellationMainLogger.Infof("<------------------------------------->")
	mainCmdConstellationMainLogger.Infof("        enter ctl+c exit        ")
	mainCmdConstellationMainLogger.Infof("<------------------------------------->")
	fmt.Println()
}

func PrintRemovedLogo() {
	mainCmdConstellationMainLogger.Infof("<------------------------------------->")
	mainCmdConstellationMainLogger.Infof("        constellation killed        ")
	mainCmdConstellationMainLogger.Infof("<------------------------------------->")
}
