package nodes

import (
	"fmt"
	"zhanghefan123/security_topology/modules/entities/real_entities/normal_node"
	"zhanghefan123/security_topology/modules/entities/types"
	"zhanghefan123/security_topology/services/http/params"
)

type LiRNode struct {
	*normal_node.NormalNode
	SpecialParams params.SpecialParams
}

func NewLiRNode(nodeId int, X, Y float64, specialParams params.SpecialParams) *LiRNode {
	lirNodeType := types.NetworkNodeType_LirNode
	normalNode := normal_node.NewNormalNode(
		types.NetworkNodeType_LirNode,
		nodeId,
		fmt.Sprintf("%s-%d", lirNodeType.String(), nodeId),
	)
	normalNode.X = X
	normalNode.Y = Y
	return &LiRNode{
		NormalNode:    normalNode,
		SpecialParams: specialParams,
	}
}
