package nodes

import (
	"fmt"
	"zhanghefan123/security_topology/modules/entities/real_entities/normal_node"
	"zhanghefan123/security_topology/modules/entities/types"
)

type LiRNode struct {
	*normal_node.NormalNode
}

func NewLiRNode(nodeId int, X, Y float64) *LiRNode {
	lirNodeType := types.NetworkNodeType_LirNode
	normalNode := normal_node.NewNormalNode(
		types.NetworkNodeType_LirNode,
		nodeId,
		fmt.Sprintf("%s-%d", lirNodeType.String(), nodeId),
	)
	normalNode.X = X
	normalNode.Y = Y
	return &LiRNode{normalNode}
}
