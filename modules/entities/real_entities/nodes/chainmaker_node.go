package nodes

import (
	"fmt"
	"zhanghefan123/security_topology/modules/entities/real_entities/normal_node"
	"zhanghefan123/security_topology/modules/entities/types"
)

type ChainmakerNode struct {
	*normal_node.NormalNode
}

// NewChainmakerNode 创建新的 chainmaker 节点
func NewChainmakerNode(nodeId int, X, Y float64) *ChainmakerNode {
	chainmakerNodeType := types.NetworkNodeType_ChainMakerNode
	normalNode := normal_node.NewNormalNode(
		types.NetworkNodeType_ChainMakerNode,
		nodeId,
		fmt.Sprintf("%s-%d", chainmakerNodeType.String(), nodeId),
	)
	normalNode.X = X
	normalNode.Y = Y
	return &ChainmakerNode{
		NormalNode: normalNode,
	}
}
