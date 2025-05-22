package fabric

import (
	"fmt"
	"github.com/spf13/cobra"
	"time"
	"zhanghefan123/security_topology/api/fabric_api"
	"zhanghefan123/security_topology/modules/logger"
)

var (
	cmdFabricLogger = logger.GetLogger(logger.ModuleMainCmdFabric)
)

func CreateFabricCmd() *cobra.Command {
	var fabricCmd = &cobra.Command{
		Use:   "fabric",
		Short: "fabric",
		Long:  "fabric",
		Run: func(cmd *cobra.Command, args []string) {
			cmdFabricLogger.Infof("start fabric command")
			err := core()
			if err != nil {
				fmt.Printf("start fabric error: %v\n", err)
			}
		},
	}
	return fabricCmd
}

func core() error {
	txRateRecorder := fabric_api.NewTxRateRecorder()
	err := txRateRecorder.StartTxRateTest(5)
	if err != nil {
		return fmt.Errorf("start tps test error: %v", err)
	}
	timer := time.NewTimer(time.Second * 10)
	select {
	case <-timer.C:
		txRateRecorder.StopTxRateTestCore()
	}
	return nil
}
