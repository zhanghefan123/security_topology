package steps

import (
	"zhanghefan123/security_topology/modules/entities/real_entities/online/entities"
	"zhanghefan123/security_topology/modules/entities/real_entities/online/types"
	"zhanghefan123/security_topology/modules/logger"
)

var (
	SimulatorLogger = logger.GetLogger(logger.ModuleSimulator)
)

type SimulatorParams struct {
	NumberOfEpochs         int                       // 总的批次数量
	NumberOfPktsPerBatch   int                       // 单个批次内部的包的数量
	Bias                   float64                   // 偏置项，防止 gain 过小导致的数值稳定性问题
	ExploreRate            float64                   // 探索率，控制探索和利用的权衡，0 代表完全利用，1 代表完全探索， 确保覆盖集之中的都能被探索到
	LearningRate           float64                   // 学习率，控制每次更新权重的幅度
	SmoothingFactor        float64                   // 平滑因子，控制在计算非法比例时进行拉普拉斯平滑的程度，0 代表不进行平滑，数值越大代表平滑程度越高
	LaplaceSmoothingFactor float64                   // 拉普拉斯平滑因子，控制在计算非法比例时进行拉普拉斯平滑的程度，0 代表不进行平滑，数值越大代表平滑程度越高
	BalancingFactor        float64                   // BalancingFactor 越大，分配给全局平均的比例越大，探索越强，反之利用越强
	Lambda                 float64                   // Lambda 参数，控制在计算 gain 时对非法比例的敏感程度，0 代表完全不敏感，数值越大代表越敏感
	SizeOfBloomFilter      int                       // 布隆过滤器的比特数量
	HashOfBloomFilter      int                       // 布隆过滤器的哈希函数个数
	MinimumDeliveryRatio   float64                   // 最小的交付率
	GainCalculationStyle   types.GainCalculationMode // 收益计算模式
	SimulationStrategy     types.SimStrategy         // 进行模拟的模式
}

type Simulator struct {
	SimGraph            *entities.SimGraph   // 模拟图
	SimulationGraphPath string               // 图配置文件路径
	SimulatorParams     *SimulatorParams     // simulator 相关参数
	SimulatorInitSteps  map[string]struct{}  // simulator 初始化步骤
	SimEvents           []*entities.SimEvent // 模拟事件
}

// NewSimulator 进行 simulator 的创建
func NewSimulator(simulatorParams *SimulatorParams, simulationGraphPath string, simulationEvents []*entities.SimEvent) *Simulator {
	return &Simulator{
		SimGraph:            nil,
		SimulationGraphPath: simulationGraphPath,
		SimulatorParams:     simulatorParams,
		SimulatorInitSteps:  make(map[string]struct{}),
		SimEvents:           simulationEvents,
	}
}
