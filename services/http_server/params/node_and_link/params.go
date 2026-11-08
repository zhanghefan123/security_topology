package node_and_link

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

func ResolveNodeNameWithNodeParam(nodeParam *NodeParam) (string, error) {
	nodeType, err := types.ResolveNodeType(nodeParam.Type)
	if err != nil {
		return "", fmt.Errorf("resolve node type failed, %s", err.Error())
	}
	nodeName := fmt.Sprintf("%s-%d", nodeType.String(), nodeParam.Index)
	return nodeName, nil
}
