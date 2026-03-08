package params

type GraphParams struct {
	SourceDestParams     *SourceDestParams    `json:"source_dest_params"`
	KShortestPathParamas *KShortestPathParmas `json:"k_shortest_path_params"`
	Nodes                []SimNodeParam       `json:"nodes"`
	PvLinks              []SimPvLinkParam     `json:"pv_links"`
	Links                []SimLinkParam       `json:"links"`
	CoveragePaths        []string             `json:"coverage_paths"`
}

// SourceDestParams 图的源节点和目的节点的参数
type SourceDestParams struct {
	Source      string `json:"source"`
	Destination string `json:"destination"`
}

// KShortestPathParmas KShortestPath算法的参数
type KShortestPathParmas struct {
	NumberOfPaths int     `json:"number_of_paths"`
	LimitOfCost   float64 `json:"limit_of_cost"`
}

// SimNodeParam 模拟节点的参数
type SimNodeParam struct {
	Index        int               `json:"index"`
	Type         string            `json:"type"`
	DropRatio    RatioDistribution `json:"drop_ratio"`
	CorruptRatio RatioDistribution `json:"corrupt_ratio"`
}

// RatioDistribution 丢包/篡改包的概率分布
type RatioDistribution struct {
	Start float64 `json:"start"` // 起始值
	End   float64 `json:"end"`   // 结束值
}

type SimPvLinkParam struct {
	SourceNode       SimNodeParam `json:"source_node"`       // 链路的源节点
	TargetNode       SimNodeParam `json:"target_node"`       // 链路的目的节点
	IntermediateNode SimNodeParam `json:"intermediate_node"` // pvlink 的中间节点
}

// SimLinkParam 模拟链路的参数
type SimLinkParam struct {
	SourceNode SimNodeParam `json:"source_node"` // 链路的源节点
	TargetNode SimNodeParam `json:"target_node"` // 链路的目的节点
}
