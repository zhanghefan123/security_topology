package params

import (
	"fmt"
	"zhanghefan123/security_topology/modules/entities/types"
)

type RatioDistribution struct {
	Start int `json:"start"` // 起始值
	End   int `json:"end"`   // 结束值
}

type MaliciousParams struct {
	CorruptRatio              RatioDistribution `json:"corrupt_ratio"`
	CorruptSpecialPacketRatio RatioDistribution `json:"corrupt_special_packet_ratio"`
}

type SpecialParams struct {
	InnerRouterType types.InnerRouterType `json:"inner_router_type"` // 内部的类型 (即这是一个非可信的路由器还是一个可信的路径验证路由器)
	MaliciousParams MaliciousParams       `json:"malicious_params"`  // 恶意参数
}

type NodeParam struct {
	Index         int           `json:"index"`
	Type          string        `json:"type"`
	X             float64       `json:"x"`
	Y             float64       `json:"y"`
	SpecialParams SpecialParams `json:"special_params"`
}

type LinkParam struct {
	SourceNode          NodeParam `json:"source_node"` // 链路的源节点
	TargetNode          NodeParam `json:"target_node"` // 链路的目的节点
	SourceInterfaceName string    // 源接口名称 (只在 raspberry pi 之中使用)
	TargetInterfaceName string    // 目的接口名称 (只在 raspberry pi 之中使用)
	LinkType            string    `json:"link_type"` // 链路的类型 - 可能是接入 access, 也可能是骨干 backbone
}

type BlockChainParams struct {
	ConsensusThreadCount int             `json:"consensus_thread_count"`
	BlockChainTypeString string          `json:"blockchain_type"`
	BlockChainType       types.ChainType // 根据 BlockChainTypeString 计算出来的
	ConsensusType        string          `json:"consensus_type"`
	StartDefence         bool            `json:"start_defence"`
}

type SecPathMabParams struct {
	TopologyType           int `json:"topology_type"`
	ExperimentType         int `json:"experiment_type"`
	SecPathMabType         int `json:"sec_path_mab_type"`
	NumberOfHops           int `json:"number_of_hops"`
	NumberOfSegmentsPerHop int `json:"number_of_segments_per_hop"`
}

type TopologyParams struct {
	Nodes        []NodeParam `json:"nodes"`
	Links        []LinkParam `json:"links"`
	PerLinkDelay float64     `json:"per_link_delay"`
}

type TopologyStartParams struct {
	BlockChainParams BlockChainParams `json:"block_chain_params"`
	SecPathMabParams SecPathMabParams `json:"sec_path_mab_params"`
	TopologyParams   TopologyParams   `json:"topology_params"`
}

type StartDefenceParameter struct {
	StartDefence bool `json:"start_defence"`
}

func ResolveNodeNameWithNodeParam(nodeParam *NodeParam) (string, error) {
	nodeType, err := types.ResolveNodeType(nodeParam.Type)
	if err != nil {
		return "", fmt.Errorf("resolve node type failed, %s", err.Error())
	}
	nodeName := fmt.Sprintf("%s-%d", nodeType.String(), nodeParam.Index)
	return nodeName, nil
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
