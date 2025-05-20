package raspberrypi

import (
	"fmt"
	"github.com/spf13/cobra"
	"zhanghefan123/security_topology/modules/entities/real_entities/raspberrypi_topology"
	"zhanghefan123/security_topology/modules/logger"
)

var (
	cmdRaspberrypiClientLogger = logger.GetLogger(logger.ModuleMainCmdRaspberrypiClient)
)

func CreateRaspberrypiClientCmd() *cobra.Command {
	var configureRaspberrypiCmd = &cobra.Command{
		Use:   "raspberrypi",
		Short: "raspberrypi",
		Long:  "raspberrypi",
		Run: func(cmd *cobra.Command, args []string) {
			cmdRaspberrypiClientLogger.Infof("start to configure raspberrypi")
			err := clientCore()
			if err != nil {
				fmt.Printf("%v", err)
			}
		},
	}
	return configureRaspberrypiCmd
}

func clientCore() error {
	err := raspberrypi_topology.RaspberrypiTopologyInstance.Init()
	if err != nil {
		return fmt.Errorf("raspberrypi init failed %v", err)
	}
	return nil
}
