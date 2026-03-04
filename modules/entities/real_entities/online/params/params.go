package params

type GraphParams struct {
	SourceDestParams    *SourceDestParams    `json:"source_dest_params"`
	KShortestPathParmas *KShortestPathParmas `json:"k_shortest_path_params"`
	Nodes               []SimNodeParam       `json:"nodes"`
	Links               []SimLinkParam       `json:"links"`
	CoveragePaths       []string             `json:"coverage_paths"`
}

// SourceDestParams 图的源节点和目的节点的参数
type SourceDestParams struct {
	Source           string `json:"source"`
	SourceIndex      int    `json:"source_index"`
	Destination      string `json:"destination"`
	DestinationIndex int    `json:"destination_index"`
}

// KShortestPathParmas KShortestPath算法的参数
type KShortestPathParmas struct {
	NumberOfPaths int     `json:"number_of_paths"`
	LimitOfCost   float64 `json:"limit_of_cost"`
}

// SimNodeParam 模拟节点的参数
type SimNodeParam struct {
	Index int    `json:"index"`
	Type  string `json:"type"`
}

// IllegalRatio 链路的非法率分布
type IllegalRatio struct {
	Start float64 `json:"start"` // 起始值
	End   float64 `json:"end"`   // 结束值
}

// SimLinkParam 模拟链路的参数
type SimLinkParam struct {
	SourceNode   SimNodeParam `json:"source_node"`   // 链路的源节点
	TargetNode   SimNodeParam `json:"target_node"`   // 链路的目的节点
	IllegalRatio IllegalRatio `json:"illegal_ratio"` // 链路的非法率分布
}
