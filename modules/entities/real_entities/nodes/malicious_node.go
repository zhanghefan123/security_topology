package nodes

import (
	"fmt"
	"zhanghefan123/security_topology/modules/entities/real_entities/normal_node"
	"zhanghefan123/security_topology/modules/entities/types"
)

type MaliciousNode struct {
	*normal_node.NormalNode
}

// NewMaliciousNode 创建新的恶意节点
func NewMaliciousNode(nodeId, X, Y int) *MaliciousNode {
	routerType := types.NetworkNodeType_Router
	normalNode := normal_node.NewNormalNode(
		types.NetworkNodeType_MaliciousNode,
		nodeId,
		fmt.Sprintf("%s-%d", routerType.String(), nodeId),
	)
	normalNode.X = X
	normalNode.Y = Y
	return &MaliciousNode{
		NormalNode: normalNode,
	}
}
