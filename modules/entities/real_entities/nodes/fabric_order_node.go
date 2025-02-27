package nodes

import (
	"fmt"
	"zhanghefan123/security_topology/modules/entities/real_entities/normal_node"
	"zhanghefan123/security_topology/modules/entities/types"
)

type FabricOrderNode struct {
	*normal_node.NormalNode
}

// NewFabricOrderNode 创建新的 FabricOrder
func NewFabricOrderNode(nodeId int, X, Y float64) *FabricOrderNode {
	fabricPeerNodeType := types.NetworkNodeType_FabricOrderNode
	normalNode := normal_node.NewNormalNode(
		types.NetworkNodeType_FabricOrderNode,
		nodeId,
		fmt.Sprintf("%s-%d", fabricPeerNodeType.String(), nodeId),
	)
	normalNode.X = X
	normalNode.Y = Y
	return &FabricOrderNode{normalNode}
}
