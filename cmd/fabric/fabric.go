package fabric

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
	"zhanghefan123/security_topology/configs"
	"zhanghefan123/security_topology/modules/logger"
	"zhanghefan123/security_topology/utils/dir"
	"zhanghefan123/security_topology/utils/execute"
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
	testNetworkPath := configs.TopConfiguration.FabricConfig.FabricNetworkPath
	_ = dir.WithContextManager(testNetworkPath, func() error {
		err := execute.Command("bash", []string{"-l", "-c", "echo $PATH"})
		if err != nil {
			return fmt.Errorf("start install channel failed: %w", err)
		}
		err = execute.Command("bash", []string{"-l", "-c", "echo $PATH"})
		if err != nil {
			return fmt.Errorf("start install chaincode failed: %w", err)
		}
		return nil
	})
	directory, _ := os.Getwd()
	fmt.Printf("current directory: %s\n", directory)
	return nil
}
