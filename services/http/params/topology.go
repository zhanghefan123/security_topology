package params

import "zhanghefan123/security_topology/modules/entities/types"

type NodeParam struct {
	Index int     `json:"index"`
	Type  string  `json:"type"`
	X     float64 `json:"x"`
	Y     float64 `json:"y"`
}

type LinkParam struct {
	SourceNode          NodeParam `json:"source_node"` // 链路的源节点
	TargetNode          NodeParam `json:"target_node"` // 链路的目的节点
	SourceInterfaceName string    // 源接口名称 (只在 raspberry pi 之中使用)
	TargetInterfaceName string    // 目的接口名称 (只在 raspberry pi 之中使用)
	LinkType            string    `json:"link_type"` // 链路的类型 - 可能是接入 access, 也可能是骨干 backbone
}

type TopologyParams struct {
	NetworkEnv           string          `json:"network_env"`
	BlockChainTypeString string          `json:"blockchain_type"`
	BlockChainType       types.ChainType // 根据 BlockChainTypeString 计算出来的
	ConsensusType        string          `json:"consensus_type"`
	AccessLinkBandwidth  int             `json:"access_link_bandwidth"`
	ConsensusNodeCpu     float64         `json:"consensus_node_cpu"`    // 单位为个
	ConsensusNodeMemory  float64         `json:"consensus_node_memory"` // 单位为 MB
	ConsensusThreadCount int             `json:"consensus_thread_count"`
	Nodes                []NodeParam     `json:"nodes"`         // 所有的节点
	Links                []LinkParam     `json:"links"`         // 所有的链路
	StartDefence         bool            `json:"start_defence"` // 是否开启防御
}

type StartDefenceParameter struct {
	StartDefence bool `json:"start_defence"`
}

func ResolveBlockChainType(blockChainTypeString string) types.ChainType {
	if blockChainTypeString == "长安链" {
		return types.ChainType_ChainMaker
	} else if blockChainTypeString == "fabric" {
		return types.ChainType_HyperledgerFabric
	} else if blockChainTypeString == "fisco-bcos" {
		return types.ChainType_FiscoBcos
	} else {
		return types.ChainType_NonChain
	}
}
