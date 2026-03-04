package steps

import (
	"zhanghefan123/security_topology/modules/entities/real_entities/online/entities"
)

type SimulatorParams struct {
	NumberOfEpochs          int     // 总的批次数量
	NumberOfPktsPerBatch    int     // 单个批次内部的包的数量
	Bias                    float64 // 偏置项，防止 gain 过小导致的数值稳定性问题
	ExploreRate             float64 // 探索率，控制探索和利用的权衡，0 代表完全利用，1 代表完全探索
	LearningRate            float64 // 学习率，控制每次更新权重的幅度
	SmoothingFactor         float64 // 平滑因子，控制在计算非法比例时进行拉普拉斯平滑的程度，0 代表不进行平滑，数值越大代表平滑程度越高
	LaplaceSmoothingFactor  float64 // 拉普拉斯平滑因子，控制在计算非法比例时进行拉普拉斯平滑的程度，0 代表不进行平滑，数值越大代表平滑程度越高
	HistoryForggetingFactor float64 // 历史遗忘因子，0 代表完全不遗忘，1 代表完全遗忘
	Lambda                  float64 // Lambda 参数，控制在计算 gain 时对非法比例的敏感程度，0 代表完全不敏感，数值越大代表越敏感
}

type Simulator struct {
	SimGraph            *entities.SimGraph  // 模拟图
	SimulationGraphPath string              // 图配置文件路径
	SimulatorParams     *SimulatorParams    // simulator 相关参数
	SimulatorInitSteps  map[string]struct{} // simulator 初始化步骤
}

// NewSimulator 进行 simulator 的创建
func NewSimulator(simulatorParams *SimulatorParams, simulationGraphPath string) *Simulator {
	return &Simulator{
		SimGraph:            nil,
		SimulationGraphPath: simulationGraphPath,
		SimulatorParams:     simulatorParams,
		SimulatorInitSteps:  make(map[string]struct{}),
	}
}
