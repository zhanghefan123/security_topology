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
func NewMaliciousNode(nodeId int, X, Y float64) *MaliciousNode {
	maliciousNodeType := types.NetworkNodeType_MaliciousNode
	normalNode := normal_node.NewNormalNode(
		types.NetworkNodeType_MaliciousNode,
		nodeId,
		fmt.Sprintf("%s-%d", maliciousNodeType.String(), nodeId),
	)
	normalNode.X = X
	normalNode.Y = Y
	return &MaliciousNode{
		NormalNode: normalNode,
	}
}
