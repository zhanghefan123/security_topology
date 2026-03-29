package online_securest_path

import (
	"fmt"
	"zhanghefan123/security_topology/modules/entities/real_entities/online/entities"
	"zhanghefan123/security_topology/modules/entities/real_entities/online/steps"
	"zhanghefan123/security_topology/modules/entities/real_entities/online/types"
	"zhanghefan123/security_topology/modules/logger"

	"github.com/spf13/cobra"
)

var (
	ExperimentResultsDir     = "./output"
	CmdOnlineSecurestPathCmd = logger.GetLogger(logger.ModuleCmdOnlineSecurestPath)
	SimulationGraphPath      = "../resources/online_topologies/simple_topology.json"
	SimulatorParamsInstance  = &steps.SimulatorParams{
		NumberOfEpochs:         400,
		NumberOfPktsPerBatch:   2000,
		Bias:                   0.01,
		ExploreRate:            0.02, // 这个越大, 节点越不可能因为一次偶然的情况被抬高
		LearningRate:           0.1,  // 学习率越大, 越容易因为近期表现好而大幅度的调整选择的策略 (如果一个表现差的突然变好，需要累积一段时间之后再), 但是学习率越大, 也可能导致最近一个变好的迅速抢占。
		SmoothingFactor:        0.1,
		LaplaceSmoothingFactor: 0.00,
		BalancingFactor:        0.01,
		Lambda:                 5,                                      // 这个变量非常的重要，不然可能无法收敛
		SizeOfBloomFilter:      400,                                    // 布隆过滤器的大小
		HashOfBloomFilter:      1,                                      // 布隆过滤器哈希函数个数
		MinimumDeliveryRatio:   0.8,                                    // 最小的交付率 (这个需要根据对网络的预期进行设置)
		GainCalculationStyle:   types.GainCalculationMode_SumEdgeGains, // 收益的计算模式
		SimulationStrategy:     types.SimStrategy_PerPacketAck,
	}
	SimulationEvents = []*entities.SimEvent{
		{
			//StartEpoch: 100,
			//UpdateRouters: []*entities.UpdateNormalRouter{
			//	{
			//		NormalRouterName:  "NormalRouter-6",
			//		StartIllegalRatio: 0.8,
			//		EndIllegalRatio:   0.9,
			//		StartDropRatio:    0.0,
			//		EndDropRatio:      0.0,
			//	},
			//	{
			//		NormalRouterName:  "NormalRouter-7",
			//		StartIllegalRatio: 0.0,
			//		EndIllegalRatio:   0.1,
			//		StartDropRatio:    0.0,
			//		EndDropRatio:      0.0,
			//	},
			//},
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
	err = simulatorInstance.StartExp3()
	if err != nil {
		fmt.Printf("start simulator err: %v", err)
		return
	}
	// 4. 进行结果的获取
	err = simulatorInstance.GetStatistics(experimentResultsDir)
	if err != nil {
		fmt.Printf("get statistics err: %v", err)
		return
	}
}
