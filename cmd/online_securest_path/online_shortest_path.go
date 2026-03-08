package online_securest_path

import (
	"fmt"
	"zhanghefan123/security_topology/modules/entities/real_entities/online/entities"
	"zhanghefan123/security_topology/modules/entities/real_entities/online/steps"
	"zhanghefan123/security_topology/modules/logger"

	"github.com/spf13/cobra"
)

var (
	ExperimentResultsDir     = "./output"
	CmdOnlineSecurestPathCmd = logger.GetLogger(logger.ModuleCmdOnlineSecurestPath)
	SimulationGraphPath      = "../resources/online_topologies/simple_topology.json"
	SimulatorParamsInstance  = &steps.SimulatorParams{
		NumberOfEpochs:         100,
		NumberOfPktsPerBatch:   2000,
		Bias:                   0.01,
		ExploreRate:            0.01,
		LearningRate:           0.3,
		SmoothingFactor:        0.1,
		LaplaceSmoothingFactor: 0.01,
		BalancingFactor:        0.01,
		Lambda:                 5, // 这个变量非常的重要，不然可能无法收敛
		LowerBoundLegalRatio:   0.9,
	}
	SimulationEvents = []*entities.SimEvent{
		{
			StartEpoch: 100,
			UpdateLinks: []*entities.UpdateLink{
				{
					PvLinkDescription: fmt.Sprintf("%s->%s->%s", "PathValidationRouter-4", "NormalRouter-5", "PathValidationRouter-7"),
					StartIllegalRatio: 0.8,
					EndIllegalRatio:   0.9,
					StartDropRatio:    0.0,
					EndDropRatio:      0.0,
				},
				{
					PvLinkDescription: fmt.Sprintf("%s->%s->%s", "PathValidationRouter-4", "NormalRouter-6", "PathValidationRouter-7"),
					StartIllegalRatio: 0.0,
					EndIllegalRatio:   0.1,
					StartDropRatio:    0.0,
					EndDropRatio:      0.0,
				},
			},
		},
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
			StartOnlineSecurestPathSimulation(SimulationGraphPath, ExperimentResultsDir)
		},
	}
	return onlineSecurestPathCmd
}

func StartOnlineSecurestPathSimulation(simulationGraphPath, experimentResultsDir string) {
	// 1. 创建实例
	simulatorInstance := steps.NewSimulator(SimulatorParamsInstance, simulationGraphPath, SimulationEvents)
	// 2. 进行初始化
	err := simulatorInstance.Init()
	if err != nil {
		fmt.Printf("init simulator err: %v", err)
		return
	}
	// 3. 进行 simulator 的运行
	err = simulatorInstance.Start()
	if err != nil {
		fmt.Printf("start simulator err: %v", err)
		return
	}
	// 4. 进行结果的获取
	err = simulatorInstance.GetStatistics(ExperimentResultsDir)
	if err != nil {
		fmt.Printf("get statistics err: %v", err)
		return
	}
}
