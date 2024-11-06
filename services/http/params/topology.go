package params

type NodeParam struct {
	Index int     `json:"index"`
	Type  string  `json:"type"`
	X     float64 `json:"x"`
	Y     float64 `json:"y"`
}

type LinkParam struct {
	SourceNode NodeParam `json:"source_node"`
	TargetNode NodeParam `json:"target_node"`
	LinkType   string    `json:"link_type"` // 链路的类型 - 可能是接入 access, 也可能是骨干 backbone
}

type TopologyParams struct {
	NetworkEnv           string      `json:"network_env"`
	BlockChainType       string      `json:"blockchain_type"`
	ConsensusType        string      `json:"consensus_type"`
	AccessLinkBandwidth  int         `json:"access_link_bandwidth"`
	ConsensusNodeCpu     float64     `json:"consensus_node_cpu"`    // 单位为个
	ConsensusNodeMemory  float64     `json:"consensus_node_memory"` // 单位为 MB
	ConsensusThreadCount int         `json:"consensus_thread_count"`
	Nodes                []NodeParam `json:"nodes"` // 所有的节点
	Links                []LinkParam `json:"links"` // 所有的链路
}
