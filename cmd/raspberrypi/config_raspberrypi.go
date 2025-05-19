package raspberrypi

import (
	"github.com/spf13/cobra"
	"zhanghefan123/security_topology/modules/logger"
)

var (
	cmdConfigRaspberrypiLogger = logger.GetLogger(logger.ModuleMainCmdConfigureRaspberrypi)
)

func CreateConfigureRaspberrypiCmd() *cobra.Command {
	var configureRaspberrypiCmd = &cobra.Command{
		Use:   "raspberrypi",
		Short: "raspberrypi",
		Long:  "raspberrypi",
		Run: func(cmd *cobra.Command, args []string) {
			cmdConfigRaspberrypiLogger.Infof("start to configure raspberrypi")
		},
	}
	return configureRaspberrypiCmd
}
