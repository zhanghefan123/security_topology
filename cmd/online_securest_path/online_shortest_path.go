package online_securest_path

import (
	"fmt"
	"github.com/spf13/cobra"
	"zhanghefan123/security_topology/modules/entities/real_entities/online/steps"
	"zhanghefan123/security_topology/modules/logger"
)

var (
	CmdOnlineSecurestPathCmd = logger.GetLogger(logger.ModuleOnlineSecurestPath)
	SimulationGraphPath      = "../resources/online_topologies/simple_topology.json"
	SimulatorParamsInstance  = &steps.SimulatorParams{
		NumberOfEpochs:          100,
		NumberOfPktsPerBatch:    100,
		Bias:                    0.5,
		ExploreRate:             0.1,
		LearningRate:            0.1,
		SmoothingFactor:         0.1,
		LaplaceSmoothingFactor:  0.1,
		HistoryForggetingFactor: 0.5,
		Lambda:                  5,
	}
)

func CreateOnlineSecurestPathCmd() *cobra.Command {
	var onlineSecurestPathCmd = &cobra.Command{
		Use:   "online_securest_path",
		Short: "online_securest_path",
		Long:  "online_securest_path",
		Run: func(cmd *cobra.Command, args []string) {
			// 1. log
			CmdOnlineSecurestPathCmd.Infof("start online securest path")
			// 2. start simulation
			StartOnlineSecurestPathSimulation(SimulationGraphPath)
		},
	}
	return onlineSecurestPathCmd
}

func StartOnlineSecurestPathSimulation(simulationGraphPath string) {
	// 1. 创建实例
	simulatorInstance := steps.NewSimulator(SimulatorParamsInstance, simulationGraphPath)
	// 2. 进行初始化
	err := simulatorInstance.Init()
	if err != nil {
		fmt.Printf("init simulator err: %v", err)
	}
	// 3. 进行 simulator 的运行
	err = simulatorInstance.Start()
	if err != nil {
		fmt.Printf("start simulator err: %v", err)
	}
}
