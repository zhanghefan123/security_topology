package raspberrypi

import (
	"fmt"
	"github.com/spf13/cobra"
	"zhanghefan123/security_topology/modules/entities/real_entities/raspberrypi_topology"
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
			err := core()
			if err != nil {
				fmt.Printf("%v", err)
			}
		},
	}
	return configureRaspberrypiCmd
}

func core() error {
	rpt := raspberrypi_topology.NewRaspeberryTopology()
	err := rpt.Init()
	if err != nil {
		return fmt.Errorf("raspberrypi init failed %v", err)
	}
	return nil
}
